package main

import (
	"fmt"
	"time"

	"github.com/gogap/spirit"
)

type Todo struct {
	Name       string    `json:"name"`
	IsDone     bool      `json:"is_done"`
	CreateTime time.Time `json:"create_time"`
}

func (p *Todo) NewTask(payload *spirit.Payload) (result interface{}, err error) {
	result = Todo{Name: "hello spirit task", IsDone: false, CreateTime: time.Now()}
	return
}

func (p *Todo) DeleteTask(payload *spirit.Payload) (result interface{}, err error) {
	err = fmt.Errorf("task of %s not exist", "hello")
	return
}

func (p *Todo) DoneTask(payload *spirit.Payload) (result interface{}, err error) {
	result = Todo{Name: "hello spirit task", IsDone: true, CreateTime: time.Now()}
	return
}
