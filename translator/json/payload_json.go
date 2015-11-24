package json

import (
	"encoding/json"
	"strings"

	"github.com/gogap/spirit"
)

var _ spirit.Payload = new(JSONPayload)

type _JSONPayload struct {
	Id       string          `json:"id"`
	Data     interface{}     `json:"data"`
	Errors   []*spirit.Error `json:"error"`
	Metadata spirit.Metadata `json:"metadata"`
	Contexts spirit.Contexts `json:"contexts"`
}

type JSONPayload struct {
	id       string
	errs     []*spirit.Error
	data     interface{}
	metadata spirit.Metadata
	contexts spirit.Contexts
}

func NewJSONPayload() *JSONPayload {
	return &JSONPayload{
		metadata: make(spirit.Metadata),
		contexts: make(spirit.Contexts),
	}
}

func (p *JSONPayload) Id() (id string) {
	return p.id
}

func (p *JSONPayload) Errors() (err []*spirit.Error) {
	return p.errs
}

func (p *JSONPayload) AppendError(err ...*spirit.Error) {
	p.errs = append(p.errs, err...)
}

func (p *JSONPayload) LastError() (err *spirit.Error) {
	if len(p.errs) > 0 {
		err = p.errs[len(p.errs)-1]
	}
	return
}

func (p *JSONPayload) ClearErrors() {
	p.errs = nil
	return
}
func (p *JSONPayload) GetData() (data interface{}, err error) {
	data = p.data
	return
}

func (p *JSONPayload) SetData(data interface{}) (err error) {
	p.data = data
	return
}

func (p *JSONPayload) DataToObject(v interface{}) (err error) {

	if p.data == nil {
		return
	}

	switch d := p.data.(type) {
	case string:
		{
			if err = json.Unmarshal([]byte(d), v); err != nil {
				var b []byte
				if b, err = json.Marshal(d); err != nil {
					return
				}

				if err = json.Unmarshal(b, v); err != nil {
					return
				}
			}
		}
	case []byte:
		{
			if err = json.Unmarshal(d, v); err != nil {
				var b []byte
				if b, err = json.Marshal(string(d)); err != nil {
					return
				}

				if err = json.Unmarshal(b, v); err != nil {
					return
				}
			}
		}
	default:
		{
			var b []byte
			if b, err = json.Marshal(p.data); err != nil {
				return
			}

			if err = json.Unmarshal(b, v); err != nil {
				return
			}
		}
	}
	return
}

func (p *JSONPayload) GetContext(name string) (v interface{}, exist bool) {
	v, exist = p.contexts[strings.ToUpper(name)]
	return
}

func (p *JSONPayload) SetContext(name string, v interface{}) (err error) {
	p.contexts[strings.ToUpper(name)] = v
	return
}

func (p *JSONPayload) Contexts() (contexts spirit.Contexts) {
	return p.contexts
}

func (p *JSONPayload) DeleteContext(name string) (err error) {
	delete(p.contexts, strings.ToUpper(name))
	return
}
