package spirit

import (
	"fmt"
	"regexp"
	"runtime"
	"sync"
	"time"

	"github.com/gogap/ali_mns"
	"github.com/gogap/errors"
)

type MessageReceiverMNS struct {
	url string

	queue ali_mns.AliMNSQueue

	recvLocker sync.Mutex

	isRunning bool

	status ComponentStatus

	inPortName    string
	componentName string

	onMsgReceived   OnReceiverMessageReceived
	onReceiverError OnReceiverError

	batchMessageNumber int32
	qpsLimit           int32
	waitSeconds        int64
	processMode        ReceiverProcessMode
}

func NewMessageReceiverMNS(url string) MessageReceiver {
	return &MessageReceiverMNS{url: url,
		processMode:        ConcurrencyMode,
		qpsLimit:           ali_mns.DefaultQPSLimit,
		batchMessageNumber: int32(runtime.NumCPU()),
		waitSeconds:        -1,
	}
}

func (p *MessageReceiverMNS) Init(url string, options Options) (err error) {
	p.url = url
	p.waitSeconds = -1
	p.batchMessageNumber = int32(runtime.NumCPU())
	p.qpsLimit = ali_mns.DefaultQPSLimit
	p.processMode = ConcurrencyMode

	var queue ali_mns.AliMNSQueue
	if queue, err = p.newAliMNSQueue(); err != nil {
		return
	}

	if v, e := options.GetInt64Value("batch_messages_number"); e == nil {
		p.batchMessageNumber = int32(v)
	}

	if p.batchMessageNumber > ali_mns.DefaultNumOfMessages {
		p.batchMessageNumber = ali_mns.DefaultNumOfMessages
	} else if p.batchMessageNumber <= 0 {
		p.batchMessageNumber = 1
	}

	if v, e := options.GetInt64Value("qps_limit"); e == nil {
		p.qpsLimit = int32(v)
	}

	if p.qpsLimit > ali_mns.DefaultQPSLimit || p.qpsLimit < 0 {
		p.qpsLimit = ali_mns.DefaultQPSLimit
	}

	if v, e := options.GetInt64Value("wait_seconds"); e == nil {
		p.waitSeconds = v
	}

	if p.waitSeconds > 30 {
		p.waitSeconds = 30
	} else if p.waitSeconds < -1 {
		p.waitSeconds = -1
	}

	if v, e := options.GetStringValue("process_mode"); e == nil {
		if v == "" {
			p.processMode = ConcurrencyMode
		} else {
			p.processMode = ReceiverProcessMode(v)
		}
	}

	if p.processMode != ConcurrencyMode && p.processMode != SequencyMode {
		panic(fmt.Sprintf("unsupport process mode: %s", p.processMode))
	}

	p.queue = queue

	return
}

func (p *MessageReceiverMNS) Type() string {
	return "mns"
}

func (p *MessageReceiverMNS) Metadata() ReceiverMetadata {
	return ReceiverMetadata{
		ComponentName: p.componentName,
		PortName:      p.inPortName,
		Type:          p.Type(),
	}
}

func (p *MessageReceiverMNS) Address() MessageAddress {
	return MessageAddress{Type: p.Type(), Url: p.url}
}

func (p *MessageReceiverMNS) BindInPort(componentName, inPortName string, onMsgReceived OnReceiverMessageReceived, onReceiverError OnReceiverError) {
	p.inPortName = inPortName
	p.componentName = componentName
	p.onMsgReceived = onMsgReceived
	p.onReceiverError = onReceiverError
}

func (p *MessageReceiverMNS) newAliMNSQueue() (queue ali_mns.AliMNSQueue, err error) {

	hostId := ""
	accessKeyId := ""
	accessKeySecret := ""
	queueName := ""

	regUrl := regexp.MustCompile("http://(.*):(.*)@(.*)/(.*)")
	regMatched := regUrl.FindAllStringSubmatch(p.url, -1)

	if len(regMatched) == 1 &&
		len(regMatched[0]) == 5 {
		accessKeyId = regMatched[0][1]
		accessKeySecret = regMatched[0][2]
		hostId = regMatched[0][3]
		queueName = regMatched[0][4]
	}

	client := ali_mns.NewAliMNSClient("http://"+hostId,
		accessKeyId,
		accessKeySecret)

	if client == nil {
		err = ERR_RECEIVER_MNS_CLIENT_IS_NIL.New(errors.Params{"type": p.Type(), "url": p.url})
		return
	}

	queue = ali_mns.NewMNSQueue(queueName, client, p.qpsLimit)

	return
}

