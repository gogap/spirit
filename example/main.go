package main

import (
	"github.com/gogap/spirit"
)

func main() {
	todoSpirit := spirit.NewClassicSpirit("example", "a example of todo component", "1.0.0")

	todoComponent := spirit.NewBaseComponent("todo")

	todo := new(Todo)

	todoComponent.RegisterHandler("new_task", todo.NewTask)
	todoComponent.RegisterHandler("delete_task", todo.DeleteTask)
	todoComponent.RegisterHandler("done_task", todo.DoneTask)

	todoSpirit.RegisterHeartbeaters(new(ConsoleHeartbeater))

	todoSpirit.Hosting(todoComponent).Build().Run()
}
