package classic

import (
	"io"
	"sync"

	"github.com/gogap/spirit"
)

const (
	readerPoolURN = "urn:spirit:reader-pool:classic"
)

var _ spirit.ReaderPool = new(ClassicReaderPool)

type _Deliveries struct {
	Deliveries []spirit.Delivery
	Error      error
}

type ClassicReaderPoolConfig struct {
	Size int `json:"size"`
}

type ClassicReaderPool struct {
	statusLocker sync.Mutex
	conf         ClassicReaderPoolConfig
}

func init() {
	spirit.RegisterReaderPool(readerPoolURN, NewClassicReaderPool)
}

func NewClassicReaderPool(options spirit.Options) (pool spirit.ReaderPool, err error) {
	conf := ClassicReaderPoolConfig{}
	if err = options.ToObject(&conf); err != nil {
		return
	}

	pool = &ClassicReaderPool{
		conf: conf,
	}

	return
}

func (p *ClassicReaderPool) SetNewReaderFunc(newFunc spirit.NewReaderFunc) (err error) {
	return
}

func (p *ClassicReaderPool) Get() (reader io.ReadCloser, err error) {
	return
}

func (p *ClassicReaderPool) Put(reader io.ReadCloser) {
	return
}

func (p *ClassicReaderPool) Close() {
	return
}
