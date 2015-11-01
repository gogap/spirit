package spirit

import (
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/gogap/logrus_mate"
)

var (
	logger = logrus_mate.Logger()
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
)

func Logger() *logrus.Logger {
	return logger
}

type StartStoper interface {
	Start() (err error)
	Stop() (err error)
	Status() Status
}

type Spirit interface {
	Build(conf SpiritConfig) (err error)
	StartStoper
}

type ClassicSpirit struct {
	statusLocker sync.Mutex

	built bool

	status Status

	readers           map[string]NewReaderFunc
	writers           map[string]NewWriterFunc
	inputTranslators  map[string]InputTranslator
	outputTranslators map[string]OutputTranslator
	inboxes           map[string]Inbox
	outboxes          map[string]Outbox
	receivers         map[string]Receiver
	senders           map[string]Sender
	labelMatchers     map[string]LabelMatcher
	routers           map[string]Router
	components        map[string][]Component

	readerOpts map[string]Options
	writerOpts map[string]Options
}

func NewClassicSpirit() (s Spirit, err error) {
	return &ClassicSpirit{
		readers:           make(map[string]NewReaderFunc),
		writers:           make(map[string]NewWriterFunc),
		inputTranslators:  make(map[string]InputTranslator),
		outputTranslators: make(map[string]OutputTranslator),
		inboxes:           make(map[string]Inbox),
		outboxes:          make(map[string]Outbox),
		receivers:         make(map[string]Receiver),
		senders:           make(map[string]Sender),
		labelMatchers:     make(map[string]LabelMatcher),
		routers:           make(map[string]Router),
		components:        make(map[string][]Component),
		readerOpts:        make(map[string]Options),
		writerOpts:        make(map[string]Options),
	}, nil
}

func (p *ClassicSpirit) Start() (err error) {
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

	for _, router := range p.routers {
		if err = router.Start(); err != nil {
			logger.WithField("module", "spirit").
				WithField("event", "start").
				Panic(err)
		}
	}

	p.status = StatusRunning

	return
}

func (p *ClassicSpirit) Stop() (err error) {
	p.statusLocker.Lock()
	defer p.statusLocker.Unlock()

	if p.status == StatusStopped {
		err = ErrSpiritAlreadyStopped
		return
	}

	for _, router := range p.routers {
		if err = router.Stop(); err != nil {
			logger.WithField("module", "spirit").
				WithField("event", "stop").
				Error(err)
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

	p.built = true

	// readers
	for _, actorConf := range conf.Readers {
		var actor interface{}
		if actor, err = p.createActor(ActorReader, actorConf); err != nil {
			logger.WithField("module", "spirit").
				WithField("event", "build").
				WithField("actor_name", actorConf.Name).
				WithField("actor_urn", actorConf.URN).
				Error(err)

			return
		}
		p.readers[actorConf.Name] = actor.(NewReaderFunc)
		p.readerOpts[actorConf.Name] = actorConf.Options
	}

	// writers
	for _, actorConf := range conf.Writers {
		var actor interface{}
		if actor, err = p.createActor(ActorWriter, actorConf); err != nil {
			logger.WithField("module", "spirit").
				WithField("event", "build").
				WithField("actor_name", actorConf.Name).
				WithField("actor_urn", actorConf.URN).
				Error(err)

			return
		}
		p.writers[actorConf.Name] = actor.(NewWriterFunc)
		p.writerOpts[actorConf.Name] = actorConf.Options
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

	if err = p.buildCompose(conf.Compose); err != nil {
		return
	}

	p.built = true

	return
}

func (p *ClassicSpirit) buildCompose(compose []ComposeRouterConfig) (err error) {
	for _, router := range compose {
		routerInstance := p.routers[router.Name]
		for _, compName := range router.Components {
			componentInstances := p.components[compName]
			for _, componentInstance := range componentInstances {
				routerInstance.AddComponent(compName, componentInstance)
			}
		}

		for _, inbox := range router.Inboxes {

			inboxInstance := p.inboxes[inbox.Name]
			routerInstance.AddInbox(inboxInstance)

			for _, receiver := range inbox.Receivers {

				receiverInstance := p.receivers[receiver.Name]

				translatorInstance := p.inputTranslators[receiver.Translator]
				newReaderFunc := p.readers[receiver.Reader]
				readerOpts := p.readerOpts[receiver.Reader]

				receiverInstance.SetTranslator(translatorInstance)
				receiverInstance.SetNewReaderFunc(newReaderFunc, readerOpts)
				receiverInstance.SetDeliveryPutter(inboxInstance)

				inboxInstance.AddReceiver(receiverInstance)
			}

		}

		componentlabelMatcher := p.labelMatchers[router.LabelMatchers.Component]
		routerInstance.SetComponentLabelMatcher(componentlabelMatcher)

		outboxLabelMatcher := p.labelMatchers[router.LabelMatchers.Outbox]
		routerInstance.SetOutboxLabelMatcher(outboxLabelMatcher)

		for _, outbox := range router.Outboxes {

			outboxInstance := p.outboxes[outbox.Name]
			routerInstance.AddOutbox(outboxInstance)

			for _, sender := range outbox.Senders {

				senderInstance := p.senders[sender.Name]

				translatorInstance := p.outputTranslators[sender.Translator]
				newWriterFunc := p.writers[sender.Writer]

				writerOpts := p.writerOpts[sender.Writer]

				senderInstance.SetTranslator(translatorInstance)
				senderInstance.SetNewWriterFunc(newWriterFunc, writerOpts)
				senderInstance.SetDeliveryGetter(outboxInstance)

				outboxInstance.AddSender(senderInstance)
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
			actor, err = newInputTranslatorFuncs[actorConf.URN](actorConf.Options)
		}
	case ActorOutputTranslator:
		{
			actor, err = newOutputTranslatorFuncs[actorConf.URN](actorConf.Options)
		}
	case ActorReceiver:
		{
			actor, err = newReceiverFuncs[actorConf.URN](actorConf.Options)
		}
	case ActorSender:
		{
			actor, err = newSenderFuncs[actorConf.URN](actorConf.Options)
		}
	case ActorInbox:
		{
			actor, err = newInboxFuncs[actorConf.URN](actorConf.Options)
		}
	case ActorOutbox:
		{
			actor, err = newOutboxFuncs[actorConf.URN](actorConf.Options)
		}
	case ActorRouter:
		{
			actor, err = newRouterFuncs[actorConf.URN](actorConf.Options)
		}
	case ActorLabelMatcher:
		{
			actor, err = newLabelMatcherFuncs[actorConf.URN](actorConf.Options)
		}
	case ActorComponent:
		{
			actor, err = newComponentFuncs[actorConf.URN](actorConf.Options)
		}
	default:
		err = ErrSpiritActorURNNotExist
	}

	return
}
