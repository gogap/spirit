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

type _Deliveries struct {
	Deliveries []spirit.Delivery
	Error      error
}

type ClassicWriterPoolConfig struct {
	Size int `json:"size"`
}

type ClassicWriterPool struct {
	statusLocker sync.Mutex
	conf         ClassicWriterPoolConfig
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
		conf: conf,
	}

	return
}

func (p *ClassicWriterPool) SetNewWriterFunc(newFunc spirit.NewWriterFunc) (err error) {
	return
}

func (p *ClassicWriterPool) Get(delivery spirit.Delivery) (writer io.WriteCloser, err error) {
	return
}

func (p *ClassicWriterPool) Put(delivery spirit.Delivery, writer io.WriteCloser) {
	return
}

func (p *ClassicWriterPool) Close() {
	return
}
