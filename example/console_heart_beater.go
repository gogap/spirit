package main

import (
	"fmt"
	"github.com/gogap/spirit"
)

type ConsoleHeartBeater struct {
}

func (p *ConsoleHeartBeater) Name() string {
	return "console heart beater"
}
func (p *ConsoleHeartBeater) HeartBeat(heartBeatMessage spirit.HeartBeatMessage) {
	fmt.Println("heart beat message", heartBeatMessage)
}
