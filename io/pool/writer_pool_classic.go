package pool

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/gogap/spirit"
)

const (
	writerPoolURN = "urn:spirit:io:pool:writer:classic"
)

var (
	DefaultWriterPoolSize = 10
)

var _ spirit.WriterPool = new(ClassicWriterPool)

type ClassicWriterPoolConfig struct {
	EnableSession bool `json:"enable_session"`
	MaxSize       int  `json:"max_size"`
	Idle          int  `json:"idle"`
}

type writerInPool struct {
	InUse  bool
	Writer io.WriteCloser
}

type ClassicWriterPool struct {
	name string

	statusLocker sync.Mutex
	conf         ClassicWriterPoolConfig

	newWriterFunc spirit.NewWriterFunc
	writerConfig  spirit.Map

	sessionWriters map[string]*writerInPool
	writerLocker   sync.Mutex

	writers    []io.WriteCloser
	writersMap map[io.WriteCloser]bool

	writerId int64

	isClosed bool
}

func init() {
	spirit.RegisterWriterPool(writerPoolURN, NewClassicWriterPool)
}

func NewClassicWriterPool(name string, options spirit.Map) (pool spirit.WriterPool, err error) {
	conf := ClassicWriterPoolConfig{}
	if err = options.ToObject(&conf); err != nil {
		return
	}

	if conf.MaxSize <= 0 {
		conf.MaxSize = DefaultWriterPoolSize
	}

	pool = &ClassicWriterPool{
		name:           name,
		conf:           conf,
		sessionWriters: make(map[string]*writerInPool),
		writersMap:     make(map[io.WriteCloser]bool),
	}

	return
}

func (p *ClassicWriterPool) Name() string {
	return p.name
}

func (p *ClassicWriterPool) URN() string {
	return writerPoolURN
}

func (p *ClassicWriterPool) SetNewWriterFunc(newFunc spirit.NewWriterFunc, config spirit.Map) (err error) {
	p.newWriterFunc = newFunc
	p.writerConfig = config
	return
}

func (p *ClassicWriterPool) getNonSessionWriter(delivery spirit.Delivery) (writer io.WriteCloser, err error) {

	if len(p.writers) > 0 {
		writer = p.writers[0]
		if len(p.writers) > 1 {
			p.writers = p.writers[1:]
		} else {
			p.writers = []io.WriteCloser{}
		}
		delete(p.writersMap, writer)

		spirit.Logger().WithField("actor", "writer pool").
			WithField("urn", writerPoolURN).
			WithField("name", p.name).
			WithField("event", "get writer").
			Debugln("get an old non-session writer")

		return
	} else if len(p.writers) < p.conf.MaxSize {
		writerId := atomic.AddInt64(&p.writerId, 1)
		writerName := fmt.Sprintf("%s_%d", p.name, writerId)
		if writer, err = p.newWriterFunc(writerName, p.writerConfig); err != nil {
			return
		} else {
			p.writers = append(p.writers, writer)
			p.writersMap[writer] = true
			spirit.Logger().WithField("actor", "writer pool").
				WithField("urn", writerPoolURN).
				WithField("name", p.name).
				WithField("event", "get writer").
				Debugln("get a new non-session writer")
		}
	} else {
		err = spirit.ErrWriterPoolTooManyWriters
		return
	}

	return
}

func (p *ClassicWriterPool) getSessionWriter(delivery spirit.Delivery) (writer io.WriteCloser, err error) {
	if sessionWriter, exist := p.sessionWriters[delivery.SessionId()]; exist {
		if sessionWriter.InUse {
			err = spirit.ErrWriterInUse
			return
		} else {
			writer = sessionWriter.Writer
			sessionWriter.InUse = true

			spirit.Logger().WithField("actor", "writer pool").
				WithField("urn", writerPoolURN).
				WithField("name", p.name).
				WithField("event", "get writer").
				WithField("session_id", delivery.SessionId()).
				Debugln("get an exist session writer")

			return
		}
	} else {

		writerId := atomic.AddInt64(&p.writerId, 1)
		writerName := fmt.Sprintf("%s_%d", p.name, writerId)

		if writer, err = p.newWriterFunc(writerName, p.writerConfig); err != nil {
			return
		} else {
			p.sessionWriters[delivery.SessionId()] = &writerInPool{InUse: true, Writer: writer}

			spirit.Logger().WithField("actor", "writer pool").
				WithField("urn", writerPoolURN).
				WithField("name", p.name).
				WithField("event", "get writer").
				WithField("session_id", delivery.SessionId()).
				Debugln("get a new session writer")
		}
	}
	return
}

