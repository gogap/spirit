package spirit

import (
	"sync"
)

type NewReceiverFunc func(config Config) (receiver Receiver, err error)

type Receiver interface {
	StartStopper

	SetTranslator(translator InputTranslator) (err error)
	SetDeliveryPutter(putter DeliveryPutter) (err error)
}

type ReadReceiver interface {
	Receiver

	SetReaderPool(pool ReaderPool) (err error)
}

var (
	receiversLocker  = sync.Mutex{}
	newReceiverFuncs = make(map[string]NewReceiverFunc)
)

func RegisterReceiver(urn string, newFunc NewReceiverFunc) (err error) {
	receiversLocker.Lock()
	receiversLocker.Unlock()

	if urn == "" {
		panic("spirit: Register receiver urn is empty")
	}

	if newFunc == nil {
		panic("spirit: Register receiver is nil")
	}

	if _, exist := newReceiverFuncs[urn]; exist {
		panic("spirit: Register called twice for receiver " + urn)
	}

	newReceiverFuncs[urn] = newFunc

	logger.WithField("module", "spirit").
		WithField("event", "register receiver").
		WithField("urn", urn).
		Debugln("receiver registered")

	return
}
