package spirit

import (
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gogap/logrus_mate"
)

var (
	logrusMate *logrus_mate.LogrusMate
	logger     = logrus_mate.Logger()
)

type ActorType string

var (
	ActorReader           ActorType = "reader"
	ActorWriter           ActorType = "writer"
	ActorInputTranslator  ActorType = "input_translator"
	ActorOutputTranslator ActorType = "output_translator"
	ActorReceiver         ActorType = "receiver"
	ActorSender           ActorType = "sender"
	ActorLabelMatcher     ActorType = "label_matcher"
	ActorRouter           ActorType = "router"
	ActorComponent        ActorType = "component"
	ActorInbox            ActorType = "inbox"
	ActorOutbox           ActorType = "outbox"
	ActorReaderPool       ActorType = "reader_pool"
	ActorWriterPool       ActorType = "writer_pool"
	ActorURNRewriter      ActorType = "urn_rewriter"
	ActorConsole          ActorType = "console"
)

func init() {
	initLogLevel := os.Getenv(SpiritInitialLogLevelEnvKey)
	if initLogLevel == "" {
		logger.Level = logrus.InfoLevel
		return
	}
	if lvl, err := logrus.ParseLevel(initLogLevel); err != nil {
		panic(err)
	} else {
		logger.Level = lvl
	}
}

func Logger(loggerName ...string) *logrus.Logger {
	if loggerName == nil ||
		len(loggerName) == 0 ||
		loggerName[0] == "" {
		return logger
	}
	return logrusMate.Logger(loggerName...)
}

type Starter interface {
	Start() (err error)
}

type Stopper interface {
	Stop() (err error)
}

type StartStopper interface {
	Starter
	Stopper
	Status() Status
}

type Spirit interface {
	Build(conf SpiritConfig) (err error)
	Run() (wg *sync.WaitGroup, err error)
}

type ClassicSpirit struct {
	statusLocker sync.Mutex

	built bool

	status Status

	readerPools       map[string]ReaderPool
	writerPools       map[string]WriterPool
	inputTranslators  map[string]InputTranslator
	outputTranslators map[string]OutputTranslator
	inboxes           map[string]Inbox
	outboxes          map[string]Outbox
	receivers         map[string]Receiver
	senders           map[string]Sender
	labelMatchers     map[string]LabelMatcher
	urnRewriters      map[string]URNRewriter
	routers           map[string]Router
	consoles          map[string]Console
	components        map[string][]Component
}

func NewClassicSpirit() (s Spirit, err error) {
	return &ClassicSpirit{
		readerPools:       make(map[string]ReaderPool),
		writerPools:       make(map[string]WriterPool),
		inputTranslators:  make(map[string]InputTranslator),
		outputTranslators: make(map[string]OutputTranslator),
		inboxes:           make(map[string]Inbox),
		outboxes:          make(map[string]Outbox),
		receivers:         make(map[string]Receiver),
		senders:           make(map[string]Sender),
		labelMatchers:     make(map[string]LabelMatcher),
		urnRewriters:      make(map[string]URNRewriter),
		routers:           make(map[string]Router),
		consoles:          make(map[string]Console),
		components:        make(map[string][]Component),
	}, nil
}

func (p *ClassicSpirit) Run() (wg *sync.WaitGroup, err error) {
	p.statusLocker.Lock()
	defer p.statusLocker.Unlock()

	if p.status != StatusStopped {
		err = ErrSpiritAlreadyRunning
		return
	}

	if !p.built {
		err = ErrSpiritNotBuild
		return
	}

	wg = new(sync.WaitGroup)

	p.status = StatusRunning

	for name, components := range p.components {
		for _, actor := range components {
			p.startActor(name, ActorComponent, actor)
		}
	}

	for name, actor := range p.outboxes {
		p.startActor(name, ActorOutbox, actor)
	}

	for name, actor := range p.inboxes {
		p.startActor(name, ActorInbox, actor)
	}

	for name, actor := range p.senders {
		p.startActor(name, ActorSender, actor)
	}

	for name, actor := range p.receivers {
		p.startActor(name, ActorReceiver, actor)
	}

	for name, actor := range p.routers {
		p.startActor(name, ActorRouter, actor)
	}

	for _, router := range p.routers {
		go p.loop(router)
	}

	for name, actor := range p.consoles {
		p.startActor(name, ActorConsole, actor)
	}

	wg.Add(1)
	p.waitSignal(wg)

	return
}

