package spirit

type MessageSenderMQS struct {
	MessageSenderMNS
}

func (p *MessageSenderMQS) Type() string {
	return "mqs"
}
