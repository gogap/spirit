package polling

import (
	"io"
	"sync"
	"time"

	"github.com/gogap/spirit"
)

const (
	pollingSenderURN = "urn:spirit:sender:polling"
)

var _ spirit.WriteSender = new(PollingSender)

type _Deliveries struct {
	Deliveries []spirit.Delivery
	Error      error
}

type PollingSenderConfig struct {
	Interval int `json:"interval"`
}

type PollingSender struct {
	statusLocker sync.Mutex
	terminaled   chan bool
	conf         PollingSenderConfig

	status spirit.Status

	writerPool spirit.WriterPool

	translator spirit.OutputTranslator
	getter     spirit.DeliveryGetter
}

func init() {
	spirit.RegisterSender(pollingSenderURN, NewPollingSender)
}

func NewPollingSender(config spirit.Map) (sender spirit.Sender, err error) {
	conf := PollingSenderConfig{}
	if err = config.ToObject(&conf); err != nil {
		return
	}

	sender = &PollingSender{
		conf:       conf,
		terminaled: make(chan bool),
	}

	return
}

func (p *PollingSender) Start() (err error) {
	p.statusLocker.Lock()
	defer p.statusLocker.Unlock()

	if p.status == spirit.StatusRunning {
		err = spirit.ErrSenderAlreadyRunning
		return
	}

	spirit.Logger().WithField("actor", "polling").
		WithField("urn", pollingSenderURN).
		WithField("event", "start").
		Infoln("enter start")

	if p.writerPool == nil {
		err = spirit.ErrSenderDidNotHaveWriterPool
		return
	}

	if p.translator == nil {
		err = spirit.ErrSenderDidNotHaveTranslator
		return
	}

	if p.getter == nil {
		err = spirit.ErrSenderDeliveryGetterIsNil
		return
	}

	p.terminaled = make(chan bool)

	p.status = spirit.StatusRunning

	go func() {
		for {
			if deliveries, err := p.getter.Get(); err != nil {
				spirit.Logger().WithField("actor", spirit.ActorSender).
					WithField("urn", pollingSenderURN).
					WithField("event", "get delivery from getter").
					Errorln(err)
			} else {
				var writer io.WriteCloser
				for _, delivery := range deliveries {
					if writer, err = p.writerPool.Get(delivery); err != nil {
						spirit.Logger().WithField("actor", spirit.ActorSender).
							WithField("urn", pollingSenderURN).
							WithField("event", "get session writer").
							WithField("session_id", delivery.SessionId()).
							Errorln(err)
					} else if err = p.translator.Out(writer, delivery); err != nil {
						spirit.Logger().WithField("actor", spirit.ActorSender).
							WithField("urn", pollingSenderURN).
							WithField("event", "translate deliveries to writer").
							WithField("session_id", delivery.SessionId()).
							Errorln(err)
					} else if p.writerPool.Put(delivery, writer); err != nil {
						spirit.Logger().WithField("actor", spirit.ActorSender).
							WithField("urn", pollingSenderURN).
							WithField("event", "put writer back to pool").
							WithField("session_id", delivery.SessionId()).
							Errorln(err)
					}
				}
			}

			if p.conf.Interval > 0 {
				spirit.Logger().WithField("actor", spirit.ActorSender).
					WithField("urn", pollingSenderURN).
					WithField("event", "sleep").
					WithField("interval", p.conf.Interval).
					Debugln("sleep before get next delivery")
				time.Sleep(time.Millisecond * time.Duration(p.conf.Interval))
			}

			select {
			case signal := <-p.terminaled:
				{
					if signal == true {
						spirit.Logger().WithField("actor", spirit.ActorSender).
							WithField("urn", pollingSenderURN).
							WithField("event", "terminal").
							Debugln("terminal singal received")
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

	return
}

func (p *PollingSender) Stop() (err error) {
	p.statusLocker.Lock()
	defer p.statusLocker.Unlock()

	if p.status == spirit.StatusStopped {
		err = spirit.ErrSenderDidNotRunning
		return
	}

	spirit.Logger().WithField("actor", "polling").
		WithField("urn", pollingSenderURN).
		WithField("event", "stop").
		Infoln("enter stop")

	p.terminaled <- true

	close(p.terminaled)

	p.writerPool.Close()

	spirit.Logger().WithField("actor", "polling").
		WithField("urn", pollingSenderURN).
		WithField("event", "stop").
		Infoln("stopped")

	return
}

func (p *PollingSender) Status() spirit.Status {
	return p.status
}

func (p *PollingSender) SetWriterPool(pool spirit.WriterPool) (err error) {
	p.writerPool = pool
	return
}

func (p *PollingSender) SetTranslator(translator spirit.OutputTranslator) (err error) {
	p.translator = translator
	return
}

func (p *PollingSender) SetDeliveryGetter(getter spirit.DeliveryGetter) (err error) {
	p.getter = getter
	return
}
