package spirit

type Spirit interface {
	SetMessageReceiverFactory(factory MessageReceiverFactory)
	GetMessageReceiverFactory() MessageReceiverFactory

	SetMessageSenderFactory(factory MessageSenderFactory)
	GetMessageSenderFactory() MessageSenderFactory

	GetComponent(name string) Component

	Hosting(components ...Component) Spirit
	Build() Spirit
	Run()
}
