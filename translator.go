package spirit

import (
	"io"
	"sync"
)

var (
	inputTranslatorsLocker  = sync.Mutex{}
	newInputTranslatorFuncs = make(map[string]NewInputTranslatorFunc)
)

var (
	outputTranslatorsLocker  = sync.Mutex{}
	newOutputTranslatorFuncs = make(map[string]NewOutputTranslatorFunc)
)

type NewInputTranslatorFunc func(config Map) (translator InputTranslator, err error)
type NewOutputTranslatorFunc func(config Map) (translator OutputTranslator, err error)

type InputTranslator interface {
	In(r io.Reader) (delivery []Delivery, err error)
}

type OutputTranslator interface {
	Out(w io.Writer, delivery Delivery) (err error)
}

func RegisterInputTranslator(urn string, newFunc NewInputTranslatorFunc) (err error) {
	inputTranslatorsLocker.Lock()
	inputTranslatorsLocker.Unlock()

	if urn == "" {
		panic("spirit: Register input translator urn is empty")
	}

	if newFunc == nil {
		panic("spirit: Register input translator is nil")
	}

	if _, exist := newInputTranslatorFuncs[urn]; exist {
		panic("spirit: Register called twice for input translator " + urn)
	}

	newInputTranslatorFuncs[urn] = newFunc

	logger.WithField("module", "spirit").
		WithField("event", "register input translator").
		WithField("urn", urn).
		Debugln("input translator registered")

	return
}

func RegisterOutputTranslator(urn string, newFunc NewOutputTranslatorFunc) (err error) {
	outputTranslatorsLocker.Lock()
	outputTranslatorsLocker.Unlock()

	if urn == "" {
		panic("spirit: Register output translator urn is empty")
	}

	if newFunc == nil {
		panic("spirit: Register output translator is nil")
	}

	if _, exist := newOutputTranslatorFuncs[urn]; exist {
		panic("spirit: Register called twice for output translator " + urn)
	}

	newOutputTranslatorFuncs[urn] = newFunc

	logger.WithField("module", "spirit").
		WithField("event", "register output translator").
		WithField("urn", urn).
		Debugln("translator registered")

	return
}
