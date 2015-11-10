package cmd

import (
	"github.com/gogap/spirit"
)

const (
	consoleURN = "urn:spirit:console:cmd"
)

var _ spirit.Console = new(CMDConsole)

type CMDConsoleConfig struct {
}

type CMDConsole struct {
	conf CMDConsoleConfig
}

func init() {
	spirit.RegisterConsole(consoleURN, NewCMDConsole)
}

func NewCMDConsole(options spirit.Options) (console spirit.Console, err error) {
	conf := CMDConsoleConfig{}

	if err = options.ToObject(&conf); err != nil {
		return
	}

	console = &CMDConsole{conf: conf}

	return
}

func (p *CMDConsole) Start() (err error) {
	return
}

func (p *CMDConsole) Stop() (err error) {
	return
}

func (p *CMDConsole) AddSpirit(spit spirit.Spirit) (err error) {
	return
}
