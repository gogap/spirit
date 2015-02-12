package spirit

type ComponentHandler func(payload *Payload) (result interface{}, err error)

type Component interface {
	Name() string
	BindHandler(inPortName string, handler ComponentHandler) Component
	BindReceiver(inPortName string, receivers ...MessageReceiver) Component
	RegisterSender(senders ...MessageSender) Component
	CallHandler(inPortName string, payload *Payload) (result interface{}, err error)
	Build() Component
	Run()
}
