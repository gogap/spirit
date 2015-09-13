package spirit

import (
	"regexp"
	"sync"
	"time"

	"github.com/gogap/ali_mqs"
	"github.com/gogap/errors"
)

type MessageReceiverMQS struct {
	url string

	queue ali_mqs.AliMQSQueue

	recvLocker sync.Mutex

	isRunning bool

	status ComponentStatus

	inPortName    string
	componentName string

	onMsgReceived   OnReceiverMessageReceived
	onReceiverError OnReceiverError

	responseChan chan ali_mqs.MessageReceiveResponse
	errorChan    chan error
}

func NewMessageReceiverMQS(url string) MessageReceiver {
	return &MessageReceiverMQS{url: url}
}

func (p *MessageReceiverMQS) Init(url string, options Options) (err error) {
	p.url = url

	var queue ali_mqs.AliMQSQueue
	if queue, err = p.newAliMQSQueue(); err != nil {
		return
	}

	p.queue = queue

	return
}

func (p *MessageReceiverMQS) Type() string {
	return "mqs"
}

func (p *MessageReceiverMQS) Metadata() ReceiverMetadata {
	return ReceiverMetadata{
		ComponentName: p.componentName,
		PortName:      p.inPortName,
		Type:          p.Type(),
	}
}

func (p *MessageReceiverMQS) Address() MessageAddress {
	return MessageAddress{Type: p.Type(), Url: p.url}
}

func (p *MessageReceiverMQS) BindInPort(componentName, inPortName string, onMsgReceived OnReceiverMessageReceived, onReceiverError OnReceiverError) {
	p.inPortName = inPortName
	p.componentName = componentName
	p.onMsgReceived = onMsgReceived
	p.onReceiverError = onReceiverError
}

func (p *MessageReceiverMQS) newAliMQSQueue() (queue ali_mqs.AliMQSQueue, err error) {

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

	client := ali_mqs.NewAliMQSClient("http://"+hostId,
		accessKeyId,
		accessKeySecret)

	if client == nil {
		err = ERR_RECEIVER_MQS_CLIENT_IS_NIL.New(errors.Params{"type": p.Type(), "url": p.url})
		return
	}

	queue = ali_mqs.NewMQSQueue(queueName, client)

	return
}

func (p *MessageReceiverMQS) IsRunning() bool {
	return p.isRunning
}

func (p *MessageReceiverMQS) Stop() {
	p.recvLocker.Lock()
	defer p.recvLocker.Unlock()

	if !p.isRunning {
		return
	}

	p.queue.Stop()
	p.isRunning = false
}

func (p *MessageReceiverMQS) Start() {
	p.recvLocker.Lock()
	defer p.recvLocker.Unlock()

	if p.isRunning {
		return
	}

	go func() {
		responseChan := make(chan ali_mqs.MessageReceiveResponse, MESSAGE_CHAN_SIZE)
		errorChan := make(chan error, MESSAGE_CHAN_SIZE)

		defer close(responseChan)
		defer close(errorChan)

		p.isRunning = true

		go p.queue.ReceiveMessage(responseChan, errorChan)

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

		for {
			select {
			case resp := <-responseChan:
				{
					go func(resp ali_mqs.MessageReceiveResponse) {
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
					}(resp)
				}
			case respErr := <-errorChan:
				{
					go func(err error) {
						defer statUpdateFunc()
						if !ali_mqs.ERR_MQS_MESSAGE_NOT_EXIST.IsEqual(err) {
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

func (p *MessageReceiverMQS) onMessageProcessedToDelete(messageId string) {
	if err := p.queue.DeleteMessage(messageId); err != nil {
		EventCenter.PushEvent(EVENT_RECEIVER_MSG_DELETED, p.Metadata(), messageId)
	}
}
