package spirit

import (
	"fmt"
)

const (
	SPIRIT = "spirit"
)

const (
	SPIRIT_CHILD_INSTANCE = "SPIRIT_CHILD_INSTANCE"
	SPIRIT_BIN_HOME       = "SPIRIT_BIN_HOME"
)

var (
	Assets *assetsManager = newAssetsManager("")
)

var (
	instManager *instanceManager = nil
)

type Spirit interface {
	Hosting(component Component, initalFuncs ...InitalFunc) Spirit
	Build() Spirit
	Run()
}

type InitalFunc func() (err error)

var (
	receiverFactory MessageReceiverFactory = NewDefaultMessageReceiverFactory()
	senderFactory   MessageSenderFactory   = NewDefaultMessageSenderFactory()
	hookFactory     MessageHookFactory     = NewDefaultMessageHookFactory()

	heartbeaters map[string]Heartbeater = make(map[string]Heartbeater)
)

func init() {
	receiverFactory.RegisterMessageReceivers(new(MessageReceiverMNS))
	receiverFactory.RegisterMessageReceivers(new(MessageReceiverMQS)) //for compatible

	senderFactory.RegisterMessageSenders(new(MessageSenderMNS))
	senderFactory.RegisterMessageSenders(new(MessageSenderMQS)) //for compatible

	hookFactory.RegisterMessageHooks(new(MessageHookBigDataRedis))

	RegisterHeartbeaters(new(AliJiankong))
}

func RegisterMessageReceivers(receivers ...MessageReceiver) {
	receiverFactory.RegisterMessageReceivers(receivers...)
}

func RegisterMessageSenders(senders ...MessageSender) {
	senderFactory.RegisterMessageSenders(senders...)
}

func RegisterMessageHooks(hooks ...MessageHook) {
	hookFactory.RegisterMessageHooks(hooks...)
}

func SetMessageReceiverFactory(factory MessageReceiverFactory) {
	if factory == nil {
		panic("message receiver factory could not be nil")
	}
	receiverFactory = factory
}

func GetMessageReceiverFactory() MessageReceiverFactory {
	return receiverFactory
}

func SetMessageSenderFactory(factory MessageSenderFactory) {
	if factory == nil {
		panic("message sender factory could not be nil")
	}
	senderFactory = factory
}

func GetMessageSenderFactory() MessageSenderFactory {
	return senderFactory
}

func SetMessageHookFactory(factory MessageHookFactory) {
	if factory == nil {
		panic("message hook factory could not be nil")
	}
	hookFactory = factory
}

func GetMessageHookFactory() MessageHookFactory {
	return hookFactory
}

func RegisterHeartbeaters(beaters ...Heartbeater) {
	if beaters == nil || len(beaters) == 0 {
		return
	}

	for _, beater := range beaters {
		if _, exist := heartbeaters[beater.Name()]; exist {
			panic(fmt.Sprintf("heart beater %s already exist", beater.Name()))
		}
		heartbeaters[beater.Name()] = beater
	}
	return
}

func RemoveHeartBeaters(names ...string) {
	if names == nil || len(names) == 0 {
		return
	}

	if heartbeaters == nil {
		return
	}

	for _, name := range names {
		delete(heartbeaters, name)
	}

	return
}