func (p *ClassicSpirit) loop(router Router) {
	for {
		for _, inbox := range router.Inboxes() {
			var deliveries []Delivery
			var err error
			if deliveries, err = inbox.Get(); err != nil {
				logger.WithField("module", "spirit").
					WithField("event", "router loop").
					Errorln(err)
			}

			for _, delivery := range deliveries {
				var handlers []ComponentHandler
				if handlers, err = router.RouteToHandlers(delivery); err != nil {
					logger.WithField("module", "spirit").
						WithField("event", "router to handlers").
						Errorln(err)
				} else {
					if len(handlers) == 0 {
						logger.WithField("module", "spirit").
							WithField("event", "router to handlers").
							Debugln("no handler found")
					}

					for _, handler := range handlers {
						if ret, e := handler(delivery.Payload()); e != nil {
							delivery.Payload().SetError(e)

							logger.WithField("module", "spirit").
								WithField("event", "set payload error").
								Debugln(e)

						} else {
							if e := delivery.Payload().SetData(ret); e != nil {
								logger.WithField("module", "spirit").
									WithField("event", "set payload data").
									Errorln(e)
							}
						}
					}

					if outboxes, e := router.RouteToOutboxes(delivery); e != nil {
						logger.WithField("module", "spirit").
							WithField("event", "router to outboxes").
							Errorln(e)
					} else {
						if len(outboxes) == 0 {
							logger.WithField("module", "spirit").
								WithField("event", "router to outboxes").
								Errorln("no outbox found")
							break
						}

						for _, outbox := range outboxes {
							if e := outbox.Put([]Delivery{delivery}); e != nil {
								logger.WithField("module", "spirit").
									WithField("event", "put delivery to outbox").
									Errorln(e)
							}
						}
					}
				}
			}
		}
		if p.status == StatusStopped {
			return
		}
	}
}

func (p *ClassicSpirit) stopActor(name string, actorType ActorType, actors ...interface{}) (err error) {
	for _, actor := range actors {
		switch stopper := actor.(type) {
		case Stopper:
			if err = stopper.Stop(); err != nil {
				logger.WithField("module", "spirit").
					WithField("event", "stop "+actorType).
					WithField("name", name).
					Panicln(err)
			}
		default:
			logger.WithField("module", "spirit").
				WithField("event", "stop "+actorType).
				WithField("name", name).
				WithField("type", reflect.TypeOf(actor)).
				Debugln("non-stopper")
		}
	}
	return
}

func (p *ClassicSpirit) startActor(name string, actorType ActorType, actors ...interface{}) (err error) {
	for _, actor := range actors {
		switch starter := actor.(type) {
		case Starter:
			if err = starter.Start(); err != nil {
				logger.WithField("module", "spirit").
					WithField("event", "start "+actorType).
					WithField("name", name).
					Panicln(err)
			}
		default:
			logger.WithField("module", "spirit").
				WithField("event", "start "+actorType).
				WithField("name", name).
				WithField("type", reflect.TypeOf(actor)).
				Debugln("non-starter")
		}
	}
	return
}

