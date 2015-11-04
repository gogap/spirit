package classic

import (
	"regexp"
	"sync"

	"github.com/gogap/spirit"
)

const (
	classicReceiverURN = "urn:spirit:router:classic"
)

var _ spirit.Router = new(ClassicRouter)

type ClassicRouterConfig struct {
	AllowNoComponent bool `json:"allow_no_component"`
}

type ClassicRouter struct {
	conf ClassicRouterConfig

	inboxes  []spirit.Inbox
	outboxes []spirit.Outbox

	components map[string][]spirit.Component

	inboxesLocker   sync.Mutex
	outboxesLocker  sync.Mutex
	componentLocker sync.Mutex

	status       spirit.Status
	statusLocker sync.Mutex

	outboxLabelMatcher    spirit.LabelMatcher
	componentLabelMatcher spirit.LabelMatcher
}

func init() {
	spirit.RegisterRouter(classicReceiverURN, NewClassicRouter)
}

func NewClassicRouter(options spirit.Options) (box spirit.Router, err error) {
	conf := ClassicRouterConfig{}
	if err = options.ToObject(&conf); err != nil {
		return
	}

	box = &ClassicRouter{
		conf:       conf,
		components: make(map[string][]spirit.Component),
	}

	return
}

func (p *ClassicRouter) Start() (err error) {
	p.statusLocker.Lock()
	defer p.statusLocker.Unlock()

	if p.status == spirit.StatusRunning {
		err = spirit.ErrRouterAlreadyRunning
		return
	}

	if err = p.startOutboxes(); err != nil {
		return
	}

	if err = p.startInboxes(); err != nil {
		return
	}

	p.status = spirit.StatusRunning

	return
}

func (p *ClassicRouter) startInboxes() (err error) {
	for _, inbox := range p.inboxes {
		if err = inbox.Start(); err != nil {
			return
		}
	}
	return
}

func (p *ClassicRouter) startOutboxes() (err error) {
	for _, outbox := range p.outboxes {
		if err = outbox.Start(); err != nil {
			return
		}
	}
	return
}

