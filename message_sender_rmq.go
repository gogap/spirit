package spirit

import (
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"sync"

	"github.com/adjust/rmq"
	"github.com/gogap/errors"
)

type MessageSenderRMQ struct {
	cacheLocker sync.Mutex

	clientCache map[string]map[string]rmq.Queue
	conns       map[string]rmq.Connection
}

func NewMessageSenderRMQ() MessageSender {
	return &MessageSenderRMQ{
		clientCache: make(map[string]map[string]rmq.Queue),
		conns:       make(map[string]rmq.Connection),
	}
}

func (p *MessageSenderRMQ) Type() string {
	return "rmq"
}

func (p *MessageSenderRMQ) Init() (err error) {
	if p.clientCache == nil {
		p.clientCache = make(map[string]map[string]rmq.Queue)
	}
	if p.conns == nil {
		p.conns = make(map[string]rmq.Connection)
	}
	return nil
}

func (p *MessageSenderRMQ) Send(url string, message ComponentMessage) (err error) {
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

	regUrl := regexp.MustCompile("(.*)://(.*)@(.*)/(.*)")
	regMatched := regUrl.FindAllStringSubmatch(url, -1)

	queueName := ""
	network := ""
	address := ""
	strDb := ""
	db := 0

	if len(regMatched) == 1 &&
		len(regMatched[0]) == 5 {
		network = regMatched[0][1]
		address = regMatched[0][2]
		strDb = regMatched[0][3]
		queueName = regMatched[0][4]
	}

	if db, err = strconv.Atoi(strDb); err != nil {
		return
	}

	cacheKey := fmt.Sprintf("%s://%s@%s", network, address, strDb)

	var conn rmq.Connection
	var queue rmq.Queue
	var exist bool

	p.cacheLocker.Lock()
	queueTag := fmt.Sprintf("consumer:%s", queueName)
	if conn, exist = p.conns[cacheKey]; !exist {
		conn = rmq.OpenConnection(queueTag, network, address, db)
		p.clientCache[cacheKey] = make(map[string]rmq.Queue)
		p.conns[cacheKey] = conn
	}

	if queue, exist = p.clientCache[cacheKey][queueName]; !exist {
		queue = conn.OpenQueue(queueName)
		p.clientCache[cacheKey][queueName] = queue
	}
	p.cacheLocker.Unlock()

	var msgData []byte

	if msgData, err = message.Serialize(); err != nil {
		return
	}

	queue.PublishBytes(msgData)

	defer EventCenter.PushEvent(EVENT_AFTER_MESSAGE_SEND, url, message)

	return
}
