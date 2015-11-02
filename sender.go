package spirit

import (
	"sync"
)

type NewSenderFunc func(options Options) (sender Sender, err error)

type Sender interface {
	Start() (err error)
	Stop() (err error)
	Status() Status

	SetWriterPool(pool WriterPool) (err error)
	SetTranslator(translator OutputTranslator) (err error)
	SetDeliveryGetter(getter DeliveryGetter) (err error)
}

var (
	sendersLocker  = sync.Mutex{}
	newSenderFuncs = make(map[string]NewSenderFunc)
)

func RegisterSender(urn string, newFunc NewSenderFunc) (err error) {
	sendersLocker.Lock()
	sendersLocker.Unlock()

	if urn == "" {
		logger.WithField("module", "spirit").Panicln("Register sender urn is empty")
	}

	if newFunc == nil {
		logger.WithField("module", "spirit").Panicln("Register sender is nil")
	}

	if _, exist := newSenderFuncs[urn]; exist {
		logger.WithField("module", "spirit").WithField("router", urn).Panicln("Register router called twice for same sender")
	}

	newSenderFuncs[urn] = newFunc

	logger.WithField("module", "spirit").
		WithField("event", "register sender").
		WithField("urn", urn).
		Debugln("sender registered")

	return
}
