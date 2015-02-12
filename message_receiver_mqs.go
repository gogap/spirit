package spirit

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sync"

	"github.com/gogap/ali_mqs"
	"github.com/gogap/errors"
	"github.com/gogap/logs"
)

type MessageReceiverMQS struct {
	url string

	queue ali_mqs.AliMQSQueue

	recvLocker  sync.Mutex
	isReceiving bool

	responseChan chan ali_mqs.MessageReceiveResponse
	errorChan    chan error
}

func NewMessageReceiverMQS(url string) MessageReceiver {
	return &MessageReceiverMQS{url: url}
}

func (p *MessageReceiverMQS) Init(url, configFile string) (err error) {
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

func (p *MessageReceiverMQS) newAliMQSQueue() (queue ali_mqs.AliMQSQueue, err error) {
	credential := ali_mqs.NewAliMQSCredential()
	if credential == nil {
		err = ERR_RECEIVER_CREDENTIAL_IS_NIL.New(errors.Params{"type": p.Type(), "url": p.url})
		return
	}

	hostId := ""
	accessKeyId := ""
	accessKeySecret := ""
	queueName := ""

	regUrl := regexp.MustCompile("(.*):(.*)@(.*)/(.*)")
	regMatched := regUrl.FindAllStringSubmatch(p.url, -1)

	if len(regMatched) == 1 &&
		len(regMatched[0]) == 5 {
		accessKeyId = regMatched[0][1]
		accessKeySecret = regMatched[0][2]
		hostId = regMatched[0][3]
		queueName = regMatched[0][4]
	}

	client := ali_mqs.NewAliMQSClient(hostId,
		accessKeyId,
		accessKeySecret,
		credential)

	if client == nil {
		err = ERR_RECEIVER_MQS_CLIENT_IS_NIL.New(errors.Params{"type": p.Type(), "url": p.url})
		return
	}

	queue = ali_mqs.NewMQSQueue(queueName, client)

	return
}

func (p *MessageReceiverMQS) Receive(message chan ComponentMessage, err chan error) {
	p.recvLocker.Lock()
	defer p.recvLocker.Unlock()
	if p.isReceiving == true {
		panic("could not start receive twice")
	}
	p.isReceiving = true

	recvLoop := func(queue ali_mqs.AliMQSQueue, message chan ComponentMessage, err chan error) {
		responseChan := make(chan ali_mqs.MessageReceiveResponse)
		errorChan := make(chan error)

		defer close(responseChan)
		defer close(errorChan)

		go queue.ReceiveMessage(responseChan, errorChan)
		logs.Info("receiver listening - ", p.url)

		for {
			select {
			case resp := <-responseChan:
				{
					fmt.Println(resp)
					if e := queue.DeleteMessage(resp.ReceiptHandle); e != nil {
						e = ERR_RECEIVER_DELETE_MSG_ERROR.New(errors.Params{"type": p.Type(), "url": p.url, "err": e})
						logs.Error(e)
						continue
					}
					msg := resp.MessageBody
					if msg != nil && len(msg) > 0 {
						compMsg := ComponentMessage{}
						if e := json.Unmarshal(msg, &compMsg); e != nil {
							e = ERR_RECEIVER_UNMARSHAL_MSG_FAILED.New(errors.Params{"type": p.Type(), "url": p.url, "err": e})
							logs.Error(e)
							continue
						}
						message <- compMsg
					}
					return
				}
			case e := <-errorChan:
				{
					e = ERR_RECEIVER_RECV_ERROR.New(errors.Params{"type": p.Type(), "url": p.url, "err": e})
					logs.Error(e)
					err <- e
					return
				}
			}
		}
	}
	recvLoop(p.queue, message, err)
}
