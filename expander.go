package spirit

import (
	"sync"
)

type URNExpander interface {
	Expand(delivery Delivery) (newURN string, err error)
}

var (
	expandersLocker     = sync.Mutex{}
	newURNExpanderFuncs = make(map[string]NewURNExpanderFunc)
)

type NewURNExpanderFunc func(options Options) (matcher URNExpander, err error)

func RegisterURNExpander(urn string, newFunc NewURNExpanderFunc) (err error) {
	expandersLocker.Lock()
	expandersLocker.Unlock()

	if urn == "" {
		panic("spirit: Register urn expander's urn is empty")
	}

	if newFunc == nil {
		panic("spirit: Register urn expander is nil")
	}

	if _, exist := newURNExpanderFuncs[urn]; exist {
		panic("spirit: Register called twice for urn expander " + urn)
	}

	newURNExpanderFuncs[urn] = newFunc

	logger.WithField("module", "spirit").
		WithField("event", "register urn expander").
		WithField("urn", urn).
		Debugln("urn expander registered")

	return
}
