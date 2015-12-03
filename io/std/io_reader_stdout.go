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
var _ spirit.Actor = new(Stdout)

type Stdout struct {
	name string

	onceInit sync.Once
	conf     StdIOConfig
	delim    string

	proc *STDProcess
}

func init() {
	spirit.RegisterReader(stdReaderURN, NewStdout)
}

func NewStdout(name string, options spirit.Map) (w io.ReadCloser, err error) {
	conf := StdIOConfig{}
	options.ToObject(&conf)

	if proc, e := takeSTDIO(_Input, conf); e != nil {
		err = e
	} else {
		w = &Stdout{
			name: name,
			conf: conf,
			proc: proc,
		}
		proc.Start()
	}

	return
}

func (p *Stdout) Name() string {
	return p.name
}

func (p *Stdout) URN() string {
	return stdReaderURN
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
