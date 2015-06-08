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
	inPortHooks   map[string][]string

	runtimeLocker sync.Mutex
	isBuilt       bool

	status ComponentStatus

	portChans     map[string]*PortChan
	stoppingChans map[string]chan bool
	stopedChans   map[string]chan bool
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
		inPortHooks:   make(map[string][]string),
		portChans:     make(map[string]*PortChan),
		stopedChans:   make(map[string]chan bool),
		stoppingChans: make(map[string]chan bool),
	}
}

func (p *BaseComponent) Name() string {
	return p.name
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

func (p *BaseComponent) AddInPortHooks(inportName string, hooks ...string) Component {
	p.runtimeLocker.Lock()
	defer p.runtimeLocker.Unlock()

	if inportName == "" {
		panic("inportName could not be empty")
	}

	for _, hook := range hooks {
		if hook == "" {
			panic("hook could not be empty")
		}
	}

	if originalHooks, exist := p.inPortHooks[inportName]; exist {
		p.inPortHooks[inportName] = append(originalHooks, hooks...)
	} else {
		p.inPortHooks[inportName] = hooks
	}

	return p
}

func (p *BaseComponent) ClearInPortHooks(inportName string) (err error) {
	p.runtimeLocker.Lock()
	defer p.runtimeLocker.Unlock()

	if inportName == "" {
		err = ERR_PORT_NAME_IS_EMPTY.New(errors.Params{"name": p.name})
		return
	}

	if _, exist := p.inPortHooks[inportName]; !exist {
		return
	} else {
		delete(p.inPortHooks, inportName)
	}
	return
}

func (p *BaseComponent) Build() Component {
	p.runtimeLocker.Lock()
	defer p.runtimeLocker.Unlock()

	if p.isBuilt {
		panic(fmt.Sprintf("the component of %s already built", p.name))
	}

	if senderFactory == nil {
		panic(fmt.Sprintf("the component of %s did not have sender factory", p.name))
	}

	for inPortName, _ := range p.receivers {
		portChan := new(PortChan)
		portChan.Error = make(chan error, MESSAGE_CHAN_SIZE)
		portChan.Message = make(chan ComponentMessage, MESSAGE_CHAN_SIZE)
		portChan.Signal = make(chan int)
		portChan.Stoped = make(chan bool)

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
			stoplogTime := time.Now()
			for {
				if isStopping {
					now := time.Now()

					if len(respChan) == 0 && len(errChan) == 0 {
						logs.Warn(fmt.Sprintf("* port - %s have no message, so it will be stop running", portName))
						stopedChan <- true
						return
					} else {
						if now.Sub(stoplogTime) >= time.Second {
							stoplogTime = now
							logs.Warn(fmt.Sprintf("* port - %s stopping, MsgLen: %d, ErrLen: %d", portName, len(respChan), len(errChan)))
						}
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
						stoplogTime = time.Now()
						logs.Warn(fmt.Sprintf("* port - %s received stop signal", portName))
					}
				case <-time.After(time.Second):
					{
						if len(respChan) == 0 && isStopping {
							stopedChan <- true
							return
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
		wgReceiverPause := sync.WaitGroup{}
		for _, Chans := range p.portChans {
			wgReceiverPause.Add(1)
			go func(singalChan chan int) {
				defer wgReceiverPause.Done()
				select {
				case singalChan <- SIG_PAUSE:
				case <-time.After(time.Second * 5):
				}
			}(Chans.Signal)
		}
		wgReceiverPause.Wait()
		p.status = STATUS_PAUSED
	} else if p.status == STATUS_PAUSED {
		wgReceiverResume := sync.WaitGroup{}
		for _, Chans := range p.portChans {
			wgReceiverResume.Add(1)
			go func(singalChan chan int) {
				defer wgReceiverResume.Done()
				select {
				case singalChan <- SIG_RESUME:
				case <-time.After(time.Second * 5):
				}
			}(Chans.Signal)
		}
		wgReceiverResume.Wait()
		p.status = STATUS_RUNNING
	} else {
		logs.Warn("[base component] pause or resume at error status")
	}

}

func (p *BaseComponent) Stop() {
	p.runtimeLocker.Lock()
	defer p.runtimeLocker.Unlock()

	//stop queues first
	logs.Warn("* begin stop port receivers")
	wgReceiverBeginStop := sync.WaitGroup{}
	for _, Chans := range p.portChans {
		wgReceiverBeginStop.Add(1)
		go func(singalChan chan int) {
			defer wgReceiverBeginStop.Done()
			select {
			case singalChan <- SIG_STOP:
			case <-time.After(time.Second * 5):
			}
		}(Chans.Signal)
	}
	wgReceiverBeginStop.Wait()

	logs.Warn("* waiting for port receivers stopped signal")

	wgReceiverStop := sync.WaitGroup{}
	for _, Chans := range p.portChans {
		wgReceiverStop.Add(1)
		go func(stopedChan chan bool) {
			defer wgReceiverStop.Done()
			select {
			case _ = <-stopedChan:
			case <-time.After(time.Second * 60):
			}
		}(Chans.Stoped)
	}
	wgReceiverStop.Wait()

	logs.Warn("* begin stop received response message handler")
	wgHandlerBeginStop := sync.WaitGroup{}
	for inportName, Chan := range p.stoppingChans {
		wgHandlerBeginStop.Add(1)
		go func(stopedChan chan bool, name string) {
			defer wgHandlerBeginStop.Done()
			select {
			case stopedChan <- true:
				{
					logs.Warn("* component begin stop port:", name)
				}
			case <-time.After(time.Second * 60):
			}
		}(Chan, inportName)
	}
	wgHandlerBeginStop.Wait()

	logs.Warn("* waiting for received response message handler stopped signal")
	wgHandlerStop := sync.WaitGroup{}
	for inportName, Chan := range p.stopedChans {
		wgHandlerStop.Add(1)
		go func(stopedChan chan bool, name string) {
			defer wgHandlerStop.Done()
			select {
			case _ = <-stopedChan:
				{
					logs.Warn("* component", name, "stoped")
				}
			case <-time.After(time.Second * 60):
			}
		}(Chan, inportName)
	}
	wgHandlerStop.Wait()

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

	var newMetadatas []MessageHookMetadata
	if _, err = p.hookMessages(inPortName, HookEventBeforeCallHandler, &message); err != nil {
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

			if newMetadatas, err = p.hookMessages(inPortName, HookEventAfterCallHandler, &message); err != nil {
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
			}

			if address, exist = message.graph[strconv.Itoa(int(message.currentGraphIndex)+1)]; exist {
				nextGraphIndex = message.currentGraphIndex + 1 //forword the next component
			} else {
				return
			}
		}
	}

	message.hooksMetaData = newMetadatas
	message.currentGraphIndex = nextGraphIndex

	go p.sendMessage(address.Type, address.Url, message)

	return
}

func (p *BaseComponent) hookMessages(inPortName string, event HookEvent, message *ComponentMessage) (metadatas []MessageHookMetadata, err error) {
	if event == HookEventBeforeCallHandler {
		return p.hookMessagesBefore(inPortName, message)
	} else if event == HookEventAfterCallHandler {
		return p.hookMessagesAfter(inPortName, message)
	}
	return
}

func (p *BaseComponent) hookMessagesBefore(inPortName string, message *ComponentMessage) (metadatas []MessageHookMetadata, err error) {
	preMetadata := message.hooksMetaData
	if preMetadata == nil || len(preMetadata) == 0 {
		return
	}

	var hooks []string
	var exist bool

	if hooks, exist = p.inPortHooks[inPortName]; !exist {
		err = ERR_INPORT_NOT_BIND_HOOK.New(errors.Params{"inPortName": inPortName})
		return
	}

	newMetadatas := []MessageHookMetadata{}
	for index, metadata := range preMetadata {
		var hook MessageHook
		if hook, err = hookFactory.Get(metadata.HookName); err != nil {
			return
		}

		if ignored, matadata, e := hook.HookBefore(metadata, message.hooksMetaData, newMetadatas, &message.payload); e != nil {
			if !errors.IsErrCode(e) {
				e = ERR_MESSAGE_HOOK_ERROR.New(errors.Params{"err": e, "name": p.name, "port": inPortName, "index": index, "count": len(hooks), "hookName": hook.Name(), "event": "before"})
			}
			logs.Error(e)

			if !ignored {
				err = e
				return
			}
		} else if !ignored {
			newMetadatas = append(newMetadatas, matadata)
		}
	}
	metadatas = newMetadatas
	return
}

func (p *BaseComponent) hookMessagesAfter(inPortName string, message *ComponentMessage) (metadatas []MessageHookMetadata, err error) {
	var hooks []string
	var exist bool

	if hooks, exist = p.inPortHooks[inPortName]; !exist {
		return
	}

	newMetadatas := []MessageHookMetadata{}
	for index, hookName := range hooks {
		var hook MessageHook
		if hook, err = hookFactory.Get(hookName); err != nil {
			return
		}

		if ignored, matadata, e := hook.HookAfter(message.hooksMetaData, newMetadatas, &message.payload); e != nil {
			if !errors.IsErrCode(e) {
				e = ERR_MESSAGE_HOOK_ERROR.New(errors.Params{"err": e, "name": p.name, "port": inPortName, "index": index, "count": len(hooks), "hookName": hook.Name(), "event": "after"})
			}
			logs.Error(e)

			if !ignored {
				err = e
				return
			}
		} else if !ignored {
			matadata.HookName = hookName
			newMetadatas = append(newMetadatas, matadata)
		}
	}
	metadatas = newMetadatas
	return
}

func (p *BaseComponent) sendMessage(addrType, url string, msg ComponentMessage) {
	var sender MessageSender
	var err error
	if sender, err = senderFactory.NewSender(addrType); err != nil {
		logs.Error(err)
		return
	}

	if err = sender.Send(url, msg); err != nil {
		logs.Error(err)
		return
	}
}
