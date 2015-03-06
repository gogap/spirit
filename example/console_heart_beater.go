package main

import (
	"fmt"
	"github.com/gogap/spirit"
)

type ConsoleHeartbeater struct {
}

func (p *ConsoleHeartbeater) Start() error {
	return nil
}

func (p *ConsoleHeartbeater) Name() string {
	return "console heart beater"
}
func (p *ConsoleHeartbeater) Heartbeat(heartbeatMessage spirit.HeartbeatMessage) {
	fmt.Println("heart beat message", heartbeatMessage)
}
