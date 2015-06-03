package spirit

import (
	"fmt"
	"reflect"

	"github.com/gogap/errors"
)

type MessageReceiverFactory interface {
	RegisterMessageReceivers(receivers ...MessageReceiver)
	IsExist(receiverType string) bool
	NewReceiver(receiverType, url string, options Options) (receiver MessageReceiver, err error)
}

type DefaultMessageReceiverFactory struct {
	receiverDrivers map[string]reflect.Type
}

func NewDefaultMessageReceiverFactory() MessageReceiverFactory {
	fact := new(DefaultMessageReceiverFactory)
	fact.receiverDrivers = make(map[string]reflect.Type)
	return fact
}

func (p *DefaultMessageReceiverFactory) RegisterMessageReceivers(receivers ...MessageReceiver) {
	if receivers == nil {
		panic("receivers is nil")
	}

	if len(receivers) == 0 {
		return
	}

	for _, receiver := range receivers {
		if _, exist := p.receiverDrivers[receiver.Type()]; exist {
			panic(fmt.Errorf("receiver driver of [%s] already exist", receiver.Type()))
		}

		vof := reflect.ValueOf(receiver)
		vType := vof.Type()
		if vof.Kind() == reflect.Ptr {
			vType = vof.Elem().Type()
		}

		p.receiverDrivers[receiver.Type()] = vType
	}
	return
}

func (p *DefaultMessageReceiverFactory) IsExist(receiverType string) bool {
	if _, exist := p.receiverDrivers[receiverType]; exist {
		return true
	}
	return false
}

func (p *DefaultMessageReceiverFactory) NewReceiver(receiverType, url string, options Options) (receiver MessageReceiver, err error) {
	if receiverType, exist := p.receiverDrivers[receiverType]; !exist {
		err = ERR_RECEIVER_DRIVER_NOT_EXIST.New(errors.Params{"type": receiverType})
		return
	} else {
		if vOfMessageReceiver := reflect.New(receiverType); vOfMessageReceiver.CanInterface() {
			iMessageReceiver := vOfMessageReceiver.Interface()
			if r, ok := iMessageReceiver.(MessageReceiver); ok {
				if err = r.Init(url, options); err != nil {
					return
				}
				receiver = r
				return
			} else {
				err = ERR_RECEIVER_CREATE_FAILED.New(errors.Params{"type": receiverType, "url": url})
				return
			}
		}
		err = ERR_RECEIVER_BAD_DRIVER.New(errors.Params{"type": receiverType})
		return
	}
}
