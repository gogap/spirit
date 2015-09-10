package spirit

type ReceiverMetadata struct {
	ComponentName string
	PortName      string
	Url           string
	Type          string
}

type PortChan struct {
	ComponentName string
	PortName      string
	Message       chan ComponentMessage
	Error         chan error
}

type MessageAddress struct {
	Type string `json:"type"`
	Url  string `json:"url"`
}

type MessageReceiver interface {
	Type() string
	Init(url string, options Options) error
	Address() MessageAddress
	Receive(portChan *PortChan)
}

type MessageSender interface {
	Type() string
	Init() error
	Send(url string, message ComponentMessage) (err error)
}
