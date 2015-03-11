package spirit

type InitalFunc func() (err error)

type Spirit interface {
	SetMessageReceiverFactory(factory MessageReceiverFactory)
	GetMessageReceiverFactory() MessageReceiverFactory

	SetMessageSenderFactory(factory MessageSenderFactory)
	GetMessageSenderFactory() MessageSenderFactory

	GetComponent(name string) Component

	RegisterHeartbeaters(beaters ...Heartbeater) Spirit
	RemoveHeartBeaters(names ...string) Spirit

	Hosting(components ...Component) Spirit
	Build() Spirit
	Run(initalFuncs ...InitalFunc)
}
