package spirit

import (
	"io"
	"sync"
)

var (
	readersLocker  = sync.Mutex{}
	newReaderFuncs = make(map[string]NewReaderFunc)
)

type NewReaderFunc func(name string, options Map) (r io.ReadCloser, err error)

func RegisterReader(urn string, newFunc NewReaderFunc) (err error) {
	readersLocker.Lock()
	readersLocker.Unlock()

	if urn == "" {
		panic("spirit: Register reader urn is empty")
	}

	if newFunc == nil {
		panic("spirit: Register reader is nil")
	}

	if _, exist := newReaderFuncs[urn]; exist {
		panic("spirit: Register called twice for reader " + urn)
	}

	newReaderFuncs[urn] = newFunc

	logger.WithField("module", "spirit").
		WithField("event", "register reader").
		WithField("urn", urn).
		Debugln("reader registered")

	return
}
