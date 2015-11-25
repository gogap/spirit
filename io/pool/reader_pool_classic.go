package pool

import (
	"io"
	"sync"

	"github.com/gogap/spirit"
)

const (
	readerPoolURN = "urn:spirit:io:pool:reader:classic"
)

var (
	DefaultReaderPoolSize = 10
)

var _ spirit.ReaderPool = new(ClassicReaderPool)

type ClassicReaderPoolConfig struct {
	MaxSize int `json:"max_size"`
}

type ClassicReaderPool struct {
	statusLocker sync.Mutex
	conf         ClassicReaderPoolConfig

	newReaderFunc spirit.NewReaderFunc
	readerConfig  spirit.Map

	readers      []io.ReadCloser
	readerLocker sync.Mutex

	readersMap map[io.ReadCloser]bool

	isClosed bool
}

func init() {
	spirit.RegisterReaderPool(readerPoolURN, NewClassicReaderPool)
}

func NewClassicReaderPool(config spirit.Map) (pool spirit.ReaderPool, err error) {
	conf := ClassicReaderPoolConfig{}
	if err = config.ToObject(&conf); err != nil {
		return
	}

	if conf.MaxSize <= 0 {
		conf.MaxSize = DefaultReaderPoolSize
	}

	pool = &ClassicReaderPool{
		conf:       conf,
		readersMap: make(map[io.ReadCloser]bool),
	}

	return
}

func (p *ClassicReaderPool) SetNewReaderFunc(newFunc spirit.NewReaderFunc, config spirit.Map) (err error) {
	p.newReaderFunc = newFunc
	p.readerConfig = config
	return
}

func (p *ClassicReaderPool) Get() (reader io.ReadCloser, err error) {
	if p.isClosed == true {
		err = spirit.ErrReaderPoolAlreadyClosed
		return
	}

	p.readerLocker.Lock()
	defer p.readerLocker.Unlock()

	if len(p.readers) > 0 {
		reader = p.readers[0]
		if len(p.readers) > 1 {
			p.readers = p.readers[1:]
		} else {
			p.readers = []io.ReadCloser{}
		}
		delete(p.readersMap, reader)

		spirit.Logger().WithField("actor", "reader pool").
			WithField("urn", readerPoolURN).
			WithField("event", "get reader").
			Debugln("get an old reader")

		return
	} else if len(p.readers) < p.conf.MaxSize {
		if reader, err = p.newReaderFunc(p.readerConfig); err != nil {
			return
		} else {
			p.readers = append(p.readers, reader)
			p.readersMap[reader] = true
			spirit.Logger().WithField("actor", "reader pool").
				WithField("urn", readerPoolURN).
				WithField("event", "get reader").
				Debugln("get a new reader")
		}
	} else {
		err = spirit.ErrReaderPoolTooManyReaders
		return
	}

	return
}

func (p *ClassicReaderPool) Put(reader io.ReadCloser) (err error) {
	if reader == nil {
		err = spirit.ErrReaderIsNil
		return
	}

	if p.isClosed == true {
		err = spirit.ErrReaderPoolAlreadyClosed
		return
	}

	p.readerLocker.Lock()
	defer p.readerLocker.Unlock()

	if _, exist := p.readersMap[reader]; exist {

		spirit.Logger().WithField("actor", "reader pool").
			WithField("urn", readerPoolURN).
			WithField("event", "put reader").
			Debugln("put an already exist reader")

		return
	}

	p.readersMap[reader] = true
	p.readers = append(p.readers, reader)

	spirit.Logger().WithField("actor", "reader pool").
		WithField("urn", readerPoolURN).
		WithField("event", "put reader").
		Debugln("put reader")

	return
}

func (p *ClassicReaderPool) Close() {
	p.readerLocker.Lock()
	defer p.readerLocker.Unlock()

	p.statusLocker.Lock()
	defer p.statusLocker.Unlock()

	for _, reader := range p.readers {
		reader.Close()
		delete(p.readersMap, reader)

		spirit.Logger().WithField("actor", "reader pool").
			WithField("urn", readerPoolURN).
			WithField("event", "close reader").
			Debugln("reader closed and deleted")
	}

	p.readers = nil
	p.isClosed = true

	return
}