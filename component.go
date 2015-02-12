package spirit

type ComponentHandler func(payload *Payload) (result interface{}, err error)

type Component interface {
	Name() string

	BindHandler(inPortName string, handlerName string) Component
	RegisterHandler(name string, handler ComponentHandler) Component
	CallHandler(handlerName string, payload *Payload) (result interface{}, err error)

	BindReceiver(inPortName string, receivers ...MessageReceiver) Component
	RegisterSender(senders ...MessageSender) Component

	Build() Component
	Run()
}
