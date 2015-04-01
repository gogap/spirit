package spirit

import (
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/gogap/errors"
	"github.com/gogap/logs"
)

var (
	MESSAGE_CHAN_SIZE = 1000
)

type BaseComponent struct {
	name          string
	receivers     map[string][]MessageReceiver
	handlers      map[string]ComponentHandler
	inPortHandler map[string]ComponentHandler

	runtimeLocker sync.Mutex
	isBuilt       bool

	status ComponentStatus

	portChans     map[string]*PortChan
	stoppingChans map[string]chan bool
	stopedChans   map[string]chan bool

	senderFactory MessageSenderFactory
}

func NewBaseComponent(componentName string) Component {
	if componentName == "" {
		panic("component name could not be empty")
	}

	return &BaseComponent{
		name:          componentName,
		receivers:     make(map[string][]MessageReceiver),
		handlers:      make(map[string]ComponentHandler),
		inPortHandler: make(map[string]ComponentHandler),
		portChans:     make(map[string]*PortChan),
		stopedChans:   make(map[string]chan bool),
		stoppingChans: make(map[string]chan bool),
	}
}

func (p *BaseComponent) Name() string {
	return p.name
}

func (p *BaseComponent) SetMessageSenderFactory(factory MessageSenderFactory) Component {
	if factory == nil {
		panic(fmt.Sprintf("message sender factory could not be nil, component name: %s", p.name))
	}
	p.senderFactory = factory
	return p
}

func (p *BaseComponent) CallHandler(handlerName string, payload *Payload) (result interface{}, err error) {
	if handlerName == "" {
		err = ERR_HANDLER_NAME_IS_EMPTY.New(errors.Params{"name": p.name})
		return
	}

	if handler, exist := p.handlers[handlerName]; exist {
		if ret, e := handler(payload); e != nil {
			err = ERR_COMPONENT_HANDLER_RETURN_ERROR.New(errors.Params{"err": e})
			return
		} else {
			result = ret
			return
		}
	} else {
		err = ERR_COMPONENT_HANDLER_NOT_EXIST.New(errors.Params{"name": p.name, "handlerName": handlerName})
		return
	}
}

func (p *BaseComponent) RegisterHandler(name string, handler ComponentHandler) Component {
	if name == "" {
		panic(fmt.Sprintf("[component-%s] handle name could not be empty", p.name))
	}

	if handler == nil {
		panic(fmt.Sprintf("[component-%s] handler could not be nil, handler name: %s", p.name, name))
	}

	if _, exist := p.handlers[name]; exist {
		panic(fmt.Sprintf("[component-%s] handler of %s, already registered", p.name, name))
	} else {
		p.handlers[name] = handler
	}
	return p
}

func (p *BaseComponent) ListHandlers() (handlers map[string]ComponentHandler, err error) {
	handlers = make(map[string]ComponentHandler)
	for name, handler := range p.handlers {
		handlers[name] = handler
	}
	return
}

func (p *BaseComponent) GetHandlers(handlerNames ...string) (handlers map[string]ComponentHandler, err error) {
	if handlerNames == nil {
		return
	}

	ret := make(map[string]ComponentHandler)
	for _, name := range handlerNames {
		if h, exist := p.handlers[name]; !exist {
			err = ERR_COMPONENT_HANDLER_NOT_EXIST.New(errors.Params{"name": p.name, "handlerName": name})
			return
		} else {
			ret[name] = h
		}
	}
	handlers = ret
	return
}

