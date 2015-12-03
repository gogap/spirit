package spirit

import (
	"sync"
)

type Console interface {
	AddSpirit(spirit Spirit) (err error)
}

var (
	consolesLocker  = sync.Mutex{}
	newConsoleFuncs = make(map[string]NewConsoleFunc)
)

type NewConsoleFunc func(name string, options Map) (console Console, err error)

func RegisterConsole(urn string, newFunc NewConsoleFunc) (err error) {
	consolesLocker.Lock()
	consolesLocker.Unlock()

	if urn == "" {
		panic("spirit: Register urn console's urn is empty")
	}

	if newFunc == nil {
		panic("spirit: Register urn console is nil")
	}

	if _, exist := newConsoleFuncs[urn]; exist {
		panic("spirit: Register called twice for urn console " + urn)
	}

	newConsoleFuncs[urn] = newFunc

	logger.WithField("module", "spirit").
		WithField("event", "register urn console").
		WithField("urn", urn).
		Debugln("urn console registered")

	return
}