func (p *ClassicSpirit) stop() (err error) {
	if p.status == StatusStopped {
		err = ErrSpiritAlreadyStopped
		return
	}

	for name, actor := range p.receivers {
		if err = p.stopActor(name, ActorReceiver, actor); err != nil {
			logger.WithField("module", "spirit").
				WithField("event", "stop receiver").
				WithField("actor_name", name).
				Errorln(err)
		} else {
			logger.WithField("module", "spirit").
				WithField("event", "stop receiver").
				WithField("actor_name", name).
				Debugln("receiver stopped")
		}
	}

	for name, actor := range p.senders {
		if err = p.stopActor(name, ActorSender, actor); err != nil {
			logger.WithField("module", "spirit").
				WithField("event", "stop sender").
				WithField("actor_name", name).
				Errorln(err)
		} else {
			logger.WithField("module", "spirit").
				WithField("event", "stop sender").
				WithField("actor_name", name).
				Debugln("sender stopped")
		}
	}

	for name, actor := range p.inboxes {
		if err = p.stopActor(name, ActorInbox, actor); err != nil {
			logger.WithField("module", "spirit").
				WithField("event", "stop inbox").
				WithField("actor_name", name).
				Errorln(err)
		} else {
			logger.WithField("module", "spirit").
				WithField("event", "stop inbox").
				WithField("actor_name", name).
				Debugln("inbox stopped")
		}
	}

	for name, actor := range p.outboxes {
		if err = p.stopActor(name, ActorOutbox, actor); err != nil {
			logger.WithField("module", "spirit").
				WithField("event", "stop outbox").
				WithField("actor_name", name).
				Errorln(err)
		} else {
			logger.WithField("module", "spirit").
				WithField("event", "stop outbox").
				WithField("actor_name", name).
				Debugln("outbox stopped")
		}
	}

	for name, actor := range p.routers {
		if err = p.stopActor(name, ActorRouter, actor); err != nil {
			logger.WithField("module", "spirit").
				WithField("event", "stop router").
				WithField("actor_name", name).
				Errorln(err)
		} else {
			logger.WithField("module", "spirit").
				WithField("event", "stop router").
				WithField("actor_name", name).
				Debugln("router stopped")
		}
	}

	for name, actor := range p.consoles {
		if err = p.stopActor(name, ActorConsole, actor); err != nil {
			logger.WithField("module", "spirit").
				WithField("event", "stop console").
				WithField("actor_name", name).
				Errorln(err)
		} else {
			logger.WithField("module", "spirit").
				WithField("event", "stop console").
				WithField("actor_name", name).
				Debugln("console stopped")
		}
	}

	for name, components := range p.components {
		for _, actor := range components {
			if err = p.stopActor(name, ActorComponent, actor); err != nil {
				logger.WithField("module", "spirit").
					WithField("event", "stop component").
					WithField("actor_name", name).
					Errorln(err)
			} else {
				logger.WithField("module", "spirit").
					WithField("event", "stop component").
					WithField("actor_name", name).
					Debugln("component stopped")
			}
		}
	}

	p.status = StatusStopped

	return
}

func (p *ClassicSpirit) Status() Status {
	return p.status
}

