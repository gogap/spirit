package spirit

import (
	"sync"

	"github.com/Sirupsen/logrus"
)

var (
	routersLocker  = sync.Mutex{}
	newRouterFuncs = make(map[string]NewRouterFunc)
)

type NewRouterFunc func(options Options) (router Router, err error)

type Router interface {
	StartStoper

	AddInbox(inbox Inbox) (err error)
	RemoveInbox(inbox Inbox) (err error)
	Inboxes() []Inbox

	AddOutbox(outbox Outbox) (err error)
	RemoveOutbox(outbox Outbox) (err error)
	Outboxes() []Outbox

	AddComponent(urn string, component Component) (err error)
	RemoveComponent(urn string) (err error)
	Components() (components map[string][]Component)

	RouteToHandlers(delivery Delivery) (handlers []ComponentHandler, err error)
	RouteToOutboxes(delivery Delivery) (outboxes []Outbox, err error)

	SetOutboxLabelMatcher(matcher LabelMatcher) (err error)
	SetComponentLabelMatcher(matcher LabelMatcher) (err error)
}

func RegisterRouter(urn string, newFunc NewRouterFunc) (err error) {
	routersLocker.Lock()
	routersLocker.Unlock()

	if urn == "" {
		logger.WithFields(logrus.Fields{"module": "spirit"}).Panicln("Register router urn is empty")
	}

	if newFunc == nil {
		logger.WithFields(logrus.Fields{"module": "spirit"}).Panicln("Register router is nil")
	}

	if _, exist := newRouterFuncs[urn]; exist {
		logger.WithFields(logrus.Fields{"module": "spirit", "router": urn}).Panicln("Register router called twice for same router")
	}

	newRouterFuncs[urn] = newFunc

	logger.WithField("module", "spirit").
		WithField("event", "register router").
		WithField("urn", urn).
		Debugln("router registered")

	return
}
