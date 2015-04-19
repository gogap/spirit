package spirit

import (
	"fmt"
	"reflect"

	"github.com/gogap/errors"
)

type MessageHookFactory interface {
	RegisterMessageHooks(hooks ...MessageHook)
	IsExist(hookType string) bool
	NewHook(hookType, configFile string) (hook MessageHook, err error)
}

type DefaultMessageHookFactory struct {
	hookDrivers map[string]reflect.Type
}

func NewDefaultMessageHookFactory() MessageHookFactory {
	fact := new(DefaultMessageHookFactory)
	fact.hookDrivers = make(map[string]reflect.Type)
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

func (p *DefaultMessageHookFactory) NewHook(hookType, configFile string) (hook MessageHook, err error) {
	if hookType, exist := p.hookDrivers[hookType]; !exist {
		err = ERR_HOOK_DRIVER_NOT_EXIST.New(errors.Params{"type": hookType})
		return
	} else {
		if vOfMessageHook := reflect.New(hookType); vOfMessageHook.CanInterface() {
			iMessageHook := vOfMessageHook.Interface()
			if r, ok := iMessageHook.(MessageHook); ok {
				if err = r.Init(configFile); err != nil {
					return
				}
				hook = r
				return
			} else {
				err = ERR_HOOK_CREATE_FAILED.New(errors.Params{"type": hookType})
				return
			}
		}
		err = ERR_HOOK_BAD_DRIVER.New(errors.Params{"type": hookType})
		return
	}
}
