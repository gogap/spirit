package spirit

import (
	"sync"
)

type URNRewriter interface {
	Rewrite(delivery Delivery) (newURN string, err error)
}

var (
	rewritersLocker     = sync.Mutex{}
	newURNRewriterFuncs = make(map[string]NewURNRewriterFunc)
)

type NewURNRewriterFunc func(options Options) (rewriter URNRewriter, err error)

func RegisterURNRewriter(urn string, newFunc NewURNRewriterFunc) (err error) {
	rewritersLocker.Lock()
	rewritersLocker.Unlock()

	if urn == "" {
		panic("spirit: Register urn rewriter's urn is empty")
	}

	if newFunc == nil {
		panic("spirit: Register urn rewriter is nil")
	}

	if _, exist := newURNRewriterFuncs[urn]; exist {
		panic("spirit: Register called twice for urn rewriter " + urn)
	}

	newURNRewriterFuncs[urn] = newFunc

	logger.WithField("module", "spirit").
		WithField("event", "register urn rewriter").
		WithField("urn", urn).
		Debugln("urn rewriter registered")

	return
}
