package spirit

type MessageAddress struct {
	Type        string `json:"type"`
	Url         string `json:"url"`
	IsDelivered bool   `json:"is_delivered"`
}

type MessageReceiver interface {
	Type() string
	Init(url, configFile string) error
	Receive(message chan ComponentMessage, err chan error)
}

type MessageSender interface {
	Type() string
	Send(url string, message ComponentMessage) (err error)
}
