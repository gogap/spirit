package spirit

import (
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/gogap/errors"
	"github.com/gogap/event_center"
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

	portChans   map[string]*PortChan
	stoppedPort map[string]bool
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
		stoppedPort:   make(map[string]bool),
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
	EventCenter.PushEvent(EVENT_BASE_COMPONENT_BEFORE_HANDLER_REGISTER, name)
	defer EventCenter.PushEvent(EVENT_BASE_COMPONENT_AFTER_HANDLER_REGISTER, name)

	if name == "" {
		panic(fmt.Sprintf("[component-%s] handle name could not be empty", p.name))
	}

	if handler == nil {
		panic(fmt.Sprintf("[component-%s] handler could not be nil, handler name: %s", p.name, name))
	}

	if _, exist := p.handlers[name]; exist {
		panic(fmt.Sprintf("[component-%s] handler of %s, already registered", p.name, name))
	} else {
		EventCenter.PushEvent(EVENT_BASE_COMPONENT_HANDLER_REGISTERED, name)
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
	EventCenter.PushEvent(EVENT_BASE_COMPONENT_BEFORE_HANDLER_BIND, p.name, inPortName, handlerName)
	defer EventCenter.PushEvent(EVENT_BASE_COMPONENT_AFTER_HANDLER_BIND, p.name, inPortName, handlerName)

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
		EventCenter.PushEvent(EVENT_BASE_COMPONENT_HANDLER_BOUND, p.name, inPortName, handlerName)
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
	EventCenter.PushEvent(EVENT_BASE_COMPONENT_BEFORE_RECEIVER_BIND, p.name, inPortName)
	defer EventCenter.PushEvent(EVENT_BASE_COMPONENT_AFTER_RECEIVER_BIND, p.name, inPortName)

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
	EventCenter.PushEvent(EVENT_BASE_COMPONENT_RECEIVER_BOUND, p.name, inPortName)

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
		portChan.ComponentName = p.name
		portChan.PortName = inPortName
		portChan.Error = make(chan error, MESSAGE_CHAN_SIZE)
		portChan.Message = make(chan ComponentMessage, MESSAGE_CHAN_SIZE)

		p.portChans[inPortName] = portChan
		p.stoppedPort[inPortName] = false
	}

	p.isBuilt = true

	return p
}

func (p *BaseComponent) Run() {
	p.runtimeLocker.Lock()
	defer p.runtimeLocker.Unlock()

	EventCenter.PushEvent(EVENT_BASE_COMPONENT_BEFORE_RUN, p.name)
	defer EventCenter.PushEvent(EVENT_BASE_COMPONENT_AFTER_RUN, p.name)

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

		go func(portName string, respChan chan ComponentMessage, errChan chan error) {
			isStopping := false
			isStopped := false
			isPaused := false

			stopSubscriber := event_center.NewSubscriber(
				func(eventName string, values ...interface{}) {
					if !isStopping {
						EventCenter.PushEvent(EVENT_BASE_COMPONENT_STOPPING, p.name, portName)
						return
					}
				})

			EventCenter.Subscribe(EVENT_CMD_STOP, stopSubscriber)
			defer EventCenter.Unsubscribe(EVENT_CMD_STOP, stopSubscriber.Id())

			stopingSubscriber := event_center.NewSubscriber(
				func(eventName string, values ...interface{}) {
					if isStopping {
						return
					}

					if values == nil || len(values) != 2 {
						return
					}

					if name := values[0].(string); name != p.name {
						return
					}

					if name := values[1].(string); name != portName {
						return
					}

					isStopping = true
				})

			EventCenter.Subscribe(EVENT_BASE_COMPONENT_STOPPING, stopingSubscriber)
			defer EventCenter.Unsubscribe(EVENT_BASE_COMPONENT_STOPPING, stopingSubscriber.Id())

			receiverStoppedSubscriber := event_center.NewSubscriber(func(eventName string, values ...interface{}) {
				if !isStopping {
					return
				}
				if values == nil || len(values) != 1 {
					return
				}

				if metadata := values[0].(ReceiverMetadata); metadata.PortName != portName || metadata.ComponentName != p.name {
					return
				}

				go func() {
					for {
						if len(respChan) == 0 && len(errChan) == 0 {
							EventCenter.PushEvent(EVENT_BASE_COMPONENT_STOPPED, p.name, inPortName)
							return
						}
					}
				}()
			})

			EventCenter.Subscribe(EVENT_RECEIVER_STOPPED, receiverStoppedSubscriber)
			defer EventCenter.Unsubscribe(EVENT_RECEIVER_STOPPED, receiverStoppedSubscriber.Id())

			stopedSubscriber := event_center.NewSubscriber(func(eventName string, values ...interface{}) {
				if !isStopping {
					return
				}

				if values == nil || len(values) != 2 {
					return
				}

				if name := values[0].(string); name != p.name {
					return
				}

				if name := values[1].(string); name != portName {
					return
				}

				isStopped = true
				isStopping = false
				isPaused = false

			})

			EventCenter.Subscribe(EVENT_BASE_COMPONENT_STOPPED, stopedSubscriber)
			defer EventCenter.Unsubscribe(EVENT_BASE_COMPONENT_STOPPED, stopedSubscriber.Id())

			cmdPauseSubscriber := event_center.NewSubscriber(
				func(eventName string, values ...interface{}) {
					if isStopped || isStopping || isPaused {
						return
					}

					isPaused = true

					EventCenter.PushEvent(EVENT_BASE_COMPONENT_PAUSED, p.name, portName)
				})

			EventCenter.Subscribe(EVENT_CMD_PAUSE, cmdPauseSubscriber)
			defer EventCenter.Unsubscribe(EVENT_CMD_PAUSE, cmdPauseSubscriber.Id())

			for {
				if isPaused {
					time.Sleep(time.Second)
					continue
				}

				if isStopped {
					p.stoppedPort[portName] = true
					return
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
				case <-time.After(time.Second):
					{
					}
				}
			}
		}(inPortName, portChan.Message, portChan.Error)
	}
}

func (p *BaseComponent) PauseOrResume() {
	p.runtimeLocker.Lock()
	defer p.runtimeLocker.Unlock()

	EventCenter.PushEvent(EVENT_CMD_PAUSE_OR_RESUME, p.name)

	if p.status == STATUS_RUNNING {
		EventCenter.PushEvent(EVENT_CMD_PAUSE, p.name)

		p.status = STATUS_PAUSED
	} else if p.status == STATUS_PAUSED {

		EventCenter.PushEvent(EVENT_AFTER_RESUME, p.name)
		p.status = STATUS_RUNNING
	} else {
		logs.Warn("[base component] pause or resume at error status")
	}

}

func (p *BaseComponent) Stop() {
	p.runtimeLocker.Lock()
	defer p.runtimeLocker.Unlock()

	EventCenter.PushEvent(EVENT_CMD_STOP, p.name)

	allStopped := false
	for !allStopped {

		allStopped = true
		for _, isStoped := range p.stoppedPort {
			if isStoped == false {
				allStopped = false
				time.Sleep(time.Second)
			}
		}
	}

	p.status = STATUS_STOPPED
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
	EventCenter.PushEvent(EVENT_BEFORE_HOOK, event)
	defer EventCenter.PushEvent(EVENT_BEFORE_HOOK, event)

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
