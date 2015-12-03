package spirit

import (
	"io"
	"sync"
)

type WriterPool interface {
	SetNewWriterFunc(newFunc NewWriterFunc, config Map) (err error)
	Get(delivery Delivery) (writer io.WriteCloser, err error)
	Put(delivery Delivery, writer io.WriteCloser) (err error)
	Close()
}

var (
	writerPoolsLocker  = sync.Mutex{}
	newWriterPoolFuncs = make(map[string]NewWriterPoolFunc)
)

type NewWriterPoolFunc func(name string, config Map) (pool WriterPool, err error)

func RegisterWriterPool(urn string, newFunc NewWriterPoolFunc) (err error) {
	writerPoolsLocker.Lock()
	writerPoolsLocker.Unlock()

	if urn == "" {
		panic("spirit: Register writer pool urn is empty")
	}

	if newFunc == nil {
		panic("spirit: Register writer pool is nil")
	}

	if _, exist := newWriterPoolFuncs[urn]; exist {
		panic("spirit: Register called twice for writer pool " + urn)
	}

	newWriterPoolFuncs[urn] = newFunc

	logger.WithField("module", "spirit").
		WithField("event", "register writer").
		WithField("urn", urn).
		Debugln("writer registered")

	return
}
