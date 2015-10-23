package spirit

import (
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/adjust/rmq"
	"github.com/gogap/errors"
)

type MessageReceiverRMQ struct {
	url     string
	address string
	network string
	db      int

	conns         []rmq.Connection
	queues        []rmq.Queue
	cleaner       *rmq.Cleaner
	cleanerConn   rmq.Connection
	enableCleaner bool

	queueTag  string
	queueName string

	concurrencyNumber  int
	batchMessageNumber int
	pollDuration       int

	recvLocker sync.Mutex
	isRunning  bool

	inPortName      string
	componentName   string
	onMsgReceived   OnReceiverMessageReceived
	onReceiverError OnReceiverError
}

func (p *MessageReceiverRMQ) Type() string {
	return "rmq"
}

func (p *MessageReceiverRMQ) Init(url string, options Options) (err error) {
	p.url = url

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

	p.address = address
	p.network = network

	if db, err = strconv.Atoi(strDb); err != nil {
		return
	}

	if v, e := options.GetBoolValue("disable_cleaner"); e == nil {
		p.enableCleaner = !v
	}

	p.batchMessageNumber = 16
	if v, e := options.GetInt64Value("batch_messages_number"); e == nil {
		p.batchMessageNumber = int(v)
	}

	if p.batchMessageNumber <= 0 {
		p.batchMessageNumber = 1
	}

	if v, e := options.GetInt64Value("concurrency_number"); e == nil {
		p.concurrencyNumber = int(v)
	}

	if p.concurrencyNumber <= 0 {
		p.concurrencyNumber = runtime.NumCPU()
	}

	if v, e := options.GetInt64Value("poll_duration"); e == nil {
		p.pollDuration = int(v)
	}

	if p.pollDuration < 0 {
		p.pollDuration = 0
	}

	p.queueTag = fmt.Sprintf("consumer:%s", queueName)

	for i := 0; i < p.concurrencyNumber; i++ {
		conn := rmq.OpenConnection(p.queueTag, network, address, db)
		queue := conn.OpenQueue(queueName)

		p.conns = append(p.conns, conn)
		p.queues = append(p.queues, queue)
	}

	return
}

func (p *MessageReceiverRMQ) Address() MessageAddress {
	return MessageAddress{Type: p.Type(), Url: p.url}
}

func (p *MessageReceiverRMQ) BindInPort(
	componentName, inPortName string,
	onMsgReceived OnReceiverMessageReceived,
	onReceiverError OnReceiverError) {

	p.inPortName = inPortName
	p.componentName = componentName
	p.onMsgReceived = onMsgReceived
	p.onReceiverError = onReceiverError

	return
}

func (p *MessageReceiverRMQ) Start() {
	p.recvLocker.Lock()
	defer p.recvLocker.Unlock()

	if p.isRunning {
		return
	}

	clsConn := rmq.OpenConnection("cleaner:"+p.queueTag,
		p.network,
		p.address,
		p.db)

	cleaner := rmq.NewCleaner(clsConn)

	for _, queue := range p.queues {
		queue.StartConsuming(p.batchMessageNumber, time.Millisecond*time.Duration(p.pollDuration))

		consumer := newRMQConsumer(
			p.Metadata(),
			p.onMsgReceived,
			p.onReceiverError)

		queue.AddConsumer(p.queueTag, consumer)
	}

	p.cleaner = cleaner

	if p.enableCleaner {
		go p.startCleaner(cleaner)
	}

	p.isRunning = true
}

func (p *MessageReceiverRMQ) Stop() {
	p.recvLocker.Lock()
	defer p.recvLocker.Unlock()

	if !p.isRunning {
		return
	}

	wg := sync.WaitGroup{}

	for _, queue := range p.queues {
		wg.Add(1)
		go func(q rmq.Queue) {
			defer wg.Done()
			q.StopConsuming()
		}(queue)
	}
	wg.Wait()

	p.isRunning = false
}

func (p *MessageReceiverRMQ) IsRunning() bool {
	return p.isRunning
}

func (p *MessageReceiverRMQ) Metadata() ReceiverMetadata {
	return ReceiverMetadata{
		ComponentName: p.componentName,
		PortName:      p.inPortName,
		Type:          p.Type(),
	}
}

func (p *MessageReceiverRMQ) startCleaner(cleaner *rmq.Cleaner) {
	for _ = range time.Tick(time.Second) {
		cleaner.Clean()
		if !p.isRunning {
			return
		}
	}
}

type rmqConsumer struct {
	metadata        ReceiverMetadata
	onMsgReceived   OnReceiverMessageReceived
	onReceiverError OnReceiverError
}

func newRMQConsumer(
	metadata ReceiverMetadata,
	onMsgReceived OnReceiverMessageReceived,
	onReceiverError OnReceiverError) *rmqConsumer {
	return &rmqConsumer{
		metadata:        metadata,
		onMsgReceived:   onMsgReceived,
		onReceiverError: onReceiverError,
	}
}

func (p *rmqConsumer) Consume(delivery rmq.Delivery) {
	if delivery.Ack() {
		if delivery.Payload() != "" && len(delivery.Payload()) > 0 {
			compMsg := ComponentMessage{}
			if e := compMsg.UnSerialize([]byte(delivery.Payload())); e != nil {
				e = ERR_RECEIVER_UNMARSHAL_MSG_FAILED.New(errors.Params{"type": p.metadata.Type, "err": e})
				p.onReceiverError(p.metadata.PortName, e)
			}
			p.onMsgReceived(p.metadata.PortName, delivery, compMsg, p.onMessageProcessedToDelete)
			EventCenter.PushEvent(EVENT_RECEIVER_MSG_RECEIVED, p.metadata, compMsg)
		}
	}
}

func (p *rmqConsumer) onMessageProcessedToDelete(context interface{}) {
	return
}
