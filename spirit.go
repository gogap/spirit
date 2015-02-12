package spirit

type Spirit interface {
	SetMessageReceiverFactory(factory MessageReceiverFactory)
	GetMessageReceiverFactory() MessageReceiverFactory

	SetMessageSenderFactory(factory MessageSenderFactory)
	GetMessageSenderFactory() MessageSenderFactory

	Hosting(components ...Component) Spirit
	Run()
}