func (p *BaseComponent) BindHandler(inPortName, handlerName string) Component {
	if inPortName == "" {
		panic(fmt.Sprintf("[component-%s] in port name could not be empty", p.name))
	}

	if handlerName == "" {
		panic(fmt.Sprintf("[component-%s] handler name could not be empty, in port name: %s", p.name, inPortName))
	}

	var handler ComponentHandler
	exist := false

	if handler, exist = p.handlers[handlerName]; !exist {
		panic(fmt.Sprintf("[component-%s] handler not exist, handler name: %s", p.name, handlerName))
	}

	if _, exist = p.inPortHandler[inPortName]; exist {
		panic(fmt.Sprintf("[component-%s] in port of %s, already have handler, handler name: %s", p.name, inPortName, handlerName))
	} else {
		p.inPortHandler[inPortName] = handler
	}

	return p
}

func (p *BaseComponent) GetReceivers(inPortName string) []MessageReceiver {
	if inPortName == "" {
		panic(fmt.Sprintf("[component-%s] in port name could not be empty", p.name))
	}

	receivers, _ := p.receivers[inPortName]
	return receivers
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

func (p *BaseComponent) Build() Component {
	p.runtimeLocker.Lock()
	defer p.runtimeLocker.Unlock()

	if p.isBuilt {
		panic(fmt.Sprintf("the component of %s already built", p.name))
	}

	if p.senderFactory == nil {
		panic(fmt.Sprintf("the component of %s did not have sender factory", p.name))
	}

	for inPortName, _ := range p.receivers {
		portChan := new(PortChan)
		portChan.Error = make(chan error, MESSAGE_CHAN_SIZE)
		portChan.Message = make(chan ComponentMessage, MESSAGE_CHAN_SIZE)
		portChan.Signal = make(chan int)

		p.portChans[inPortName] = portChan
		p.stopedChans[inPortName] = make(chan bool)
		p.stoppingChans[inPortName] = make(chan bool)
	}

	p.isBuilt = true

	return p
}

func (p *BaseComponent) Run() {
	p.runtimeLocker.Lock()
	defer p.runtimeLocker.Unlock()

	if !p.isBuilt {
		panic(fmt.Sprintf("the component of %s should be build first", p.name))
	}

	if p.status > 0 {
		panic(fmt.Sprintf("the component of %s already running", p.name))
	}

	for inPortName, typedReceivers := range p.receivers {
		var portChan *PortChan
		exist := false
		if portChan, exist = p.portChans[inPortName]; !exist {
			panic(fmt.Sprintf("port chan of component: %s, not exist", p.name))
		}

		for _, receiver := range typedReceivers {
			go receiver.Receive(portChan)
		}
	}

	p.status = STATUS_RUNNING

	p.ReceiverLoop()
	return
}

func (p *BaseComponent) ReceiverLoop() {
	loopInPortNames := []string{}
	for inPortName, _ := range p.receivers {
		loopInPortNames = append(loopInPortNames, inPortName)
	}

	for _, inPortName := range loopInPortNames {
		portChan := p.portChans[inPortName]
		stopedChan := p.stopedChans[inPortName]
		stoppingChan := p.stoppingChans[inPortName]
		go func(portName string, respChan chan ComponentMessage, errChan chan error, stoppingChan chan bool, stopedChan chan bool) {
			isStopping := false
			for {
				if isStopping {
					if len(respChan) == 0 && len(errChan) == 0 {
						logs.Warn(fmt.Sprintf("* port - %s have no message, so it will be stop running", portName))
						stopedChan <- true
						return
					} else {
						logs.Warn(fmt.Sprintf("* port - %s stopping, MsgLen: %d, ErrLen: %d", portName, len(respChan), len(errChan)))
						time.Sleep(time.Second)
					}
				}
				select {
				case compMsg := <-respChan:
					{
						go p.handleComponentMessage(portName, compMsg)
					}
				case respErr := <-errChan:
					{
						logs.Error(respErr)
					}
				case isStopping = <-stoppingChan:
					{
						logs.Warn(fmt.Sprintf("* port - %s received stop signal", portName))
					}
				case <-time.After(time.Second):
					{
						if len(respChan) == 0 && isStopping {
							stopedChan <- true
							break
						}
					}
				}
			}
		}(inPortName, portChan.Message, portChan.Error, stoppingChan, stopedChan)
	}
}

func (p *BaseComponent) PauseOrResume() {
	p.runtimeLocker.Lock()
	defer p.runtimeLocker.Unlock()

	if p.status == STATUS_RUNNING {
		for _, Chans := range p.portChans {
			select {
			case Chans.Signal <- SIG_PAUSE:
			case <-time.After(time.Second * 2):
			}
		}
		p.status = STATUS_PAUSED
	} else if p.status == STATUS_PAUSED {
		for _, Chans := range p.portChans {
			select {
			case Chans.Signal <- SIG_RESUME:
			case <-time.After(time.Second * 2):
			}
		}
		p.status = STATUS_RUNNING
	} else {
		logs.Warn("[base component] pause or resume at error status")
	}

}

func (p *BaseComponent) Stop() {
	p.runtimeLocker.Lock()
	defer p.runtimeLocker.Unlock()

	for _, Chans := range p.portChans {
		select {
		case Chans.Signal <- SIG_STOP:
		case <-time.After(time.Second):
		}
	}

	for _, Chan := range p.stoppingChans {
		select {
		case Chan <- true:
		case <-time.After(time.Second * 60):
		}
	}

	for inportName, stopedChan := range p.stopedChans {
		select {
		case _ = <-stopedChan:
			{
				logs.Warn("* component", inportName, "stoped")
			}
		}
	}

	p.status = STATUS_STOPED
}

func (p *BaseComponent) Status() ComponentStatus {
	return p.status
}

func (p *BaseComponent) callHandlerWithRecover(handler ComponentHandler, payload *Payload) (content interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 1024)
			runtime.Stack(buf, false)
			err = ERR_COMPONENT_HANDLER_PANIC.New(errors.Params{"name": p.name, "err": string(buf)})
		}
	}()

	return handler(payload)
}