func (p *MessageReceiverMNS) IsRunning() bool {
	return p.isRunning
}

func (p *MessageReceiverMNS) Stop() {
	p.recvLocker.Lock()
	defer p.recvLocker.Unlock()

	if !p.isRunning {
		return
	}

	p.queue.Stop()
	p.isRunning = false
}

func (p *MessageReceiverMNS) Start() {
	p.recvLocker.Lock()
	defer p.recvLocker.Unlock()

	if p.isRunning {
		return
	}

	go func() {
		batchResponseChan := make(chan ali_mns.BatchMessageReceiveResponse, 1)
		errorChan := make(chan error, 1)
		responseChan := make(chan ali_mns.MessageReceiveResponse, p.batchMessageNumber)

		defer close(batchResponseChan)
		defer close(errorChan)
		defer close(responseChan)

		p.isRunning = true

		go p.queue.BatchReceiveMessage(batchResponseChan, errorChan, p.batchMessageNumber, p.waitSeconds)

		lastStatUpdated := time.Now()
		statUpdateFunc := func() {
			if time.Now().Sub(lastStatUpdated).Seconds() >= 1 {
				lastStatUpdated = time.Now()
				EventCenter.PushEvent(EVENT_RECEIVER_MSG_COUNT_UPDATED, p.Metadata(), []ChanStatistics{
					{"receiver_message", len(batchResponseChan), cap(batchResponseChan)},
					{"receiver_error", len(errorChan), cap(errorChan)},
				})
			}
		}

		processMessageFunc := func(resp ali_mns.MessageReceiveResponse) {
			defer statUpdateFunc()

			metadata := p.Metadata()

			if resp.MessageBody != nil && len(resp.MessageBody) > 0 {
				compMsg := ComponentMessage{}
				if e := compMsg.UnSerialize(resp.MessageBody); e != nil {
					e = ERR_RECEIVER_UNMARSHAL_MSG_FAILED.New(errors.Params{"type": metadata.Type, "err": e})
					p.onReceiverError(p.inPortName, e)
				}

				p.onMsgReceived(p.inPortName, resp.ReceiptHandle, compMsg, p.onMessageProcessedToDelete)
				EventCenter.PushEvent(EVENT_RECEIVER_MSG_RECEIVED, p.Metadata(), compMsg)
			}
		}

		workerCount := p.batchMessageNumber
		if workerCount <= 0 || p.processMode == SequencyMode {
			workerCount = 1
		}

		for i := 0; i < int(workerCount); i++ {
			go func(respChan chan ali_mns.MessageReceiveResponse, workerId int) {
				for p.isRunning {
					select {
					case resp := <-respChan:
						{
							processMessageFunc(resp)
						}
					case <-time.After(time.Second):
						{
							if len(respChan) == 0 && len(batchResponseChan) == 0 && !p.isRunning {
								return
							}
						}
					}
				}
			}(responseChan, i)
		}

		for {
			select {
			case resps := <-batchResponseChan:
				{
					for _, resp := range resps.Messages {
						responseChan <- resp
					}
				}
			case respErr := <-errorChan:
				{
					go func(err error) {
						defer statUpdateFunc()
						if !ali_mns.ERR_MNS_MESSAGE_NOT_EXIST.IsEqual(err) {
							EventCenter.PushEvent(EVENT_RECEIVER_MSG_ERROR, p.Metadata(), err)
						}
					}(respErr)
				}
			case <-time.After(time.Second):
				{
					statUpdateFunc()
					if len(batchResponseChan) == 0 && len(errorChan) == 0 && !p.isRunning {
						return
					}
				}
			}
		}
	}()

}

func (p *MessageReceiverMNS) onMessageProcessedToDelete(context interface{}) {
	if context != nil {
		if messageId, ok := context.(string); ok && messageId != "" {
			if err := p.queue.DeleteMessage(messageId); err != nil {
				EventCenter.PushEvent(EVENT_RECEIVER_MSG_DELETED, p.Metadata(), messageId)
			}
		}
	}
}
