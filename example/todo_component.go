package main

import (
	"fmt"
	"time"

	"github.com/gogap/spirit"
)

type Todo struct {
	User       string    `json:"user"`
	Task       string    `json:"task"`
	IsDone     bool      `json:"is_done"`
	CreateTime time.Time `json:"create_time"`
}

func (p *Todo) NewTask(payload spirit.Payload) (result interface{}, err error) {
	reqTodo := Todo{}
	payload.FillContentToObject(&reqTodo)

	result = Todo{User: reqTodo.User, Task: "hello spirit task", IsDone: false, CreateTime: time.Now()}
	return
}

func (p *Todo) DeleteTask(payload spirit.Payload) (result interface{}, err error) {
	err = fmt.Errorf("task of %s not exist", "hello")
	return
}

func (p *Todo) DoneTask(payload spirit.Payload) (result interface{}, err error) {
	result = Todo{User: "gogap", Task: "hello spirit task", IsDone: true, CreateTime: time.Now()}
	return
}
