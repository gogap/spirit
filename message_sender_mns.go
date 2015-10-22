package spirit

import (
	"regexp"
	"runtime"

	"github.com/gogap/ali_mns"
	"github.com/gogap/errors"
)

type MessageSenderMNS struct {
	clientCache map[string]ali_mns.AliMNSQueue
}

func NewMessageSenderMNS() MessageSender {
	return &MessageSenderMNS{clientCache: make(map[string]ali_mns.AliMNSQueue)}

}

func (p *MessageSenderMNS) Type() string {
	return "mns"
}

func (p *MessageSenderMNS) Init() (err error) {
	if p.clientCache == nil {
		p.clientCache = make(map[string]ali_mns.AliMNSQueue)
	}
	return nil
}

func (p *MessageSenderMNS) Send(url string, message ComponentMessage) (err error) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 1024)
			runtime.Stack(buf, false)
			err = ERR_SENDER_SEND_FAILED.New(errors.Params{"type": p.Type(), "url": url, "err": string(buf)})
		}
	}()

	EventCenter.PushEvent(EVENT_BEFORE_MESSAGE_SEND, url, message)

	if url == "" {
		err = ERR_MESSAGE_ADDRESS_IS_EMPTY.New()
		return
	}

	var msgData []byte

	if msgData, err = message.Serialize(); err != nil {
		return
	}

	var client ali_mns.AliMNSQueue
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

		cli := ali_mns.NewAliMNSClient("http://"+hostId,
			accessKeyId,
			accessKeySecret)

		if cli == nil {
			err = ERR_SENDER_MNS_CLIENT_IS_NIL.New(errors.Params{"type": p.Type(), "url": url})
			return
		}

		client = ali_mns.NewMNSQueue(queueName, cli)

		p.clientCache[url] = client
	}

	msg := ali_mns.MessageSendRequest{
		MessageBody:  msgData,
		DelaySeconds: 0,
		Priority:     8}

	defer EventCenter.PushEvent(EVENT_AFTER_MESSAGE_SEND, url, message)

	if _, e := client.SendMessage(msg); e != nil {
		err = ERR_SENDER_SEND_FAILED.New(errors.Params{"type": p.Type(), "url": url, "err": e})
		return
	}

	return
}
