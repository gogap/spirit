package main

import (
	"fmt"
	"github.com/gogap/spirit"
)

type ConsoleHeartbeater struct {
}

func (p *ConsoleHeartbeater) Start(options spirit.Options) error {
	return nil
}

func (p *ConsoleHeartbeater) Name() string {
	return "console_heart_beater"
}
func (p *ConsoleHeartbeater) Heartbeat(heartbeatMessage spirit.HeartbeatMessage) {
	fmt.Println("heart beat message", heartbeatMessage)
}
