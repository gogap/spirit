package main

import (
	"github.com/gogap/spirit"
)

func main() {
	todoSpirit := spirit.NewClassicSpirit("todo", "a todo component", "1.0.0")

	todoComponent := spirit.NewBaseComponent("todo")

	todo := new(Todo)

	todoComponent.BindHandler("port.new", todo.NewTask)
	todoComponent.BindHandler("port.delete", todo.DeleteTask)
	todoComponent.BindHandler("port.done", todo.DoneTask)

	todoSpirit.Hosting(todoComponent)

	todoSpirit.Run()
}
