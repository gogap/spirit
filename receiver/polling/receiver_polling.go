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

var _ spirit.Receiver = new(PollingReceiver)

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

func NewPollingReceiver(options spirit.Options) (receiver spirit.Receiver, err error) {
	conf := PollingReceiverConfig{}
	if err = options.ToObject(&conf); err != nil {
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

	spirit.Logger().WithField("actor", "receiver").
		WithField("type", "polling").
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
		if reader, err = p.readerPool.Get(); err != nil {
			spirit.Logger().WithField("actor", "receiver").
				WithField("type", "polling").
				WithField("event", "begin receive").
				Panicln(err)
		}

		for {
			if deliveries, err := p.translator.In(reader); err != nil {
				spirit.Logger().WithField("actor", "receiver").
					WithField("type", "polling").
					WithField("event", "translate reader").
					WithField("length", len(deliveries)).
					Errorln(err)

				reader.Close()

				if reader, err = p.readerPool.Get(); err != nil {
					spirit.Logger().WithField("actor", "receiver").
						WithField("type", "polling").
						WithField("event", "get a reader because of reader error").
						WithField("length", len(deliveries)).
						Errorln(err)
				}

				spirit.Logger().WithField("actor", "receiver").
					WithField("type", "polling").
					WithField("event", "receiver deliveries").
					WithField("length", len(deliveries)).
					Debugln("translator delivery from reader")
			} else {
				p.putter.Put(deliveries)
				if err = p.readerPool.Put(reader); err != nil {
					spirit.Logger().WithField("actor", "receiver").
						WithField("type", "polling").
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

	spirit.Logger().WithField("actor", "receiver").
		WithField("type", "polling").
		WithField("event", "start").
		Debugln("started")

	return
}

func (p *PollingReceiver) Stop() (err error) {
	p.statusLocker.Lock()
	defer p.statusLocker.Unlock()

	if p.status == spirit.StatusStopped {
		err = spirit.ErrReceiverDidNotRunning
		return
	}

	p.terminaled <- true

	p.readerPool.Close()

	close(p.deliveriesChan)
	close(p.terminaled)

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
