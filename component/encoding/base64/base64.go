package base64

import (
	"encoding/base64"
	"errors"
	"sync"

	"github.com/gogap/spirit"
)

var _ spirit.Component = new(Base64Component)

const (
	base64URN = "urn:spirit:component:encoding:base64"
)

var (
	ErrDataTypeIsNotString = errors.New("payload data type is not string")
)

type Base64Component struct {
	statusLocker sync.Mutex
}

func init() {
	spirit.RegisterComponent(base64URN, NewBase64Component)
}

func NewBase64Component(options spirit.Options) (component spirit.Component, err error) {
	component = &Base64Component{}

	return
}

func (p *Base64Component) URN() string {
	return base64URN
}

func (p *Base64Component) Labels() spirit.Labels {
	return spirit.Labels{
		"version": "0.0.1",
	}
}

func (p *Base64Component) Handlers() spirit.Handlers {
	return spirit.Handlers{
		"encode": p.Encode,
		"decode": p.Decode,
	}
}

func (p *Base64Component) Encode(payload spirit.Payload) (result interface{}, err error) {
	var vData interface{}
	if vData, err = payload.GetData(); err != nil {
		return
	} else if data, ok := vData.(string); ok {
		result = base64.StdEncoding.EncodeToString([]byte(data))
	} else {
		err = ErrDataTypeIsNotString
		return
	}

	return
}

func (p *Base64Component) Decode(payload spirit.Payload) (result interface{}, err error) {
	var vData interface{}
	if vData, err = payload.GetData(); err != nil {
		return
	} else if data, ok := vData.(string); ok {
		if ret, e := base64.StdEncoding.DecodeString(data); e != nil {
			err = e
			return
		} else {
			result = string(ret)
		}
	} else {
		err = ErrDataTypeIsNotString
		return
	}

	return
}
