package spirit

import (
	"sync"
)

type NewReceiverFunc func(options Options) (receiver Receiver, err error)

type Receiver interface {
	StartStoper

	SetNewReaderFunc(newFunc NewReaderFunc, options Options) (err error)
	SetTranslator(translator InputTranslator) (err error)
	SetDeliveryPutter(putter DeliveryPutter) (err error)
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
