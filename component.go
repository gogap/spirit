package spirit

type ComponentStatus int

const (
	STATUS_READY    ComponentStatus = 0
	STATUS_RUNNING  ComponentStatus = 1
	STATUS_PAUSED   ComponentStatus = 2
	STATUS_STOPPING ComponentStatus = 3
	STATUS_STOPED   ComponentStatus = 4
)

type ComponentHandler func(payload *Payload) (result interface{}, err error)

type Component interface {
	Name() string

	BindHandler(inPortName string, handlerName string) Component
	RegisterHandler(name string, handler ComponentHandler) Component
	CallHandler(handlerName string, payload *Payload) (result interface{}, err error)

	ListHandlers() (handlers map[string]ComponentHandler, err error)
	GetHandlers(handlerNames ...string) (handlers map[string]ComponentHandler, err error)

	BindReceiver(inPortName string, receivers ...MessageReceiver) Component
	GetReceivers(inPortName string) []MessageReceiver

	SetMessageSenderFactory(factory MessageSenderFactory) Component

	Build() Component
	Run()

	PauseOrResume()
	Stop()

	Status() ComponentStatus
}
