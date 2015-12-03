package spirit

import (
	"io"
	"sync"
)

var (
	writersLocker  = sync.Mutex{}
	newWriterFuncs = make(map[string]NewWriterFunc)
)

type NewWriterFunc func(name string, options Map) (w io.WriteCloser, err error)

func RegisterWriter(urn string, newFunc NewWriterFunc) (err error) {
	writersLocker.Lock()
	writersLocker.Unlock()

	if urn == "" {
		panic("spirit: Register writer urn is empty")
	}

	if newFunc == nil {
		panic("spirit: Register writer is nil")
	}

	if _, exist := newWriterFuncs[urn]; exist {
		panic("spirit: Register called twice for writer " + urn)
	}

	newWriterFuncs[urn] = newFunc

	logger.WithField("module", "spirit").
		WithField("event", "register writer").
		WithField("urn", urn).
		Debugln("writer registered")

	return
}
