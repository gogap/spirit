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
	onceInit sync.Once
	conf     StdIOConfig

	proc *STDProcess
}

func init() {
	spirit.RegisterWriter(stdWriterURN, NewStdin)
}

func NewStdin(config spirit.Config) (w io.WriteCloser, err error) {
	conf := StdIOConfig{}
	if err = config.ToObject(&conf); err != nil {
		return
	}

	if proc, e := takeSTDIO(_Output, conf); e != nil {
		err = e
	} else {
		w = &Stdin{
			conf: conf,
			proc: proc,
		}
		proc.Start()
	}

	return
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
