package json

import (
	"github.com/gogap/spirit"
)

var _ spirit.Payload = new(JSONPayload)

type _JSONPayload struct {
	Id       string          `json:"id"`
	Data     interface{}     `json:"data"`
	Error    error           `json:"error"`
	Metadata spirit.Metadata `json:"metadata"`
	Contexts spirit.Contexts `json:"contexts"`
}

type JSONPayload map[string]interface{}

func (p JSONPayload) Id() (id string) {
	v, _ := p["id"]
	if v == nil {
		id = ""
	} else {
		id = v.(string)
	}
	return
}

func (p JSONPayload) GetData() (data interface{}, err error) {
	data, _ = p["data"]
	return
}

func (p JSONPayload) SetData(data interface{}) (err error) {
	p["data"] = data
	return
}

func (p JSONPayload) GetError() (err error) {
	v, _ := p["error"]
	if v == nil {
		return
	}
	err = v.(error)
	return
}

func (p JSONPayload) SetError(err error) {
	p["error"] = err
	return
}

func (p JSONPayload) AppendMetadata(name string, values ...interface{}) (err error) {
	return
}

func (p JSONPayload) GetMetadata(name string) (values []interface{}, exist bool) {
	var metadata spirit.Metadata
	var ok bool
	if v, ex := p["metadata"]; !ex {
		return
	} else if metadata, ok = v.(spirit.Metadata); ok {
		values, exist = metadata[name]
	}

	return
}

func (p JSONPayload) GetContext(name string) (v interface{}, exist bool) {
	var contexts spirit.Contexts
	var ok bool
	if vContext, ex := p["contexts"]; !ex {
		return
	} else if contexts, ok = vContext.(spirit.Contexts); ok {
		v, exist = contexts[name]
	}

	return
}

func (p JSONPayload) Metadata() (metadata spirit.Metadata) {
	return
}

func (p JSONPayload) SetContext(name string, v interface{}) (err error) {
	return
}

func (p JSONPayload) DeleteContext(name string) (err error) {
	return
}

func (p JSONPayload) Contexts() (contexts spirit.Contexts) {
	return
}
