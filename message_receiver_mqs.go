package spirit

type MessageReceiverMQS struct {
	MessageReceiverMNS
}

func (p *MessageReceiverMQS) Type() string {
	return "mqs"
}
