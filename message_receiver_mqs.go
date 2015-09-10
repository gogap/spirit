package spirit

import (
	"regexp"
	"sync"
	"time"

	"github.com/gogap/ali_mqs"
	"github.com/gogap/errors"
	"github.com/gogap/event_center"
)

type MessageReceiverMQS struct {
	url string

	queue ali_mqs.AliMQSQueue

	recvLocker  sync.Mutex
	isReceiving bool
	isStoped    bool
	isPaused    bool

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

func (p *MessageReceiverMQS) Address() MessageAddress {
	return MessageAddress{Type: p.Type(), Url: p.url}
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

func (p *MessageReceiverMQS) Receive(portChan *PortChan) {
	p.recvLocker.Lock()
	defer p.recvLocker.Unlock()

	receiverMetadata := ReceiverMetadata{
		ComponentName: portChan.ComponentName,
		PortName:      portChan.PortName,
		Url:           p.url,
		Type:          p.Type(),
	}

	responseChan := make(chan ali_mqs.MessageReceiveResponse, MESSAGE_CHAN_SIZE)
	errorChan := make(chan error, MESSAGE_CHAN_SIZE)

	defer close(responseChan)
	defer close(errorChan)

	stopSubscriber := event_center.NewSubscriber(func(eventName string, values ...interface{}) {
		if !p.isStoped {
			p.queue.Stop()
			p.isStoped = true

			EventCenter.PushEvent(EVENT_RECEIVER_STOPPED, receiverMetadata)
		}
	})

	EventCenter.Subscribe(EVENT_CMD_STOP, stopSubscriber)
	defer EventCenter.Unsubscribe(EVENT_CMD_STOP, stopSubscriber.Id())

	pauseSubscriber := event_center.NewSubscriber(
		func(eventName string, values ...interface{}) {
			if p.isStoped {
				return
			}

			if p.isPaused {
				return
			}

			p.queue.Stop()
			p.isPaused = true

			EventCenter.PushEvent(EVENT_RECEIVER_STOPPED, receiverMetadata)
		})

	resumeSubscriber := event_center.NewSubscriber(
		func(eventName string, values ...interface{}) {
			if p.isStoped {
				return
			}

			if !p.isPaused {
				return
			}

			go p.queue.ReceiveMessage(responseChan, errorChan)
			p.isPaused = false

			EventCenter.PushEvent(EVENT_RECEIVER_STARTED, receiverMetadata)

		})

	msgProcessedSubscriber := event_center.NewSubscriber(
		func(eventName string, values ...interface{}) {

			if values == nil || len(values) != 2 {
				return
			}

			if name := values[0].(string); name != portChan.PortName {
				return
			}

			if receiptHandle := values[1].(string); receiptHandle == "" {
				return
			} else if err := p.queue.DeleteMessage(receiptHandle); err != nil {
				EventCenter.PushEvent(EVENT_RECEIVER_MSG_DELETED, receiverMetadata, receiptHandle)
			}
		})

	EventCenter.Subscribe(EVENT_CMD_PAUSE, pauseSubscriber)
	defer EventCenter.Unsubscribe(EVENT_CMD_PAUSE, pauseSubscriber.Id())

	EventCenter.Subscribe(EVENT_CMD_RESUME, resumeSubscriber)
	defer EventCenter.Unsubscribe(EVENT_CMD_RESUME, resumeSubscriber.Id())

	EventCenter.Subscribe(EVENT_RECEIVER_MSG_PROCESSED, msgProcessedSubscriber)
	defer EventCenter.Unsubscribe(EVENT_RECEIVER_MSG_PROCESSED, msgProcessedSubscriber.Id())

	recvLoop := func(
		metadata ReceiverMetadata,
		queue ali_mqs.AliMQSQueue,
		compMsgChan chan ComponentMessage,
		compErrChan chan error) {

		EventCenter.PushEvent(EVENT_RECEIVER_STARTED, metadata)
		go queue.ReceiveMessage(responseChan, errorChan)
		p.isPaused = false

		for {
			select {
			case resp := <-responseChan:
				{
					go func(metadata ReceiverMetadata,
						queue ali_mqs.AliMQSQueue,
						compMsgChan chan ComponentMessage,
						compErrChan chan error,
						resp ali_mqs.MessageReceiveResponse) {

						defer EventCenter.PushEvent(EVENT_RECEIVER_MSG_COUNT_UPDATED, metadata, []ChanStatistics{
							{"receiver_message", len(responseChan), cap(responseChan)},
							{"receiver_error", len(errorChan), cap(errorChan)},
						})

						if resp.MessageBody != nil && len(resp.MessageBody) > 0 {
							compMsg := ComponentMessage{}
							if e := compMsg.UnSerialize(resp.MessageBody); e != nil {
								e = ERR_RECEIVER_UNMARSHAL_MSG_FAILED.New(errors.Params{"type": metadata.Type, "url": metadata.Url, "err": e})
								compErrChan <- e
							}
							compMsgChan <- compMsg
						}
					}(metadata, queue, compMsgChan, compErrChan, resp)
				}
			case respErr := <-errorChan:
				{
					go func(err error) {

						EventCenter.PushEvent(EVENT_RECEIVER_MSG_ERROR, metadata, err)

						if !ali_mqs.ERR_MQS_MESSAGE_NOT_EXIST.IsEqual(err) {
							compErrChan <- err
						}
					}(respErr)
				}
			case <-time.After(time.Second):
				{
					if len(responseChan) == 0 && len(errorChan) == 0 && p.isStoped {
						return
					}
				}
			}
		}
	}
	recvLoop(receiverMetadata, p.queue, portChan.Message, portChan.Error)
}
