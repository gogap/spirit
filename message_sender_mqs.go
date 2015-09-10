package spirit

import (
	"regexp"

	"github.com/gogap/ali_mqs"
	"github.com/gogap/errors"
)

type MessageSenderMQS struct {
	clientCache map[string]ali_mqs.AliMQSQueue
}

func NewMessageSenderMQS() MessageSender {
	return &MessageSenderMQS{clientCache: make(map[string]ali_mqs.AliMQSQueue)}

}

func (p *MessageSenderMQS) Type() string {
	return "mqs"
}

func (p *MessageSenderMQS) Init() (err error) {
	if p.clientCache == nil {
		p.clientCache = make(map[string]ali_mqs.AliMQSQueue)
	}
	return nil
}

func (p *MessageSenderMQS) Send(url string, message ComponentMessage) (err error) {

	EventCenter.PushEvent(EVENT_BEFORE_MESSAGE_SEND, url, message)

	if url == "" {
		err = ERR_MESSAGE_ADDRESS_IS_EMPTY.New()
		return
	}

	var msgData []byte

	if msgData, err = message.Serialize(); err != nil {
		return
	}

	var client ali_mqs.AliMQSQueue
	if c, exist := p.clientCache[url]; exist {
		client = c
	} else {
		hostId := ""
		accessKeyId := ""
		accessKeySecret := ""
		queueName := ""

		regUrl := regexp.MustCompile("http://(.*):(.*)@(.*)/(.*)")
		regMatched := regUrl.FindAllStringSubmatch(url, -1)

		if len(regMatched) == 1 &&
			len(regMatched[0]) == 5 {
			accessKeyId = regMatched[0][1]
			accessKeySecret = regMatched[0][2]
			hostId = regMatched[0][3]
			queueName = regMatched[0][4]
		}

		cli := ali_mqs.NewAliMQSClient("http://"+hostId,
			accessKeyId,
			accessKeySecret)

		if cli == nil {
			err = ERR_SENDER_MQS_CLIENT_IS_NIL.New(errors.Params{"type": p.Type(), "url": url})
			return
		}

		client = ali_mqs.NewMQSQueue(queueName, cli)

		p.clientCache[url] = client
	}

	msg := ali_mqs.MessageSendRequest{
		MessageBody:  msgData,
		DelaySeconds: 0,
		Priority:     8}

	defer EventCenter.PushEvent(EVENT_AFTER_MESSAGE_SEND, url, msg)

	if _, e := client.SendMessage(msg); e != nil {
		err = ERR_SENDER_SEND_FAILED.New(errors.Params{"type": p.Type(), "url": url, "err": e})
		return
	}

	return
}