func (p *ClassicSpirit) Build(conf SpiritConfig) (err error) {
	p.statusLocker.Lock()
	defer p.statusLocker.Unlock()

	if p.status != StatusStopped {
		err = ErrSpiritAlreadyRunning
		return
	}

	if p.built {
		err = ErrSpiritAlreadyBuilt
		return
	}

	if err = conf.Validate(); err != nil {
		return
	}

	var logrusMateConf logrus_mate.LogrusMateConfig

	if conf.Log.ConfigFile != "" {
		if logrusMateConf, err = logrus_mate.LoadLogrusMateConfig(conf.Log.ConfigFile); err != nil {
			return
		}

		if logrusMate, err = logrus_mate.NewLogrusMate(logrusMateConf); err != nil {
			return
		}

		defaultLogger := logrusMate.Logger(conf.Log.DefaultLogger)

		if defaultLogger == nil {
			err = ErrDefaultLogerNotExist
			return
		}
		logger = defaultLogger
	}

	p.built = true

	// reader pool
	for _, pool := range conf.ReaderPools {
		var poolActor interface{}
		if poolActor, err = p.createActor(ActorReaderPool, pool.ActorConfig); err != nil {
			logger.WithField("module", "spirit").
				WithField("event", "build").
				WithField("actor_name", pool.Name).
				WithField("actor_urn", pool.URN).
				Error(err)
			return
		}

		readerPool := poolActor.(ReaderPool)

		var actor interface{}
		if actor, err = p.createActor(ActorReader, *pool.Reader); err != nil {
			logger.WithField("module", "spirit").
				WithField("event", "build").
				WithField("actor_name", pool.Reader.Name).
				WithField("actor_urn", pool.Reader.URN).
				Error(err)
			return
		}

		readerPool.SetNewReaderFunc(actor.(NewReaderFunc), pool.Reader.Config)
		p.readerPools[pool.ActorConfig.Name] = readerPool
	}

	// writer pool
	for _, pool := range conf.WriterPools {
		var poolActor interface{}
		if poolActor, err = p.createActor(ActorWriterPool, pool.ActorConfig); err != nil {
			logger.WithField("module", "spirit").
				WithField("event", "build").
				WithField("actor_name", pool.Name).
				WithField("actor_urn", pool.URN).
				Error(err)
			return
		}

		writerPool := poolActor.(WriterPool)

		if pool.Writer != nil {
			var actor interface{}
			if actor, err = p.createActor(ActorWriter, *pool.Writer); err != nil {
				logger.WithField("module", "spirit").
					WithField("event", "build").
					WithField("actor_name", pool.Writer.Name).
					WithField("actor_urn", pool.Writer.URN).
					Error(err)
				return
			}

			writerPool.SetNewWriterFunc(actor.(NewWriterFunc), pool.Writer.Config)
		}

		p.writerPools[pool.ActorConfig.Name] = writerPool
	}

	// input translators
	for _, actorConf := range conf.InputTranslators {
		var actor interface{}
		if actor, err = p.createActor(ActorInputTranslator, actorConf); err != nil {
			logger.WithField("module", "spirit").
				WithField("event", "build").
				WithField("actor_name", actorConf.Name).
				WithField("actor_urn", actorConf.URN).
				Error(err)

			return
		}
		p.inputTranslators[actorConf.Name] = actor.(InputTranslator)
	}

	//output translators
	for _, actorConf := range conf.OutputTranslators {
		var actor interface{}
		if actor, err = p.createActor(ActorOutputTranslator, actorConf); err != nil {
			logger.WithField("module", "spirit").
				WithField("event", "build").
				WithField("actor_name", actorConf.Name).
				WithField("actor_urn", actorConf.URN).
				Error(err)

			return
		}
		p.outputTranslators[actorConf.Name] = actor.(OutputTranslator)
	}

	// receivers
	for _, actorConf := range conf.Receivers {
		var actor interface{}
		if actor, err = p.createActor(ActorReceiver, actorConf); err != nil {
			logger.WithField("module", "spirit").
				WithField("event", "build").
				WithField("actor_name", actorConf.Name).
				WithField("actor_urn", actorConf.URN).
				Error(err)

			return
		}
		p.receivers[actorConf.Name] = actor.(Receiver)
	}

	// inboxes
	for _, actorConf := range conf.Inboxes {
		var actor interface{}
		if actor, err = p.createActor(ActorInbox, actorConf); err != nil {
			logger.WithField("module", "spirit").
				WithField("event", "build").
				WithField("actor_name", actorConf.Name).
				WithField("actor_urn", actorConf.URN).
				Error(err)

			return
		}
		p.inboxes[actorConf.Name] = actor.(Inbox)
	}

	// routers
	for _, actorConf := range conf.Routers {
		var actor interface{}
		if actor, err = p.createActor(ActorRouter, actorConf); err != nil {
			logger.WithField("module", "spirit").
				WithField("event", "build").
				WithField("actor_name", actorConf.Name).
				WithField("actor_urn", actorConf.URN).
				Error(err)

			return
		}
		p.routers[actorConf.Name] = actor.(Router)
	}

	// label matchers
	for _, actorConf := range conf.LabelMatchers {
		var actor interface{}
		if actor, err = p.createActor(ActorLabelMatcher, actorConf); err != nil {
			logger.WithField("module", "spirit").
				WithField("event", "build").
				WithField("actor_name", actorConf.Name).
				WithField("actor_urn", actorConf.URN).
				Error(err)

			return
		}
		p.labelMatchers[actorConf.Name] = actor.(LabelMatcher)
	}

	// urn rewriter
	for _, actorConf := range conf.URNRewriters {
		var actor interface{}
		if actor, err = p.createActor(ActorURNRewriter, actorConf); err != nil {
			logger.WithField("module", "spirit").
				WithField("event", "build").
				WithField("actor_name", actorConf.Name).
				WithField("actor_urn", actorConf.URN).
				Error(err)

			return
		}
		p.urnRewriters[actorConf.Name] = actor.(URNRewriter)
	}

	// components
	for _, actorConf := range conf.Components {
		var actor interface{}
		if actor, err = p.createActor(ActorComponent, actorConf); err != nil {
			logger.WithField("module", "spirit").
				WithField("event", "build").
				WithField("actor_name", actorConf.Name).
				WithField("actor_urn", actorConf.URN).
				Error(err)

			return
		}
		if comps, exist := p.components[actorConf.Name]; exist {
			p.components[actorConf.Name] = append(comps, actor.(Component))
		} else {
			p.components[actorConf.Name] = []Component{actor.(Component)}
		}

	}

	// outboxes
	for _, actorConf := range conf.Outboxes {
		var actor interface{}
		if actor, err = p.createActor(ActorOutbox, actorConf); err != nil {
			logger.WithField("module", "spirit").
				WithField("event", "build").
				WithField("actor_name", actorConf.Name).
				WithField("actor_urn", actorConf.URN).
				Error(err)

			return
		}
		p.outboxes[actorConf.Name] = actor.(Outbox)
	}

	// senders
	for _, actorConf := range conf.Senders {
		var actor interface{}
		if actor, err = p.createActor(ActorSender, actorConf); err != nil {
			logger.WithField("module", "spirit").
				WithField("event", "build").
				WithField("actor_name", actorConf.Name).
				WithField("actor_urn", actorConf.URN).
				Error(err)

			return
		}
		p.senders[actorConf.Name] = actor.(Sender)
	}

	// consoles
	for _, actorConf := range conf.Consoles {
		var actor interface{}
		if actor, err = p.createActor(ActorConsole, actorConf); err != nil {
			logger.WithField("module", "spirit").
				WithField("event", "build").
				WithField("actor_name", actorConf.Name).
				WithField("actor_urn", actorConf.URN).
				Error(err)

			return
		}
		p.consoles[actorConf.Name] = actor.(Console)
	}

	if err = p.buildCompose(conf.Compose); err != nil {
		return
	}

	p.built = true

	return
}

