package std

import (
	"errors"
	"io"
	"sync"

	"github.com/gogap/spirit"
)

const (
	stdWriterURN = "urn:spirit:io:writer:std"
)

var (
	ErrSTDWriterDidNotBindProcess = errors.New("std writer did not bind process")
)

var _ io.WriteCloser = new(Stdin)

type Stdin struct {
	name string

	onceInit sync.Once
	conf     StdIOConfig

	proc *STDProcess
}

func init() {
	spirit.RegisterWriter(stdWriterURN, NewStdin)
}

func NewStdin(name string, options spirit.Map) (w io.WriteCloser, err error) {
	conf := StdIOConfig{}
	if err = options.ToObject(&conf); err != nil {
		return
	}

	if proc, e := takeSTDIO(_Output, conf); e != nil {
		err = e
	} else {
		w = &Stdin{
			name: name,
			conf: conf,
			proc: proc,
		}
		proc.Start()
	}

	return
}

func (p *Stdin) Name() string {
	return p.name
}

func (p *Stdin) URN() string {
	return stdWriterURN
}

func (p *Stdin) Write(data []byte) (n int, err error) {
	if p.proc == nil {
		err = ErrSTDWriterDidNotBindProcess
		return
	}
	return p.proc.Write(data)
}

func (p *Stdin) Close() (err error) {
	if p.proc == nil {
		return
	}

	p.proc.Stop()

	return
}
