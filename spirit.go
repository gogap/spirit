package spirit

type InitalFunc func(configFile string) (err error)

type Spirit interface {
	SetMessageReceiverFactory(factory MessageReceiverFactory)
	GetMessageReceiverFactory() MessageReceiverFactory

	SetMessageSenderFactory(factory MessageSenderFactory)
	GetMessageSenderFactory() MessageSenderFactory

	SetMessageHookFactory(factory MessageHookFactory)
	GetMessageHookFactory() MessageHookFactory

	GetComponent(name string) Component

	RegisterHeartbeaters(beaters ...Heartbeater) Spirit
	RemoveHeartBeaters(names ...string) Spirit

	Hosting(components ...Component) Spirit
	Build() Spirit
	Run(initalFuncs ...InitalFunc)
}
