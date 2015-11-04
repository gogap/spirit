package spirit

import (
	"sync"
)

type ComponentHandler func(payload Payload) (result interface{}, err error)

type Handlers map[string]ComponentHandler

type Component interface {
	URN() string
	Labels() Labels
	Handlers() Handlers
}

type ServiceComponent interface {
	StartStoper

	Component
}

var (
	componentsLocker  = sync.Mutex{}
	newComponentFuncs = make(map[string]NewComponentFunc)
)

type NewComponentFunc func(options Options) (component Component, err error)

func RegisterComponent(urn string, newFunc NewComponentFunc) (err error) {
	componentsLocker.Lock()
	componentsLocker.Unlock()

	if urn == "" {
		panic("spirit: Register component urn is empty")
	}

	if newFunc == nil {
		panic("spirit: Register component is nil")
	}

	if _, exist := newComponentFuncs[urn]; exist {
		panic("spirit: Register called twice for component " + urn)
	}

	newComponentFuncs[urn] = newFunc

	logger.WithField("module", "spirit").
		WithField("event", "register component").
		WithField("urn", urn).
		Debugln("component registered")

	return
}
