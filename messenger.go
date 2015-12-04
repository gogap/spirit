package spirit

import (
	"sync"
)

type Messenger interface {
	Actor
	StartStopper

	SetRouter(router Router) (err error)
}

var (
	messengersLocker  = sync.Mutex{}
	newMessengerFuncs = make(map[string]NewMessengerFunc)
)

type NewMessengerFunc func(name string, options Map) (messenger Messenger, err error)

func RegisterMessenger(urn string, newFunc NewMessengerFunc) (err error) {
	messengersLocker.Lock()
	messengersLocker.Unlock()

	if urn == "" {
		panic("spirit: Register urn messenger's urn is empty")
	}

	if newFunc == nil {
		panic("spirit: Register urn messenger is nil")
	}

	if _, exist := newMessengerFuncs[urn]; exist {
		panic("spirit: Register called twice for urn messenger " + urn)
	}

	newMessengerFuncs[urn] = newFunc

	logger.WithField("module", "spirit").
		WithField("event", "register urn messenger").
		WithField("urn", urn).
		Debugln("urn messenger registered")

	return
}
