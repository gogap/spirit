package json

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/gogap/spirit"
)

var _ spirit.Component = new(JSONComponent)

const (
	jsonURN = "urn:spirit:component:encoding:json"
)

var (
	ErrDataTypeIsNotString = errors.New("payload data type is not string")
)

type JSONComponent struct {
	name         string
	statusLocker sync.Mutex
}

func init() {
	spirit.RegisterComponent(jsonURN, NewJSONComponent)
}

func NewJSONComponent(name string, options spirit.Map) (component spirit.Component, err error) {
	component = &JSONComponent{
		name: name,
	}

	return
}

func (p *JSONComponent) Name() string {
	return p.name
}

func (p *JSONComponent) URN() string {
	return jsonURN
}

func (p *JSONComponent) Labels() spirit.Labels {
	return spirit.Labels{
		"version": "0.0.1",
	}
}

func (p *JSONComponent) Handlers() spirit.Handlers {
	return spirit.Handlers{
		"marshal":   p.Marshal,
		"unmarshal": p.Unmarshal,
	}
}

func (p *JSONComponent) Marshal(payload spirit.Payload) (result interface{}, err error) {
	var vData interface{}
	if vData, err = payload.GetData(); err != nil {
		return
	}

	if vData != nil {
		var data []byte
		if data, err = json.Marshal(vData); err != nil {
			return
		}

		result = data
	}

	return
}

func (p *JSONComponent) Unmarshal(payload spirit.Payload) (result interface{}, err error) {
	var vData interface{}
	var byteData []byte

	if vData, err = payload.GetData(); err != nil {
		return
	} else if data, ok := vData.(string); ok {
		byteData = []byte(data)
	} else if data, ok := vData.([]byte); ok {
		byteData = data
	} else {
		err = ErrDataTypeIsNotString
		return
	}

	var jsonVal map[string]interface{}
	if err = json.Unmarshal(byteData, &jsonVal); err != nil {
		return
	}
	result = jsonVal

	return
}
