package cmd

import (
	"github.com/gogap/spirit"
)

const (
	consoleURN = "urn:spirit:console:cmd"
)

var _ spirit.Console = new(CMDConsole)
var _ spirit.Actor = new(CMDConsole)

type CMDConsoleConfig struct {
}

type CMDConsole struct {
	name string
	conf CMDConsoleConfig
}

func init() {
	spirit.RegisterConsole(consoleURN, NewCMDConsole)
}

func NewCMDConsole(name string, config spirit.Map) (console spirit.Console, err error) {
	conf := CMDConsoleConfig{}

	if err = config.ToObject(&conf); err != nil {
		return
	}

	console = &CMDConsole{name: name, conf: conf}

	return
}

func (p *CMDConsole) Name() string {
	return p.name
}

func (p *CMDConsole) URN() string {
	return consoleURN
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
