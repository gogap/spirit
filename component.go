package spirit

type ComponentStatus int

const (
	SIG_PAUSE  = 1000
	SIG_RESUME = 1001
	SIG_STOP   = 1002
)

const (
	STATUS_READY    ComponentStatus = 0
	STATUS_RUNNING                  = 1
	STATUS_PAUSED                   = 2
	STATUS_STOPPING                 = 3
	STATUS_STOPPED                  = 4
)

type ComponentHandler func(payload *Payload) (result interface{}, err error)

type Component interface {
	Name() string

	BindHandler(inPortName string, handlerName string) Component
	RegisterHandler(name string, handler ComponentHandler) Component
	CallHandler(handlerName string, payload *Payload) (result interface{}, err error)

	AddInPortHooks(inportName string, hooks ...string) Component
	ClearInPortHooks(inportName string) (err error)

	ListHandlers() (handlers map[string]ComponentHandler, err error)
	GetHandlers(handlerNames ...string) (handlers map[string]ComponentHandler, err error)

	BindReceiver(inPortName string, receivers ...MessageReceiver) Component
	GetReceivers(inPortName string) []MessageReceiver

	Build() Component
	Run()

	PauseOrResume()
	Stop()

	Status() ComponentStatus
}