func (p *ClassicRouter) Stop() (err error) {
	p.statusLocker.Lock()
	defer p.statusLocker.Unlock()

	if p.status == spirit.StatusStopped {
		err = spirit.ErrRouterDidNotRunning
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go p.stopInboxes()

	wg.Add(1)
	go p.stopOutboxes()

	wg.Wait()

	p.status = spirit.StatusStopped

	return
}

func (p *ClassicRouter) stopInboxes() (err error) {
	wg := sync.WaitGroup{}
	for _, inbox := range p.inboxes {
		wg.Add(1)
		go func(inbox spirit.Inbox) {
			defer wg.Done()
			inbox.Stop()
		}(inbox)
	}
	wg.Wait()
	return
}

func (p *ClassicRouter) stopOutboxes() (err error) {
	wg := sync.WaitGroup{}
	for _, outbox := range p.outboxes {
		wg.Add(1)
		go func(outbox spirit.Outbox) {
			defer wg.Done()
			outbox.Stop()
		}(outbox)
	}
	wg.Wait()
	return
}

func (p *ClassicRouter) Status() spirit.Status {
	return p.status
}

func (p *ClassicRouter) AddInbox(inbox spirit.Inbox) (err error) {
	if inbox == nil {
		return
	}

	p.inboxesLocker.Lock()
	defer p.inboxesLocker.Unlock()

	for _, oldInbox := range p.inboxes {
		if oldInbox == inbox {
			err = spirit.ErrRouterAlreadyHaveThisInbox
			return
		}
	}

	p.inboxes = append(p.inboxes, inbox)

	return
}

func (p *ClassicRouter) RemoveInbox(inbox spirit.Inbox) (err error) {
	p.inboxesLocker.Lock()
	defer p.inboxesLocker.Unlock()

	var newInboxes []spirit.Inbox

	for _, oldInbox := range p.inboxes {
		if oldInbox != inbox {
			newInboxes = append(newInboxes, inbox)
		}
	}

	return
}

func (p *ClassicRouter) Inboxes() (inboxes []spirit.Inbox) {
	return p.inboxes
}

func (p *ClassicRouter) AddOutbox(outbox spirit.Outbox) (err error) {
	if outbox == nil {
		return
	}

	p.outboxesLocker.Lock()
	defer p.outboxesLocker.Unlock()

	for _, oldOutbox := range p.outboxes {
		if oldOutbox == outbox {
			err = spirit.ErrRouterAlreadyHaveThisOutbox
			return
		}
	}

	p.outboxes = append(p.outboxes, outbox)

	return
}

func (p *ClassicRouter) RemoveOutbox(outbox spirit.Outbox) (err error) {
	if outbox == nil {
		return
	}

	p.outboxesLocker.Lock()
	defer p.outboxesLocker.Unlock()

	var newOutboxes []spirit.Outbox

	for _, oldOutbox := range p.outboxes {
		if oldOutbox != outbox {
			newOutboxes = append(newOutboxes, outbox)
		}
	}

	return
}

func (p *ClassicRouter) Outboxes() (outboxes []spirit.Outbox) {
	return p.outboxes
}

func (p *ClassicRouter) AddComponent(name string, component spirit.Component) (err error) {
	if component == nil {
		return
	}

	if component.URN() == "" {
		err = spirit.ErrComponentURNIsEmpty
		return
	}

	if name == "" {
		err = spirit.ErrComponentNameIsEmpty
		return
	}

	p.componentLocker.Lock()
	defer p.componentLocker.Unlock()

	if comps, exist := p.components[component.URN()]; exist {
		for _, comp := range comps {
			if comp == component ||
				p.componentLabelMatcher.Match(comp.Labels(), component.Labels()) {
				err = spirit.ErrRouterAlreadyHaveThisComponent
				return
			}
		}

		p.components[component.URN()] = append(comps, component)
	} else {
		p.components[component.URN()] = []spirit.Component{component}
	}

	return
}

func (p *ClassicRouter) RemoveComponent(urn string) (err error) {
	if urn == "" {
		return
	}

	p.componentLocker.Lock()
	defer p.componentLocker.Unlock()

	if _, exist := p.components[urn]; exist {
		delete(p.components, urn)
	}

	err = spirit.ErrComponentNotExist

	return
}

func (p *ClassicRouter) Components() (components map[string][]spirit.Component) {
	return p.components
}

func (p *ClassicRouter) RouteToHandlers(delivery spirit.Delivery) (handlers []spirit.ComponentHandler, err error) {
	urn := delivery.URN()

	if urn == "" {
		err = spirit.ErrDeliveryURNIsEmpty
		return
	}

	componentURN := ""
	componentHandlerURN := ""

	regURN := regexp.MustCompile("(.*)#(.*)")
	regMatched := regURN.FindAllStringSubmatch(urn, -1)

	if len(regMatched) == 1 &&
		len(regMatched[0]) == 3 {
		componentURN = regMatched[0][1]
		componentHandlerURN = regMatched[0][2]
	}

	var components []spirit.Component
	var exist bool

	if components, exist = p.components[componentURN]; !exist {
		if !p.conf.AllowNoComponent {
			err = spirit.ErrRouterComponentNotExist
			return
		}
	}

	lenComps := len(components)

	if lenComps == 0 {
		if !p.conf.AllowNoComponent {
			err = spirit.ErrRouterToComponentHandlerFailed
		}
		return
	}

	if p.componentLabelMatcher == nil {
		if lenComps > 1 {
			err = spirit.ErrRouterDidNotHaveComponentLabelMatcher
			return
		}

		if h, exist := components[0].Handlers()[componentHandlerURN]; !exist {
			err = spirit.ErrComponentHandlerNotExit
			return
		} else {
			handlers = []spirit.ComponentHandler{h}
			return
		}
	}

	for _, component := range components {
		if p.componentLabelMatcher.Match(delivery.Labels(), component.Labels()) {
			if h, exist := component.Handlers()[componentHandlerURN]; !exist {
				err = spirit.ErrComponentHandlerNotExit
				return
			} else {
				handlers = []spirit.ComponentHandler{h}
				return
			}
		}
	}

	return
}

func (p *ClassicRouter) RouteToOutboxes(delivery spirit.Delivery) (outboxes []spirit.Outbox, err error) {
	boxes := []spirit.Outbox{}

	if p.outboxLabelMatcher != nil {
		for _, outbox := range p.outboxes {
			if p.outboxLabelMatcher.Match(delivery.Labels(), outbox.Labels()) {
				boxes = append(boxes, outbox)
			}
		}
	} else if len(p.outboxes) == 1 {
		boxes = append(boxes, p.outboxes...)
	}

	outboxes = boxes

	return
}

func (p *ClassicRouter) SetOutboxLabelMatcher(matcher spirit.LabelMatcher) (err error) {
	p.outboxLabelMatcher = matcher
	return
}

func (p *ClassicRouter) SetComponentLabelMatcher(matcher spirit.LabelMatcher) (err error) {
	p.componentLabelMatcher = matcher
	return
}
