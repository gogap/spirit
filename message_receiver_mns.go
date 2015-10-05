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

	responseChan chan ali_mns.MessageReceiveResponse
	errorChan    chan error

	batchMessageNumber int32
	qpsLimit           int32
	waitSeconds        int64
	processMode        string
}

func NewMessageReceiverMNS(url string) MessageReceiver {
	return &MessageReceiverMNS{url: url,
		processMode:        "concurrency",
		qpsLimit:           ali_mns.DefaultQPSLimit,
		batchMessageNumber: int32(runtime.NumCPU()),
		waitSeconds:        -1,
	}
}

func (p *MessageReceiverMNS) Init(url string, options Options) (err error) {
	p.url = url

	var queue ali_mns.AliMNSQueue
	if queue, err = p.newAliMNSQueue(); err != nil {
		return
	}

	if v, e := options.GetInt64Value("batch_messages_number"); e == nil {
		p.batchMessageNumber = int32(v)
	}

	if v, e := options.GetInt64Value("qps_limit"); e == nil {
		p.qpsLimit = int32(v)
	}

	if v, e := options.GetInt64Value("wait_seconds"); e == nil {
		if v > 30 {
			p.waitSeconds = 30
		} else {
			p.waitSeconds = v
		}
	}

	if v, e := options.GetStringValue("process_mode"); e == nil {
		p.processMode = v
	} else if v == "" {
		p.processMode = "concurrency"
	}

	if p.processMode != "concurrency" && p.processMode != "sequency" {
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
		responseChan := make(chan ali_mns.BatchMessageReceiveResponse, 1)
		errorChan := make(chan error, MESSAGE_CHAN_SIZE)

		defer close(responseChan)
		defer close(errorChan)

		p.isRunning = true

		go p.queue.BatchReceiveMessage(responseChan, errorChan, p.batchMessageNumber, p.waitSeconds)

		lastStatUpdated := time.Now()
		statUpdateFunc := func() {
			if time.Now().Sub(lastStatUpdated).Seconds() >= 1 {
				lastStatUpdated = time.Now()
				EventCenter.PushEvent(EVENT_RECEIVER_MSG_COUNT_UPDATED, p.Metadata(), []ChanStatistics{
					{"receiver_message", len(responseChan), cap(responseChan)},
					{"receiver_error", len(errorChan), cap(errorChan)},
				})
			}
		}

		handlerFunc := func(resp ali_mns.MessageReceiveResponse) {
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

		for {
			select {
			case resps := <-responseChan:
				{
					for _, resp := range resps.Messages {
						if p.processMode == "concurrency" {
							go handlerFunc(resp)
						} else {
							handlerFunc(resp)
						}
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
					if len(responseChan) == 0 && len(errorChan) == 0 && !p.isRunning {
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