func (p *ClassicWriterPool) Get(delivery spirit.Delivery) (writer io.WriteCloser, err error) {
	if p.isClosed == true {
		err = spirit.ErrWriterPoolAlreadyClosed
		return
	}

	p.writerLocker.Lock()
	defer p.writerLocker.Unlock()

	if p.conf.EnableSession {
		return p.getSessionWriter(delivery)
	} else {
		return p.getNonSessionWriter(delivery)
	}

	return
}

func (p *ClassicWriterPool) putSessionWriter(delivery spirit.Delivery, writer io.WriteCloser) (err error) {
	if sessionWriter, exist := p.sessionWriters[delivery.SessionId()]; exist {
		if sessionWriter.InUse {
			sessionWriter.InUse = false
		}

		if sessionWriter.Writer != writer {
			spirit.Logger().WithField("actor", "writer pool").
				WithField("urn", writerPoolURN).
				WithField("name", p.name).
				WithField("event", "put writer").
				WithField("session_id", delivery.SessionId()).
				Debugln("put session writer")
			sessionWriter.Writer = writer
		}
	} else {
		p.sessionWriters[delivery.SessionId()] = &writerInPool{InUse: false, Writer: writer}
	}

	return
}

func (p *ClassicWriterPool) putNonSessionWriter(delivery spirit.Delivery, writer io.WriteCloser) (err error) {
	if writer == nil {
		err = spirit.ErrWriterIsNil
		return
	}

	if p.isClosed == true {
		err = spirit.ErrWriterPoolAlreadyClosed
		return
	}

	if _, exist := p.writersMap[writer]; exist {

		spirit.Logger().WithField("actor", "writer pool").
			WithField("urn", writerPoolURN).
			WithField("name", p.name).
			WithField("event", "put writer").
			Debugln("put an already exist writer")

		return
	}

	p.writersMap[writer] = true
	p.writers = append(p.writers, writer)

	spirit.Logger().WithField("actor", "writer pool").
		WithField("urn", writerPoolURN).
		WithField("name", p.name).
		WithField("event", "put writer").
		Debugln("put non-session writer")

	return
}

func (p *ClassicWriterPool) Put(delivery spirit.Delivery, writer io.WriteCloser) (err error) {
	if p.isClosed == true {
		err = spirit.ErrWriterPoolAlreadyClosed
		return
	}

	if p.conf.Idle < 0 {
		return
	}

	p.writerLocker.Lock()
	defer p.writerLocker.Unlock()

	if p.conf.EnableSession {
		err = p.putSessionWriter(delivery, writer)
	} else {
		err = p.putNonSessionWriter(delivery, writer)
	}

	return
}

func (p *ClassicWriterPool) Close() {
	p.writerLocker.Lock()
	defer p.writerLocker.Unlock()

	p.statusLocker.Lock()
	defer p.statusLocker.Unlock()

	if p.conf.EnableSession {
		for sid, sessionWriter := range p.sessionWriters {
			sessionWriter.Writer.Close()
			delete(p.sessionWriters, sid)

			spirit.Logger().WithField("actor", "writer pool").
				WithField("urn", writerPoolURN).
				WithField("name", p.name).
				WithField("event", "close writer").
				WithField("session_id", sid).
				Debugln("writer closed and deleted")
		}
	} else {
		for _, writer := range p.writers {
			writer.Close()
			delete(p.writersMap, writer)

			spirit.Logger().WithField("actor", "writer pool").
				WithField("urn", writerPoolURN).
				WithField("name", p.name).
				WithField("event", "close writer").
				Debugln("writer closed and deleted")
		}
	}

	p.isClosed = true

	return
}
