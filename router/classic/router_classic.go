package classic

import (
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/gogap/spirit"
)

const (
	classicRouterURN = "urn:spirit:router:classic"
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

	componentHandlers map[spirit.Component]spirit.Handlers
	componentLabels   map[spirit.Component]spirit.Labels

	inboxesLocker   sync.Mutex
	outboxesLocker  sync.Mutex
	componentLocker sync.Mutex

	status       spirit.Status
	statusLocker sync.Mutex

	outboxLabelMatcher    spirit.LabelMatcher
	componentLabelMatcher spirit.LabelMatcher

	urnRewriter spirit.URNRewriter
}

func init() {
	spirit.RegisterRouter(classicRouterURN, NewClassicRouter)
}

func NewClassicRouter(config spirit.Map) (box spirit.Router, err error) {
	conf := ClassicRouterConfig{}
	if err = config.ToObject(&conf); err != nil {
		return
	}

	box = &ClassicRouter{
		conf:              conf,
		components:        make(map[string][]spirit.Component),
		componentHandlers: make(map[spirit.Component]spirit.Handlers),
		componentLabels:   make(map[spirit.Component]spirit.Labels),
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

	p.stopInboxes(&wg)
	p.stopOutboxes(&wg)

	wg.Wait()

	p.status = spirit.StatusStopped

	return
}

func (p *ClassicRouter) stopInboxes(wg *sync.WaitGroup) {
	for _, inbox := range p.inboxes {
		wg.Add(1)
		go func(inbox spirit.Inbox) {
			defer wg.Done()
			if err := inbox.Stop(); err != nil {
				spirit.Logger().WithField("actor", spirit.ActorRouter).
					WithField("urn", classicRouterURN).
					WithField("event", "stop inbox").
					Errorln(err)
			}
		}(inbox)
	}
	return
}

func (p *ClassicRouter) stopOutboxes(wg *sync.WaitGroup) {
	for _, outbox := range p.outboxes {
		wg.Add(1)
		go func(outbox spirit.Outbox) {
			defer wg.Done()
			if err := outbox.Stop(); err != nil {
				spirit.Logger().WithField("actor", spirit.ActorRouter).
					WithField("urn", classicRouterURN).
					WithField("event", "stop outbox").
					Errorln(err)
			}
		}(outbox)
	}
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

	handlers := detectComponentHandlers(component)
	if handlers == nil || len(handlers) == 0 {
		err = spirit.ErrComponentDidNotHaveHandler
		return
	}

	labels := detectComponentLabels(component)

	p.componentLocker.Lock()
	defer p.componentLocker.Unlock()

	namedURN := name + "@" + component.URN()
	isGlobalInstance := false

	if name[0] == '_' {
		isGlobalInstance = true
	}

	if comps, exist := p.components[namedURN]; exist {
		for _, comp := range comps {

			compLabels := detectComponentLabels(comp)

			if comp == component ||
				p.componentLabelMatcher.Match(compLabels, labels) {
				err = spirit.ErrRouterAlreadyHaveThisComponent
				return
			}
		}
		p.components[namedURN] = append(comps, component)
	} else {
		p.components[namedURN] = []spirit.Component{component}
	}

	if isGlobalInstance {
		if comps, exist := p.components[component.URN()]; exist {
			if len(comps) > 0 {
				err = spirit.ErrRouterOnlyOneGlobalComponentAllowed
				return
			}
		}
		p.components[component.URN()] = p.components[namedURN]
	}

	if _, exist := p.componentHandlers[component]; !exist {
		p.componentHandlers[component] = handlers
	} else {
		spirit.Logger().WithField("actor", spirit.ActorRouter).
			WithField("urn", classicRouterURN).
			WithField("event", "bind component handlers").
			WithField("component_urn", component.URN()).
			Warnln("component handler already exist")
	}

	if _, exist := p.componentLabels[component]; !exist {
		p.componentLabels[component] = labels
	} else {
		spirit.Logger().WithField("actor", spirit.ActorRouter).
			WithField("urn", classicRouterURN).
			WithField("event", "bind component labels").
			WithField("component_urn", component.URN()).
			Warnln("component label already exist")
	}

	return
}

func detectComponentHandlers(component spirit.Component) (handlers spirit.Handlers) {
	compValue := reflect.ValueOf(component)
	handlers = make(spirit.Handlers)

	switch handlerLister := compValue.Interface().(type) {
	case spirit.HandlerLister:
		{

			for name, funcV := range handlerLister.Handlers() {
				handlers[name] = funcV
			}

			spirit.Logger().WithField("actor", spirit.ActorRouter).
				WithField("urn", classicRouterURN).
				WithField("event", "bind component handlers").
				WithField("component_urn", component.URN()).
				Debugln("bind handler's by Handlers()")
		}
	default:
		compType := reflect.TypeOf(component)
		for i := 0; i < compType.NumMethod(); i++ {
			if compValue.Method(i).Type().ConvertibleTo(reflect.TypeOf(new(spirit.ComponentHandler)).Elem()) {
				handler := compValue.Method(i).Convert(reflect.TypeOf(new(spirit.ComponentHandler)).Elem())
				handlers[compType.Method(i).Name] = handler.Interface().(spirit.ComponentHandler)
			}
		}

		spirit.Logger().WithField("actor", spirit.ActorRouter).
			WithField("urn", classicRouterURN).
			WithField("event", "bind component handlers").
			WithField("component_urn", component.URN()).
			Debugln("bind handler's by Reflact")
	}

	return
}

func detectComponentLabels(component spirit.Component) spirit.Labels {
	compValue := reflect.ValueOf(component)
	labels := make(spirit.Labels)

	switch labelLister := compValue.Interface().(type) {
	case spirit.LabelLister:
		{
			for name, funcV := range labelLister.Labels() {
				labels[name] = funcV
			}

			spirit.Logger().WithField("actor", spirit.ActorRouter).
				WithField("urn", classicRouterURN).
				WithField("event", "bind component labels").
				WithField("component_urn", component.URN()).
				Debugln("bind label's by Labels()")
		}
	default:
		spirit.Logger().WithField("actor", spirit.ActorRouter).
			WithField("urn", classicRouterURN).
			WithField("event", "bind component labels").
			WithField("component_urn", component.URN()).
			Debugln("label not exist")
	}

	return labels
}

func (p *ClassicRouter) RemoveComponent(urn string) (err error) {
	if urn == "" {
		return
	}

	p.componentLocker.Lock()
	defer p.componentLocker.Unlock()

	if components, exist := p.components[urn]; exist {
		for _, component := range components {
			delete(p.componentHandlers, component)
			delete(p.componentLabels, component)
		}
		delete(p.components, urn)
	}

	err = spirit.ErrComponentNotExist

	return
}

func (p *ClassicRouter) Components() (components map[string][]spirit.Component) {
	return p.components
}

func (p *ClassicRouter) RouteToHandlers(delivery spirit.Delivery) (handlers []spirit.ComponentHandler, err error) {
	strURNs := delivery.URN()
	if p.urnRewriter != nil {
		if strURNs, err = p.urnRewriter.Rewrite(delivery); err != nil {
			return
		}
	}

	if strURNs == "" {
		if !p.conf.AllowNoComponent {
			err = spirit.ErrDeliveryURNIsEmpty
			return
		}
	}

	var urns = []string{}
	var tmpHandlers = []spirit.ComponentHandler{}

	if strURNs != "" {
		urns = strings.Split(strURNs, "|")

		for _, urn := range urns {
			componentName := ""
			componentURN := ""
			componentHandlerURN := ""
			componentIndex := ""

			regURN := regexp.MustCompile("(.*)@(.*)#(.*)")
			regMatched := regURN.FindAllStringSubmatch(urn, -1)

			if len(regMatched) == 1 &&
				len(regMatched[0]) == 4 {
				componentName = regMatched[0][1]
				componentURN = regMatched[0][2]
				componentHandlerURN = regMatched[0][3]

				componentIndex = componentName + "@" + componentURN

			} else {
				regURN = regexp.MustCompile("(.*)#(.*)")
				regMatched = regURN.FindAllStringSubmatch(urn, -1)

				if len(regMatched) == 1 &&
					len(regMatched[0]) == 3 {
					componentURN = regMatched[0][1]
					componentHandlerURN = regMatched[0][2]

					componentIndex = componentURN
				}
			}

			if componentIndex == "" {
				err = spirit.ErrRouterDeliveryURNFormatError
				return
			}

			var components []spirit.Component
			var exist bool

			if components, exist = p.components[componentIndex]; !exist {
				if !p.conf.AllowNoComponent {
					err = spirit.ErrRouterComponentNotExist
					return
				}
				continue
			}

			lenComps := len(components)

			if lenComps == 0 {
				if !p.conf.AllowNoComponent {
					err = spirit.ErrRouterToComponentHandlerFailed
					return
				}
				continue
			}

			if p.componentLabelMatcher == nil {
				if lenComps > 1 {
					err = spirit.ErrRouterDidNotHaveComponentLabelMatcher
					return
				}

				component := components[0]

				if componentHandlers, exist := p.componentHandlers[component]; exist {
					if h, exist := componentHandlers[componentHandlerURN]; !exist {
						err = spirit.ErrComponentHandlerNotExit
						return
					} else {
						tmpHandlers = append(tmpHandlers, h)
						continue
					}
				}
			}

			for _, component := range components {
				if p.componentLabelMatcher.Match(delivery.Labels(), p.componentLabels[component]) {
					if componentHandlers, exist := p.componentHandlers[component]; exist {
						if h, exist := componentHandlers[componentHandlerURN]; !exist {
							err = spirit.ErrComponentHandlerNotExit
							return
						} else {
							tmpHandlers = append(tmpHandlers, h)
							break
						}
					}
				}
			}
		}
	}

	if len(tmpHandlers) == 0 && p.conf.AllowNoComponent {
		return
	}

	if len(tmpHandlers) != len(urns) {
		err = spirit.ErrRouterHandlerCountNotEqualURNsCount
		return
	}

	handlers = tmpHandlers
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

func (p *ClassicRouter) SetURNRewriter(rewriter spirit.URNRewriter) (err error) {
	p.urnRewriter = rewriter
	return
}
