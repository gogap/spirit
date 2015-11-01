package spirit

import (
	"io"
	"sync"
)

type ReaderPool interface {
	SetNewReaderFunc(newFunc NewReaderFunc) (err error)
	Get() (reader io.ReadCloser, err error)
	Put(reader io.ReadCloser)
	Close()
}

var (
	readerPoolsLocker  = sync.Mutex{}
	newReaderPoolFuncs = make(map[string]NewReaderPoolFunc)
)

type NewReaderPoolFunc func(options Options) (pool ReaderPool, err error)

func RegisterReaderPool(urn string, newFunc NewReaderPoolFunc) (err error) {
	readerPoolsLocker.Lock()
	readerPoolsLocker.Unlock()

	if urn == "" {
		panic("spirit: Register reader pool urn is empty")
	}

	if newFunc == nil {
		panic("spirit: Register reader pool is nil")
	}

	if _, exist := newReaderPoolFuncs[urn]; exist {
		panic("spirit: Register called twice for reader pool " + urn)
	}

	newReaderPoolFuncs[urn] = newFunc

	logger.WithField("module", "spirit").
		WithField("event", "register reader pool").
		WithField("urn", urn).
		Debugln("reader pool registered")

	return
}