func (p *ClassicSpirit) buildCompose(compose []ComposeRouterConfig) (err error) {
	for _, composeObj := range compose {
		routerInstance := p.routers[composeObj.Router]

		if composeObj.URNRewriter != nil {
			rewriter := p.urnRewriters[*composeObj.URNRewriter]
			if err = routerInstance.SetURNRewriter(rewriter); err != nil {
				return
			}
		}

		for _, compName := range composeObj.Components {
			componentInstances := p.components[compName]
			for _, componentInstance := range componentInstances {
				routerInstance.AddComponent(compName, componentInstance)
			}
		}

		for _, inbox := range composeObj.Inboxes {
			inboxInstance := p.inboxes[inbox.Name]
			routerInstance.AddInbox(inboxInstance)

			for _, receiver := range inbox.Receivers {
				receiverInstance := p.receivers[receiver.Name]

				switch rcver := receiverInstance.(type) {
				case ReadReceiver:
					{
						readerPool := p.readerPools[receiver.ReaderPool]
						translatorInstance := p.inputTranslators[receiver.Translator]

						rcver.SetTranslator(translatorInstance)
						rcver.SetReaderPool(readerPool)
						rcver.SetDeliveryPutter(inboxInstance)
					}
				case Receiver:
					{
						rcver.SetDeliveryPutter(inboxInstance)
					}
				default:
					{
						err = ErrReceiverTypeNotSupport

						logger.WithField("module", "spirit").
							WithField("event", "bind putter to receiver").
							WithField("inbox_name", inbox.Name).
							WithField("receiver_name", receiver.Name).
							Errorln(err)
					}
				}
			}
		}

		componentlabelMatcher := p.labelMatchers[composeObj.LabelMatchers.Component]
		routerInstance.SetComponentLabelMatcher(componentlabelMatcher)

		outboxLabelMatcher := p.labelMatchers[composeObj.LabelMatchers.Outbox]
		routerInstance.SetOutboxLabelMatcher(outboxLabelMatcher)

		for _, outbox := range composeObj.Outboxes {
			outboxInstance := p.outboxes[outbox.Name]
			routerInstance.AddOutbox(outboxInstance)

			senderInstance := p.senders[outbox.Sender.Name]

			switch sdr := senderInstance.(type) {
			case WriteSender:
				{
					translatorInstance := p.outputTranslators[outbox.Sender.Translator]
					writerPool := p.writerPools[outbox.Sender.WriterPool]
					sdr.SetTranslator(translatorInstance)
					sdr.SetWriterPool(writerPool)
					sdr.SetDeliveryGetter(outboxInstance)
				}
			case Sender:
				{
					sdr.SetDeliveryGetter(outboxInstance)
				}
			default:
				{
					err = ErrReceiverTypeNotSupport

					logger.WithField("module", "spirit").
						WithField("event", "bind getter to sender").
						WithField("inbox_name", outbox.Name).
						WithField("receiver_name", outbox.Sender.Name).
						Errorln(err)
				}
			}
		}
	}
	return
}

