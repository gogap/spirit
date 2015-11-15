package std

import (
	"io"
	"sync"

	"github.com/gogap/spirit"
)

const (
	stdReaderURN = "urn:spirit:io:reader:std"
)

var _ io.ReadCloser = new(Stdout)

type Stdout struct {
	onceInit sync.Once
	conf     StdIOConfig
	delim    string

	proc *STDProcess
}

func init() {
	spirit.RegisterReader(stdReaderURN, NewStdout)
}

func NewStdout(config spirit.Config) (w io.ReadCloser, err error) {
	conf := StdIOConfig{}
	config.ToObject(&conf)

	if proc, e := takeSTDIO(_Input, conf); e != nil {
		err = e
	} else {
		w = &Stdout{
			conf: conf,
			proc: proc,
		}
		proc.Start()
	}

	return
}

func (p *Stdout) Read(data []byte) (n int, err error) {
	if p.proc == nil || err != nil {
		return
	}

	return p.proc.Read(data)
}

func (p *Stdout) Close() (err error) {
	if p.proc == nil {
		return
	}

	p.proc.Stop()

	return
}
