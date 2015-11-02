package pool

import (
	"io"
	"sync"

	"github.com/gogap/spirit"
)

const (
	writerPoolURN = "urn:spirit:io:pool:writer:classic"
)

var _ spirit.WriterPool = new(ClassicWriterPool)

type ClassicWriterPoolConfig struct {
}

type writerInPool struct {
	InUse  bool
	Writer io.WriteCloser
}

type ClassicWriterPool struct {
	statusLocker sync.Mutex
	conf         ClassicWriterPoolConfig

	newWriterFunc spirit.NewWriterFunc
	writerOptions spirit.Options

	sessionWriters map[string]*writerInPool
	writerLocker   sync.Mutex

	isClosed bool
}

func init() {
	spirit.RegisterWriterPool(writerPoolURN, NewClassicWriterPool)
}

func NewClassicWriterPool(options spirit.Options) (pool spirit.WriterPool, err error) {
	conf := ClassicWriterPoolConfig{}
	if err = options.ToObject(&conf); err != nil {
		return
	}

	pool = &ClassicWriterPool{
		conf:           conf,
		sessionWriters: make(map[string]*writerInPool),
	}

	return
}

func (p *ClassicWriterPool) SetNewWriterFunc(newFunc spirit.NewWriterFunc, options spirit.Options) (err error) {
	p.newWriterFunc = newFunc
	p.writerOptions = options
	return
}

func (p *ClassicWriterPool) Get(delivery spirit.Delivery) (writer io.WriteCloser, err error) {
	if p.isClosed == true {
		err = spirit.ErrWriterPoolAlreadyClosed
		return
	}

	p.writerLocker.Lock()
	defer p.writerLocker.Unlock()

	if sessionWriter, exist := p.sessionWriters[delivery.SessionId()]; exist {
		if sessionWriter.InUse {
			err = spirit.ErrWriterInUse
			return
		} else {
			writer = sessionWriter.Writer
			sessionWriter.InUse = true

			spirit.Logger().WithField("actor", "writer pool").
				WithField("urn", writerPoolURN).
				WithField("event", "get writer").
				WithField("session_id", delivery.SessionId()).
				Debugln("get an exist writer")

			return
		}
	} else {
		if writer, err = p.newWriterFunc(p.writerOptions); err != nil {
			return
		} else {
			p.sessionWriters[delivery.SessionId()] = &writerInPool{InUse: true, Writer: writer}

			spirit.Logger().WithField("actor", "writer pool").
				WithField("urn", writerPoolURN).
				WithField("event", "get writer").
				WithField("session_id", delivery.SessionId()).
				Debugln("get a new writer")
		}
	}

	return
}

func (p *ClassicWriterPool) Put(delivery spirit.Delivery, writer io.WriteCloser) (err error) {
	if p.isClosed == true {
		err = spirit.ErrWriterPoolAlreadyClosed
		return
	}

	p.writerLocker.Lock()
	defer p.writerLocker.Unlock()

	if sessionWriter, exist := p.sessionWriters[delivery.SessionId()]; exist {
		if sessionWriter.InUse {
			sessionWriter.InUse = false
		}

		if sessionWriter.Writer != writer {
			spirit.Logger().WithField("actor", "writer pool").
				WithField("urn", writerPoolURN).
				WithField("event", "put writer").
				WithField("session_id", delivery.SessionId()).
				Debugln("put writer")
			sessionWriter.Writer = writer
		}
	} else {
		p.sessionWriters[delivery.SessionId()] = &writerInPool{InUse: false, Writer: writer}
	}

	return
}

func (p *ClassicWriterPool) Close() {
	p.writerLocker.Lock()
	defer p.writerLocker.Unlock()

	p.statusLocker.Lock()
	defer p.statusLocker.Unlock()

	for sid, sessionWriter := range p.sessionWriters {
		sessionWriter.Writer.Close()
		delete(p.sessionWriters, sid)

		spirit.Logger().WithField("actor", "writer pool").
			WithField("urn", writerPoolURN).
			WithField("event", "close writer").
			WithField("session_id", sid).
			Debugln("writer closed and deleted")
	}

	p.isClosed = true

	return
}
