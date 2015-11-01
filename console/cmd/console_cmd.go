package cmd

import (
	"github.com/gogap/spirit"
)

const (
	consoleURN = "urn:spirit:console:cmd"
)

var _ spirit.Console = new(CMDConsole)

type CMDConsole struct {
}

func init() {
	spirit.RegisterConsole(consoleURN, NewCMDConsole)
}

func NewCMDConsole(options spirit.Options) (console spirit.Console, err error) {
	return
}

func (p *CMDConsole) Start() (err error) {
	return
}

func (p *CMDConsole) Stop() (err error) {
	return
}
