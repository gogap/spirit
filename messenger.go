package spirit

type MessageAddress struct {
	Type string `json:"type"`
	Url  string `json:"url"`
}

type MessageReceiver interface {
	Type() string
	Init(url, configFile string) error
	Address() MessageAddress
	Receive(message chan ComponentMessage, err chan error)
}

type MessageSender interface {
	Type() string
	Init() error
	Send(url string, message ComponentMessage) (err error)
}
