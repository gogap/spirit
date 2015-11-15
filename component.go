package spirit

import (
	"sync"
)

type ComponentHandler func(payload Payload) (result interface{}, err error)

type Handlers map[string]ComponentHandler

type Component interface {
	URN() string
}

type HandlerLister interface {
	Handlers() Handlers
}

var (
	componentsLocker  = sync.Mutex{}
	newComponentFuncs = make(map[string]NewComponentFunc)
)

type NewComponentFunc func(config Config) (component Component, err error)

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
