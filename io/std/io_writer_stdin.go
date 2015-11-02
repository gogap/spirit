package std

import (
	"io"
	"sync"

	"github.com/gogap/spirit"
)

const (
	stdWriterURN = "urn:spirit:io:writer:std"
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

func NewStdin(options spirit.Options) (w io.WriteCloser, err error) {
	conf := StdIOConfig{}
	if err = options.ToObject(&conf); err != nil {
		return
	}

	w = &Stdin{
		conf: conf,
	}
	return
}

func (p *Stdin) Write(data []byte) (n int, err error) {
	if p.proc == nil {
		p.onceInit.Do(func() {
			if proc, e := takeSTDIO(_Output, p.conf); e != nil {
				err = e
			} else {
				p.proc = proc
				p.proc.Start()
			}
		})
	}

	if p.proc == nil || err != nil {
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
