package spirit

type ReceiverMetadata struct {
	ComponentName string
	PortName      string
	Type          string
}

type MessageAddress struct {
	Type string `json:"type"`
	Url  string `json:"url"`
}

type OnReceiverMessageProcessed func(messageId string)
type OnReceiverMessageReceived func(inPortName, messageId string, compMsg ComponentMessage, callbacks ...OnReceiverMessageProcessed)
type OnReceiverError func(inPortName string, err error)

type MessageReceiver interface {
	Type() string
	Init(url string, options Options) error
	Address() MessageAddress
	BindInPort(
		componentName, inPortName string,
		onMsgReceived OnReceiverMessageReceived,
		onReceiverError OnReceiverError)
	Start()
	Stop()
	IsRunning() bool
	Metadata() ReceiverMetadata
}

type MessageSender interface {
	Type() string
	Init() error
	Send(url string, message ComponentMessage) (err error)
}
