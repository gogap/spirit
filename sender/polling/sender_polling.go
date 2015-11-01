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

var _ spirit.Sender = new(PollingSender)

type _Deliveries struct {
	Deliveries []spirit.Delivery
	Error      error
}

type PollingSenderConfig struct {
	Interval       int  `json:"interval"`
	DisableSession bool `json:"disable_session"`
}

type PollingSender struct {
	statusLocker sync.Mutex
	terminaled   chan bool
	conf         PollingSenderConfig

	status spirit.Status

	newWriterFunc spirit.NewWriterFunc
	writerOptions spirit.Options

	translator spirit.OutputTranslator
	getter     spirit.DeliveryGetter

	globalWriter   io.WriteCloser
	sessionWriters map[string]io.WriteCloser
	WriterLocker   sync.Mutex
}

func init() {
	spirit.RegisterSender(pollingSenderURN, NewPollingSender)
}

func NewPollingSender(options spirit.Options) (receiver spirit.Sender, err error) {
	conf := PollingSenderConfig{}
	if err = options.ToObject(&conf); err != nil {
		return
	}

	receiver = &PollingSender{
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
		WithField("type", "classic").
		WithField("event", "start").
		Infoln("enter start")

	if p.newWriterFunc == nil {
		err = spirit.ErrSenderCanNotCreaterWriter
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
				spirit.Logger().WithField("actor", "sender").
					WithField("type", "polling").
					WithField("event", "get delivery from getter").
					Errorln(err)
			} else {
				var writer io.WriteCloser
				for _, delivery := range deliveries {
					if writer, err = p.getWriter(delivery.SessionId()); err != nil {
						spirit.Logger().WithField("actor", "sender").
							WithField("type", "polling").
							WithField("event", "get session writer").
							WithField("session_id", delivery.SessionId()).
							Errorln(err)
					} else if err = p.translator.Out(writer, delivery); err != nil {
						p.delWriter(delivery.SessionId())
						spirit.Logger().WithField("actor", "sender").
							WithField("type", "polling").
							WithField("event", "translate deliveries to writer").
							WithField("session_id", delivery.SessionId()).
							Errorln(err)
					}
				}
			}

			if p.conf.Interval > 0 {
				spirit.Logger().WithField("actor", "sender").
					WithField("type", "polling").
					WithField("event", "sleep").
					WithField("interval", p.conf.Interval).
					Debugln("sleep before get next delivery")
				time.Sleep(time.Millisecond * time.Duration(p.conf.Interval))
			}

			if len(p.terminaled) > 0 {
				select {
				case signal := <-p.terminaled:
					{
						if signal == true {
							spirit.Logger().WithField("actor", "sender").
								WithField("type", "polling").
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
		}

	}()

	return
}

func (p *PollingSender) getWriter(sessionId string) (writer io.WriteCloser, err error) {
	p.WriterLocker.Lock()
	defer p.WriterLocker.Unlock()

	if p.conf.DisableSession {
		if p.globalWriter == nil {
			writer, err = p.newWriterFunc(p.writerOptions)
			p.globalWriter = writer
		} else {
			writer = p.globalWriter
		}
		return
	}

	if w, exist := p.sessionWriters[sessionId]; exist {
		writer = w
		return
	} else {
		writer, err = p.newWriterFunc(p.writerOptions)
		return
	}

	return
}

func (p *PollingSender) delWriter(sessionId string) {
	p.WriterLocker.Lock()
	defer p.WriterLocker.Unlock()

	if p.conf.DisableSession {
		p.globalWriter = nil
		return
	}

	if _, exist := p.sessionWriters[sessionId]; exist {
		delete(p.sessionWriters, sessionId)
	}

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
		WithField("type", "classic").
		WithField("event", "stop").
		Infoln("enter stop")

	p.terminaled <- true

	close(p.terminaled)

	spirit.Logger().WithField("actor", "polling").
		WithField("type", "classic").
		WithField("event", "stop").
		Infoln("stopped")

	return
}

func (p *PollingSender) Status() spirit.Status {
	return p.status
}

func (p *PollingSender) SetNewWriterFunc(newFunc spirit.NewWriterFunc, options spirit.Options) (err error) {
	p.newWriterFunc = newFunc
	p.writerOptions = options
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
