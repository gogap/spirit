package spirit

import (
	"sync"
)

type LabelMatcher interface {
	Match(la Labels, lb Labels) (matched bool)
}

var (
	labelMatchersLocker  = sync.Mutex{}
	newLabelMatcherFuncs = make(map[string]NewLabelMatcherFunc)
)

type NewLabelMatcherFunc func(options Options) (matcher LabelMatcher, err error)

func RegisterLabelMatcher(urn string, newFunc NewLabelMatcherFunc) (err error) {
	labelMatchersLocker.Lock()
	labelMatchersLocker.Unlock()

	if urn == "" {
		panic("spirit: Register label matcher urn is empty")
	}

	if newFunc == nil {
		panic("spirit: Register label matcher is nil")
	}

	if _, exist := newLabelMatcherFuncs[urn]; exist {
		panic("spirit: Register called twice for label matcher " + urn)
	}

	newLabelMatcherFuncs[urn] = newFunc

	logger.WithField("module", "spirit").
		WithField("event", "register label matcher").
		WithField("urn", urn).
		Debugln("label matcher registered")

	return
}
