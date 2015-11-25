package spirit

import (
	"sync"
)

var (
	inboxesLocker = sync.Mutex{}
	newInboxFuncs = make(map[string]NewInboxFunc)
)

type NewInboxFunc func(config Map) (inbox Inbox, err error)

type PutMessageFunc func(deliveries []Delivery) (err error)

type DeliveryPutter interface {
	Put(deliveries []Delivery) (err error)
}

type DeliveryGetter interface {
	Get() (deliveries []Delivery, err error)
}

type Inbox interface {
	StartStopper

	// AddReceiver(receiver Receiver) (err error)

	DeliveryPutter
	DeliveryGetter
}

func RegisterInbox(urn string, newFunc NewInboxFunc) (err error) {
	inboxesLocker.Lock()
	inboxesLocker.Unlock()

	if urn == "" {
		panic("spirit: Register inbox urn is empty")
	}

	if newFunc == nil {
		panic("spirit: Register inbox is nil")
	}

	if _, exist := newInboxFuncs[urn]; exist {
		panic("spirit: Register called twice for inbox " + urn)
	}

	newInboxFuncs[urn] = newFunc

	logger.WithField("module", "spirit").
		WithField("event", "register inbox").
		WithField("urn", urn).
		Debugln("inbox registered")

	return
}