func (p *BaseComponent) handleComponentMessage(inPortName string, message ComponentMessage) {
	var handler ComponentHandler
	var err error
	var exist bool

	if message.graph == nil {
		logs.Error(ERR_MESSAGE_GRAPH_IS_NIL.New())
		return
	}

	if handler, exist = p.inPortHandler[inPortName]; !exist {
		panic(fmt.Sprintf("in port of %s not exist", inPortName))
	}

	var address MessageAddress
	var nextGraphIndex int32 = 0
	var content interface{}

	if content, err = p.callHandlerWithRecover(handler, &message.payload); err != nil {
		if !errors.IsErrCode(err) {
			err = ERR_COMPONENT_HANDLER_RETURN_ERROR.New(errors.Params{"err": err, "name": p.name, "port": inPortName})
		}

		logs.Error(err)

		if address, exist = message.graph[ERROR_MSG_ADDR]; exist {
			errCode := err.(errors.ErrCode)
			message.payload.err.AddressId = message.currentGraphIndex
			message.payload.err.Id = errCode.Id()
			message.payload.err.Namespace = errCode.Namespace()
			message.payload.err.Code = errCode.Code()
			message.payload.err.Message = errCode.Error()

			nextGraphIndex = ERROR_MSG_ADDR_INT //forword to the error port
		} else {
			return
		}
	} else {
		message.payload.SetContent(content)
		if address, exist = message.graph[strconv.Itoa(int(message.currentGraphIndex)+1)]; exist {
			nextGraphIndex = message.currentGraphIndex + 1 //forword the next component
		} else {
			return
		}
	}

	message.currentGraphIndex = nextGraphIndex

	go func(addrType, url string, msg ComponentMessage) {
		var sender MessageSender
		if sender, err = p.senderFactory.NewSender(addrType); err != nil {
			logs.Error(err)
			return
		}

		if err = sender.Send(url, msg); err != nil {
			logs.Error(err)
			return
		}
	}(address.Type, address.Url, message)

	return
}
