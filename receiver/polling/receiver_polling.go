package polling

import (
	"io"
	"sync"
	"time"

	"github.com/gogap/spirit"
)

const (
	pollingReceiverURN = "urn:spirit:receiver:polling"
)

var _ spirit.ReadReceiver = new(PollingReceiver)

type _Deliveries struct {
	Deliveries []spirit.Delivery
	Error      error
}

type PollingReceiverConfig struct {
	Interval   int `json:"interval"`
	BufferSize int `json:"buffer_size"`
	Timeout    int `json:"timeout"`
}

type PollingReceiver struct {
	statusLocker sync.Mutex
	terminaled   chan bool
	conf         PollingReceiverConfig

	status spirit.Status

	readerPool spirit.ReaderPool

	deliveriesChan chan _Deliveries

	translator spirit.InputTranslator
	putter     spirit.DeliveryPutter
}

func init() {
	spirit.RegisterReceiver(pollingReceiverURN, NewPollingReceiver)
}

func NewPollingReceiver(config spirit.Map) (receiver spirit.Receiver, err error) {
	conf := PollingReceiverConfig{}
	if err = config.ToObject(&conf); err != nil {
		return
	}

	receiver = &PollingReceiver{
		conf:           conf,
		terminaled:     make(chan bool),
		deliveriesChan: make(chan _Deliveries, conf.BufferSize),
	}

	return
}

func (p *PollingReceiver) Start() (err error) {
	p.statusLocker.Lock()
	defer p.statusLocker.Unlock()

	spirit.Logger().WithField("actor", spirit.ActorReceiver).
		WithField("urn", pollingReceiverURN).
		WithField("event", "start").
		Debugln("enter start")

	if p.status == spirit.StatusRunning {
		err = spirit.ErrReceiverAlreadyRunning
		return
	}

	if p.readerPool == nil {
		err = spirit.ErrReceiverDidNotHaveReaderPool
		return
	}

	if p.translator == nil {
		err = spirit.ErrReceiverDidNotHaveTranslator
		return
	}

	if p.putter == nil {
		err = spirit.ErrReceiverDeliveryPutterIsNil
		return
	}

	p.terminaled = make(chan bool)

	p.status = spirit.StatusRunning

	go func() {
		var reader io.ReadCloser
		var err error

		for {
			if reader, err = p.readerPool.Get(); err != nil {
				spirit.Logger().WithField("actor", spirit.ActorReceiver).
					WithField("urn", pollingReceiverURN).
					WithField("event", "get reader from reader pool").
					Errorln(err)
			}

			if deliveries, err := p.translator.In(reader); err != nil {
				spirit.Logger().WithField("actor", spirit.ActorReceiver).
					WithField("urn", pollingReceiverURN).
					WithField("event", "translate reader").
					WithField("length", len(deliveries)).
					Errorln(err)

				reader.Close()

				spirit.Logger().WithField("actor", spirit.ActorReceiver).
					WithField("urn", pollingReceiverURN).
					WithField("event", "receiver deliveries").
					WithField("length", len(deliveries)).
					Debugln("translator delivery from reader")

			} else {
				if err = p.putter.Put(deliveries); err != nil {
					spirit.Logger().WithField("actor", spirit.ActorReceiver).
						WithField("urn", pollingReceiverURN).
						WithField("event", "put deliveries").
						Errorln(err)
				}

				if err = p.readerPool.Put(reader); err != nil {
					spirit.Logger().WithField("actor", spirit.ActorReceiver).
						WithField("urn", pollingReceiverURN).
						WithField("event", "put reader back to pool").
						Errorln(err)
				}
			}

			if p.conf.Interval > 0 {
				time.Sleep(time.Millisecond * time.Duration(p.conf.Interval))
			}

			select {
			case signal := <-p.terminaled:
				{
					if signal == true {
						return
					}
				}
			case <-time.After(time.Microsecond * 10):
				{
					continue
				}
			}
		}

	}()

	spirit.Logger().WithField("actor", spirit.ActorReceiver).
		WithField("urn", pollingReceiverURN).
		WithField("event", "start").
		Infoln("started")

	return
}

func (p *PollingReceiver) Stop() (err error) {
	p.statusLocker.Lock()
	defer p.statusLocker.Unlock()

	spirit.Logger().WithField("actor", spirit.ActorReceiver).
		WithField("urn", pollingReceiverURN).
		WithField("event", "stop").
		Debugln("enter stop")

	if p.status == spirit.StatusStopped {
		err = spirit.ErrReceiverDidNotRunning
		return
	}

	p.terminaled <- true

	p.readerPool.Close()

	close(p.deliveriesChan)
	close(p.terminaled)

	spirit.Logger().WithField("actor", spirit.ActorRouter).
		WithField("urn", pollingReceiverURN).
		WithField("event", "stop").
		Infoln("stopped")

	return
}

func (p *PollingReceiver) Status() spirit.Status {
	return p.status
}

func (p *PollingReceiver) SetReaderPool(pool spirit.ReaderPool) (err error) {
	p.readerPool = pool
	return
}

func (p *PollingReceiver) SetTranslator(translator spirit.InputTranslator) (err error) {
	p.translator = translator
	return
}

func (p *PollingReceiver) SetDeliveryPutter(putter spirit.DeliveryPutter) (err error) {
	p.putter = putter
	return
}