func (p *ClassicSpirit) createActor(actorType ActorType, actorConf ActorConfig) (actor interface{}, err error) {
	switch actorType {
	case ActorReader:
		{
			if f, exist := newReaderFuncs[actorConf.URN]; exist {
				actor = f
			} else {
				err = ErrReaderURNNotExist
			}
		}
	case ActorWriter:
		{
			if f, exist := newWriterFuncs[actorConf.URN]; exist {
				actor = f
			} else {
				err = ErrWriterURNNotExist
			}
		}
	case ActorInputTranslator:
		{
			actor, err = newInputTranslatorFuncs[actorConf.URN](actorConf.Config)
		}
	case ActorOutputTranslator:
		{
			actor, err = newOutputTranslatorFuncs[actorConf.URN](actorConf.Config)
		}
	case ActorReceiver:
		{
			actor, err = newReceiverFuncs[actorConf.URN](actorConf.Config)
		}
	case ActorSender:
		{
			actor, err = newSenderFuncs[actorConf.URN](actorConf.Config)
		}
	case ActorInbox:
		{
			actor, err = newInboxFuncs[actorConf.URN](actorConf.Config)
		}
	case ActorOutbox:
		{
			actor, err = newOutboxFuncs[actorConf.URN](actorConf.Config)
		}
	case ActorRouter:
		{
			actor, err = newRouterFuncs[actorConf.URN](actorConf.Config)
		}
	case ActorLabelMatcher:
		{
			actor, err = newLabelMatcherFuncs[actorConf.URN](actorConf.Config)
		}
	case ActorURNRewriter:
		{
			actor, err = newURNRewriterFuncs[actorConf.URN](actorConf.Config)
		}
	case ActorComponent:
		{
			actor, err = newComponentFuncs[actorConf.URN](actorConf.Config)
		}
	case ActorReaderPool:
		{
			actor, err = newReaderPoolFuncs[actorConf.URN](actorConf.Config)
		}
	case ActorWriterPool:
		{
			actor, err = newWriterPoolFuncs[actorConf.URN](actorConf.Config)
		}
	case ActorConsole:
		{
			actor, err = newConsoleFuncs[actorConf.URN](actorConf.Config)
		}
	default:
		err = ErrSpiritActorURNNotExist
	}

	return
}

func (p *ClassicSpirit) waitSignal(wg *sync.WaitGroup) {
	var err error
	isStopping := false
	isStopped := false
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	stopingLocker := sync.Mutex{}

	for {
		select {
		case signal := <-interrupt:
			{
				switch signal {
				case os.Interrupt, syscall.SIGTERM:
					{
						stopingLocker.Lock()
						if isStopping {
							logger.WithField("module", "spirit").
								WithField("event", "stop spirit").
								Warnln("kill spirit")

							if !isStopped {
								wg.Done()
								isStopped = true
							}

							stopingLocker.Unlock()
							return
						}
						isStopping = true
						stopingLocker.Unlock()

						logger.WithField("module", "spirit").
							WithField("event", "stop spirit").
							Info("stopping spirit")

						go func() {
							if err = p.stop(); err != nil {
								logger.WithField("module", "spirit").
									WithField("event", "stop spirit").
									Errorln(err)
							}

							stopingLocker.Lock()
							if !isStopped {
								wg.Done()
								isStopped = true
							}
							stopingLocker.Unlock()

							return
						}()
					}
				}
			}
		case <-time.After(time.Second):
			{
				if isStopping && p.status == StatusStopped {
					return
				}
				continue
			}
		}
	}
}
