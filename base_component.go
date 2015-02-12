package spirit

import (
	"fmt"
	"sync"

	"github.com/gogap/errors"
	"github.com/gogap/logs"
)

type BaseComponent struct {
	name      string
	receivers map[string][]MessageReceiver
	senders   map[string]MessageSender
	handlers  map[string]ComponentHandler

	runtimeLocker sync.Mutex
	isBuilded     bool
	isRunning     bool

	messageChans map[string]chan ComponentMessage
	errChans     map[string]chan error
}

func NewBaseComponent(componentName string) *BaseComponent {
	if componentName == "" {
		panic("component name could not be empty")
	}

	return &BaseComponent{
		name:      componentName,
		receivers: make(map[string][]MessageReceiver),
		senders:   make(map[string]MessageSender),
		handlers:  make(map[string]ComponentHandler),

		messageChans: make(map[string]chan ComponentMessage),
		errChans:     make(map[string]chan error),
	}
}

func (p *BaseComponent) Name() string {
	return p.name
}

func (p *BaseComponent) CallHandler(inPortName string, payload *Payload) (result interface{}, err error) {
	if inPortName == "" {
		err = ERR_PORT_NAME_IS_EMPTY.New(errors.Params{"name": p.name})
		return
	}

	if handler, exist := p.handlers[inPortName]; exist {
		if ret, e := handler(payload); e != nil {
			err = ERR_COMPONENT_HANDLER_RETURN_ERROR.New(errors.Params{"err": e})
			return
		} else {
			result = ret
			return
		}
	} else {
		err = ERR_COMPONENT_HANDLER_NOT_EXIST.New(errors.Params{"name": p.name, "portName": inPortName})
		return
	}
}

func (p *BaseComponent) BindHandler(inPortName string, handler ComponentHandler) Component {
	if inPortName == "" {
		panic(fmt.Sprintf("[component-%s] in port name could not be empty", p.name))
	}

	if handler == nil {
		panic(fmt.Sprintf("[component-%s] handler could not be nil, in port name: %s", p.name, inPortName))
	}

	if _, exist := p.handlers[inPortName]; exist {
		panic(fmt.Sprintf("[component-%s] in port of %s, already have handler", p.name, inPortName))
	} else {
		p.handlers[inPortName] = handler
	}
	return p
}

func (p *BaseComponent) BindReceiver(inPortName string, receivers ...MessageReceiver) Component {
	if inPortName == "" {
		panic(fmt.Sprintf("[component-%s] in port name could not be empty", p.name))
	}

	if receivers == nil || len(receivers) == 0 {
		panic(fmt.Sprintf("[component-%s] receivers could not be nil or 0 length, in port name: %s", p.name, inPortName))
	}

	inPortReceivers := map[MessageReceiver]bool{}

	for _, receiver := range receivers {
		if _, exist := inPortReceivers[receiver]; !exist {
			inPortReceivers[receiver] = true
		} else {
			panic(fmt.Sprintf("[component-%s] duplicate receiver type with in port, in port name: %s, receiver type: %s", p.name, inPortName, receiver.Type()))
		}
	}

	p.receivers[inPortName] = receivers

	return p
}

func (p *BaseComponent) RegisterSender(senders ...MessageSender) Component {
	if senders == nil || len(senders) == 0 {
		panic(fmt.Sprintf("[component-%s] senders could not be nil or 0 length", p.name))
	}

	componentSenders := map[string]MessageSender{}

	for _, sender := range senders {
		if _, exist := componentSenders[sender.Type()]; !exist {
			componentSenders[sender.Type()] = sender
		} else {
			panic(fmt.Sprintf("[component-%s] duplicate sender type with in port, sender type: %s", p.name, sender.Type()))
		}
	}

	p.senders = componentSenders

	return p
}

func (p *BaseComponent) Build() Component {
	p.runtimeLocker.Lock()
	defer p.runtimeLocker.Unlock()

	if p.isBuilded {
		panic(fmt.Sprintf("the component of %s already built", p.Name()))
	}

	for inPortName, _ := range p.receivers {
		p.errChans[inPortName] = make(chan error)
		p.messageChans[inPortName] = make(chan ComponentMessage)
	}

	p.isBuilded = true

	return nil
}

func (p *BaseComponent) Run() {
	p.runtimeLocker.Lock()
	defer p.runtimeLocker.Unlock()

	if !p.isBuilded {
		panic(fmt.Sprintf("the component of %s should be build first", p.Name()))
	}

	if p.isRunning == true {
		panic(fmt.Sprintf("the component of %s already running", p.Name()))
	}

	for inPortName, typedReceivers := range p.receivers {
		var errChan chan error
		var messageChan chan ComponentMessage
		exist := false
		if errChan, exist = p.errChans[inPortName]; !exist {
			panic(fmt.Sprintf("error chan of component: %s, not exist", p.Name()))
		}

		if messageChan, exist = p.messageChans[inPortName]; !exist {
			panic(fmt.Sprintf("error chan of component: %s, not exist", p.Name()))
		}

		for _, receiver := range typedReceivers {
			go receiver.Receive(messageChan, errChan)
		}
	}
	return
}

func (p *BaseComponent) ReceiverLoop() {
	loopInPortNames := []string{}
	for inPortName, _ := range p.receivers {
		loopInPortNames = append(loopInPortNames, inPortName)
	}
	for {
		for _, inPortName := range loopInPortNames {
			select {
			case compMsg := <-p.messageChans[inPortName]:
				{
					p.handlerComponentMessage(inPortName, compMsg)
				}
			}
		}
	}
}
func (p *BaseComponent) handlerComponentMessage(inPortName string, message ComponentMessage) {
	var handler ComponentHandler
	var err error
	var exist bool

	if message.graph == nil {
		logs.Error(ERR_MESSAGE_GRAPH_IS_NIL.New())
		return
	}

	var address MessageAddress
	var nextGraphIndex int32 = 0

	if address, exist = message.graph[message.currentGraphIndex+1]; exist {
		nextGraphIndex = message.currentGraphIndex + 1
	} else if address, exist = message.graph[0]; exist {
		nextGraphIndex = 0
	} else {
		err = ERR_MESSAGE_GRAPH_ADDRESS_NOT_FOUND.New(errors.Params{"compName": p.name, "portName": inPortName})
		logs.Error(err)
		return
	}

	if handler, exist = p.handlers[inPortName]; !exist {
		panic(fmt.Sprintf("in port of %s not exist", inPortName))
	}

	var content interface{}
	if content, err = handler(&message.payload); err != nil {
		err = ERR_COMPONENT_HANDLER_RETURN_ERROR.New(errors.Params{"err": err})
		logs.Error(err)

		//at first component, the first componet get the final result
		if nextGraphIndex == message.currentGraphIndex {
			return
		} else if address, exist = message.graph[0]; exist {
			nextGraphIndex = 0
			return
		} else {
			err = ERR_MESSAGE_GRAPH_ADDRESS_NOT_FOUND.New(errors.Params{"compName": p.name, "portName": inPortName})
			logs.Error(err)
			return
		}
	} else {
		message.payload.SetContent(content)
	}

	message.currentGraphIndex = nextGraphIndex

	var sender MessageSender
	if sender, exist = p.senders[address.Type]; !exist {
		err = ERR_SENDER_TYPE_NOT_EXIST.New(errors.Params{"compName": p.name, "type": address.Type})
		logs.Error(err)
		return
	}

	err = sender.Send(address.Url, message)
	logs.Error(err)
	return
}
