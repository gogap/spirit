package spirit

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/gogap/errors"
)

type MessageHookFactory interface {
	RegisterMessageHooks(hooks ...MessageHook)
	IsExist(hookType string) bool
	CreateHook(hookType, name string, options Options) (hook MessageHook, err error)
	Get(name string) (hook MessageHook, err error)
}

type DefaultMessageHookFactory struct {
	hookDrivers   map[string]reflect.Type
	instanceCache map[string]MessageHook

	instanceLock sync.Mutex
}

func NewDefaultMessageHookFactory() MessageHookFactory {
	fact := new(DefaultMessageHookFactory)
	fact.hookDrivers = make(map[string]reflect.Type)
	fact.instanceCache = make(map[string]MessageHook)
	return fact
}

func (p *DefaultMessageHookFactory) RegisterMessageHooks(hooks ...MessageHook) {
	if hooks == nil {
		panic("hooks is nil")
	}

	if len(hooks) == 0 {
		return
	}

	for _, hook := range hooks {
		if _, exist := p.hookDrivers[hook.Name()]; exist {
			panic(fmt.Errorf("hook driver of [%s] already exist", hook.Name()))
		}

		vof := reflect.ValueOf(hook)
		vType := vof.Type()
		if vof.Kind() == reflect.Ptr {
			vType = vof.Elem().Type()
		}

		p.hookDrivers[hook.Name()] = vType
	}
	return
}

func (p *DefaultMessageHookFactory) IsExist(hookType string) bool {
	if _, exist := p.hookDrivers[hookType]; exist {
		return true
	}
	return false
}

func (p *DefaultMessageHookFactory) CreateHook(hookType, name string, options Options) (hook MessageHook, err error) {
	p.instanceLock.Lock()
	defer p.instanceLock.Unlock()

	if hookType == "" {
		err = ERR_HOOK_DRIVER_TYPE_IS_EMPTY.New()
		return
	}

	if name == "" {
		err = ERR_HOOK_NAME_IS_EMPTY.New()
		return
	}

	if _, exist := p.instanceCache[name]; exist {
		err = ERR_HOOK_INSTANCE_ALREADY_INITALED.New(errors.Params{"type": hookType, "name": name})
		return
	}

	if hookDriver, exist := p.hookDrivers[hookType]; !exist {
		err = ERR_HOOK_DRIVER_NOT_EXIST.New(errors.Params{"type": hookType})
		return
	} else {
		if vOfMessageHook := reflect.New(hookDriver); vOfMessageHook.CanInterface() {
			iMessageHook := vOfMessageHook.Interface()
			if r, ok := iMessageHook.(MessageHook); ok {
				if err = r.Init(options); err != nil {
					return
				}
				hook = r
				p.instanceCache[name] = r
				return
			} else {
				err = ERR_HOOK_CREATE_FAILED.New(errors.Params{"type": hookType, "name": name})
				return
			}
		}
		err = ERR_HOOK_BAD_DRIVER.New(errors.Params{"type": hookType})
		return
	}
}

func (p *DefaultMessageHookFactory) Get(name string) (hook MessageHook, err error) {
	if instance, exist := p.instanceCache[name]; !exist {
		err = ERR_HOOK_INSTANCE_NOT_INITALED.New(errors.Params{"name": name})
		return
	} else {
		hook = instance
	}
	return
}
