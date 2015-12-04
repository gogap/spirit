package classic

import (
	"github.com/gogap/errors"
	"github.com/gogap/spirit"
	"sync"
	"time"
)

const (
	classicMessengerURN = "urn:spirit:messenger:classic"
)

var _ spirit.Messenger = new(ClassicMessenger)

type ClassicMessengerConfig struct {
}

type ClassicMessenger struct {
	name string
	conf ClassicMessengerConfig

	router spirit.Router

	status       spirit.Status
	statusLocker sync.Mutex
	terminaled   chan bool
}

func init() {
	spirit.RegisterMessenger(classicMessengerURN, NewClassicMessenger)
}

func NewClassicMessenger(name string, options spirit.Map) (messenger spirit.Messenger, err error) {
	conf := ClassicMessengerConfig{}

	if err = options.ToObject(&conf); err != nil {
		return
	}

	messenger = &ClassicMessenger{
		name:       name,
		conf:       conf,
		terminaled: make(chan bool),
	}

	return
}

func (p *ClassicMessenger) URN() string {
	return classicMessengerURN
}

func (p *ClassicMessenger) Name() string {
	return p.name
}

func (p *ClassicMessenger) SetRouter(router spirit.Router) (err error) {
	p.router = router
	return
}

func (p *ClassicMessenger) Start() (err error) {
	p.statusLocker.Lock()
	defer p.statusLocker.Unlock()

	spirit.Logger().
		WithField("actor", spirit.ActorMessenger).
		WithField("urn", p.URN()).
		WithField("name", p.Name()).
		WithField("event", "start").
		Debugln("enter start")

	if p.status == spirit.StatusRunning {
		err = spirit.ErrMessengerAlreadyRunning
		return
	}

	if p.router == nil {
		err = spirit.ErrMessengerDidNotHaveRouter
		return
	}

	p.terminaled = make(chan bool)

	p.status = spirit.StatusRunning

	go p.messageLoop()

	spirit.Logger().
		WithField("actor", spirit.ActorMessenger).
		WithField("urn", p.URN()).
		WithField("name", p.Name()).
		WithField("event", "start").
		Infoln("started")

	return
}

func (p *ClassicMessenger) Stop() (err error) {
	p.statusLocker.Lock()
	defer p.statusLocker.Unlock()

	spirit.Logger().
		WithField("actor", spirit.ActorMessenger).
		WithField("urn", p.URN()).
		WithField("name", p.Name()).
		WithField("event", "stop").
		Debugln("enter stop")

	if p.status == spirit.StatusStopped {
		err = spirit.ErrMessengerDidNotRunning
		return
	}

	p.terminaled <- true

	close(p.terminaled)

	spirit.Logger().
		WithField("actor", spirit.ActorRouter).
		WithField("urn", p.URN()).
		WithField("name", p.Name()).
		WithField("event", "stop").
		Infoln("stopped")

	return
}

func (p *ClassicMessenger) Status() spirit.Status {
	return p.status
}

func (p *ClassicMessenger) messageLoop() {
	for {
		for _, inbox := range p.router.Inboxes() {
			var deliveries []spirit.Delivery
			var err error
			if deliveries, err = inbox.Get(); err != nil {
				spirit.Logger().
					WithField("actor", spirit.ActorMessenger).
					WithField("urn", p.URN()).
					WithField("name", p.Name()).
					WithField("event", "router loop").
					Errorln(err)
			}

		next_delivery:
			for _, delivery := range deliveries {
				var routeItems []spirit.RouteItem
				if routeItems, err = p.router.RouteToHandlers(delivery); err != nil {
					spirit.Logger().
						WithField("actor", spirit.ActorMessenger).
						WithField("urn", p.URN()).
						WithField("name", p.Name()).
						WithField("event", "route item").
						Errorln(err)
				} else {
					if len(routeItems) == 0 {
						spirit.Logger().
							WithField("actor", spirit.ActorMessenger).
							WithField("urn", p.URN()).
							WithField("name", p.Name()).
							WithField("event", "route item").
							Debugln("no route item found")
					}

					for _, item := range routeItems {
						if ret, e := item.Handler(delivery.Payload()); e != nil {

							switch retErr := e.(type) {
							case *spirit.Error:
								{
									delivery.Payload().AppendError(retErr)
								}
							case spirit.Error:
								{
									delivery.Payload().AppendError(&retErr)
								}
							case errors.ErrCode:
								{
									e := &spirit.Error{
										Code:       retErr.Code(),
										Id:         retErr.Id(),
										Namespace:  retErr.Namespace(),
										Message:    retErr.Error(),
										StackTrace: retErr.StackTrace(),
										Context:    map[string]interface{}(retErr.Context()),
									}

									delivery.Payload().AppendError(e)
								}
							default:
								errCode := spirit.ErrComponentHandlerError.New().Append(e)
								errRet := &spirit.Error{
									Code:       errCode.Code(),
									Id:         errCode.Id(),
									Namespace:  errCode.Namespace(),
									Message:    errCode.Error(),
									StackTrace: errCode.StackTrace(),
									Context:    map[string]interface{}(errCode.Context()),
								}

								delivery.Payload().AppendError(errRet)
							}

							spirit.Logger().
								WithField("actor", spirit.ActorMessenger).
								WithField("urn", p.URN()).
								WithField("name", p.Name()).
								WithField("event", "route item execute error").
								Debugln(e)

							if !item.ContinueOnError {
								break
							}
						} else {
							if e := delivery.Payload().SetData(ret); e != nil {
								spirit.Logger().
									WithField("actor", spirit.ActorMessenger).
									WithField("urn", p.URN()).
									WithField("name", p.Name()).
									WithField("event", "set data to delviery").
									Errorln(e)

								break next_delivery
							}
						}
					}

					if outboxes, e := p.router.RouteToOutboxes(delivery); e != nil {
						spirit.Logger().
							WithField("actor", spirit.ActorMessenger).
							WithField("urn", p.URN()).
							WithField("name", p.Name()).
							WithField("event", "router to outboxes").
							Errorln(e)
					} else {
						if len(outboxes) == 0 {
							spirit.Logger().
								WithField("actor", spirit.ActorMessenger).
								WithField("urn", p.URN()).
								WithField("name", p.Name()).
								WithField("event", "router to outboxes").
								Errorln("no outbox found")
							break next_delivery
						}

						for _, outbox := range outboxes {

							spirit.Logger().
								WithField("actor", spirit.ActorMessenger).
								WithField("urn", p.URN()).
								WithField("name", p.Name()).
								WithField("event", "router to outboxes").
								WithField("outbox labels", outbox.Labels()).
								Debugln("outbox found")

							if e := outbox.Put([]spirit.Delivery{delivery}); e != nil {
								spirit.Logger().
									WithField("actor", spirit.ActorMessenger).
									WithField("urn", p.URN()).
									WithField("name", p.Name()).
									WithField("event", "put delivery to outbox").
									Errorln(e)
							}
						}
					}
				}
			}
		}

		if p.checkTerminaled() {
			return
		}
	}
}

func (p *ClassicMessenger) checkTerminaled() bool {
	select {
	case signal, more := <-p.terminaled:
		{
			if signal == true || more == false {
				return true
			}
		}
	case <-time.After(time.Microsecond * 10):
		{
		}
	}

	return false
}
