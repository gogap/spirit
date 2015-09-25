package spirit

import (
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gogap/errors"
	"github.com/gogap/logs"
)

var (
	MESSAGE_CHAN_SIZE = 1000
)

type BaseComponent struct {
	name           string
	receivers      map[string][]MessageReceiver
	handlers       map[string]ComponentHandler
	inPortHandler  map[string]ComponentHandler
	inPortHooks    map[string][]string
	inPortUrlIndex map[string]string

	inboxMessage int64
	inboxError   int64

	runtimeLocker sync.Mutex
	isBuilt       bool

	status ComponentStatus

	isStopping bool
}

func NewBaseComponent(componentName string) Component {
	if componentName == "" {
		panic("component name could not be empty")
	}

	return &BaseComponent{
		name:           componentName,
		receivers:      make(map[string][]MessageReceiver),
		handlers:       make(map[string]ComponentHandler),
		inPortHandler:  make(map[string]ComponentHandler),
		inPortHooks:    make(map[string][]string),
		inPortUrlIndex: make(map[string]string),
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

			receiver.BindInPort(p.name, inPortName, p.handleComponentMessage, p.handleRecevierError)
			addr := receiver.Address()
			p.inPortUrlIndex[addr.Type+"|"+addr.Url] = inPortName

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

	p.isBuilt = true

	return p
}

func (p *BaseComponent) Run() {
	p.runtimeLocker.Lock()
	defer p.runtimeLocker.Unlock()

	if !p.isBuilt {
		panic(fmt.Sprintf("the component of %s should be build first", p.name))
	}

	if p.status != STATUS_READY && p.status != STATUS_STOPPED {
		panic(fmt.Sprintf("the component of %s already running", p.name))
	}

	p.startReceivers()

	p.status = STATUS_RUNNING

	return
}

func (p *BaseComponent) startReceivers() {
	atomic.SwapInt64(&p.inboxMessage, 0)
	atomic.SwapInt64(&p.inboxError, 0)

	wg := sync.WaitGroup{}
	for _, typedReceivers := range p.receivers {
		for _, receiver := range typedReceivers {
			wg.Add(1)
			go func(receiver MessageReceiver) {
				defer wg.Done()
				receiver.Start()
				for !receiver.IsRunning() {
					time.Sleep(time.Millisecond * 100)
				}
				EventCenter.PushEvent(EVENT_RECEIVER_STARTED, receiver.Metadata())
			}(receiver)
		}
	}
	wg.Wait()
}

func (p *BaseComponent) stopReceivers() {
	wg := sync.WaitGroup{}
	for _, typedReceivers := range p.receivers {
		for _, receiver := range typedReceivers {
			wg.Add(1)
			go func(receiver MessageReceiver) {
				defer wg.Done()
				receiver.Stop()
				EventCenter.PushEvent(EVENT_RECEIVER_STOPPED, receiver.Metadata())
			}(receiver)
		}
	}
	wg.Wait()
}

func (p *BaseComponent) Pause() {
	p.runtimeLocker.Lock()
	defer p.runtimeLocker.Unlock()

	if p.status == STATUS_RUNNING {
		p.stopReceivers()
		p.status = STATUS_PAUSED
	} else if p.status == STATUS_PAUSED {
		return
	} else {
		logs.Warn("[base component] pause or resume at error status")
		return
	}
}

func (p *BaseComponent) Resume() {
	p.runtimeLocker.Lock()
	defer p.runtimeLocker.Unlock()

	if p.status == STATUS_RUNNING {
		return
	} else if p.status == STATUS_PAUSED {
		p.startReceivers()
		p.status = STATUS_RUNNING
	} else {
		logs.Warn("[base component] pause or resume at error status")
		return
	}
}

func (p *BaseComponent) Stop() {
	p.runtimeLocker.Lock()
	defer p.runtimeLocker.Unlock()

	if p.status != STATUS_RUNNING && p.status != STATUS_PAUSED {
		logs.Warn("[base component] stop at error status, current status: ", p.status)
		return
	}

	p.status = STATUS_STOPPING
	EventCenter.PushEvent(EVENT_BASE_COMPONENT_STOPPING, p.name, p.inboxMessage, p.inboxError)

	p.stopReceivers()

	for p.inboxMessage != 0 || p.inboxError != 0 {
		time.Sleep(time.Second)
		EventCenter.PushEvent(EVENT_BASE_COMPONENT_STOPPING, p.name, p.inboxMessage, p.inboxError)
	}

	p.status = STATUS_STOPPED
	EventCenter.PushEvent(EVENT_BASE_COMPONENT_STOPPED, p.name)
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

func (p *BaseComponent) handleComponentMessage(inPortName string, messageId string, message ComponentMessage, callbacks ...OnReceiverMessageProcessed) {
	atomic.AddInt64(&p.inboxMessage, 1)
	defer atomic.AddInt64(&p.inboxMessage, -1)

	defer func() {
		if callbacks != nil {
			for _, callback := range callbacks {
				callback(messageId)
			}
		}
	}()

	var handler ComponentHandler
	var err error
	var exist bool

	if message.graph == nil {
		logs.Error(ERR_MESSAGE_GRAPH_IS_NIL.New())
		return
	}

	if handler, exist = p.inPortHandler[inPortName]; !exist {
		logs.Error(fmt.Sprintf("in port of %s not exist", inPortName))
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

	if nextInPortName, exist := p.inPortUrlIndex[address.Type+"|"+address.Url]; exist {
		p.handleComponentMessage(nextInPortName, "", message)
		EventCenter.PushEvent(EVENT_BASE_COMPONENT_INTERNAL_MESSAGE, p.name, address, message)
	} else {
		p.sendMessage(address, message)
	}

	return
}

func (p *BaseComponent) handleRecevierError(inportName string, err error) {
	atomic.AddInt64(&p.inboxError, 1)
	defer atomic.AddInt64(&p.inboxError, -1)

	logs.Error(err)
}

func (p *BaseComponent) hookMessages(inPortName string, event HookEvent, message *ComponentMessage) (metadatas []MessageHookMetadata, err error) {
	EventCenter.PushEvent(EVENT_BEFORE_MESSAGE_HOOK, event)
	defer EventCenter.PushEvent(EVENT_AFTER_MESSAGE_HOOK, event)

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

func (p *BaseComponent) sendMessage(address MessageAddress, msg ComponentMessage) {
	var sender MessageSender
	var err error
	if sender, err = senderFactory.NewSender(address.Type); err != nil {
		logs.Error(err)
		return
	}

	if err = sender.Send(address.Url, msg); err != nil {
		logs.Error(err)
		return
	}
}
