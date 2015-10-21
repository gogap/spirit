package spirit

import (
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
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
	consumers     []*rmqConsumer
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
		p.concurrencyNumber = p.batchMessageNumber * int(runtime.NumCPU())
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

	for _, queue := range p.queues {
		queue.StartConsuming(p.batchMessageNumber, time.Millisecond*time.Duration(p.pollDuration))

		consumer := newRMQConsumer(queue, p.Metadata(), p.onMsgReceived, p.onReceiverError)
		queue.AddConsumer(p.queueTag, consumer)
		p.consumers = append(p.consumers, consumer)

		consumer.start()
	}

	if p.enableCleaner {
		p.startCleaner()
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

	for _, consumer := range p.consumers {
		wg.Add(1)
		go func(q rmq.Consumer) {
			defer wg.Done()
			consumer.stop()
		}(consumer)

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

func (p *MessageReceiverRMQ) startCleaner() {
	go func() {
		clsConn := rmq.OpenConnection("cleaner:"+p.queueTag,
			p.network,
			p.address,
			p.db)

		cleaner := rmq.NewCleaner(clsConn)

		for _ = range time.Tick(time.Second) {
			cleaner.Clean()
			if !p.isRunning {
				return
			}
		}
	}()
}

type rmqConsumer struct {
	name string

	queue           rmq.Queue
	metadata        ReceiverMetadata
	onMsgReceived   OnReceiverMessageReceived
	onReceiverError OnReceiverError

	isRunning bool

	statusLock   sync.Mutex
	msgOnProcess int64
}

func newRMQConsumer(
	queue rmq.Queue,
	metadata ReceiverMetadata,
	onMsgReceived OnReceiverMessageReceived,
	onReceiverError OnReceiverError) *rmqConsumer {
	return &rmqConsumer{
		queue:           queue,
		metadata:        metadata,
		onMsgReceived:   onMsgReceived,
		onReceiverError: onReceiverError,
	}
}

func (p *rmqConsumer) start() {
	p.statusLock.Lock()
	defer p.statusLock.Unlock()

	if p.isRunning {
		return
	}

	atomic.StoreInt64(&p.msgOnProcess, 0)
	p.isRunning = true
}

func (p *rmqConsumer) stop() {
	p.statusLock.Lock()
	defer p.statusLock.Unlock()

	if !p.isRunning {
		return
	}

	p.isRunning = false
}

func (p *rmqConsumer) Consume(delivery rmq.Delivery) {
	if !p.isRunning {
		time.Sleep(time.Hour)
		return
	}

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
