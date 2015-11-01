package xml

import (
	"encoding/xml"

	"github.com/gogap/spirit"
)

var _ spirit.Payload = new(XMLPayload)

type _Metadata map[string][]interface{}
type _Contexts map[string]interface{}

type _XMLPayload struct {
	XMLName  xml.Name  `xml:"Payload"`
	Id       string    `xml:"Id"`
	Data     string    `xml:"Data"`
	Error    error     `xml:"Error"`
	Metadata _Metadata `xml:"Metadata"`
	Contexts _Contexts `xml:"Contexts"`
}

type XMLPayload map[string]interface{}

func (p XMLPayload) Id() (id string) {
	v, _ := p["id"]
	id = v.(string)
	return
}

func (p XMLPayload) GetData() (data interface{}, err error) {
	data, _ = p["data"]
	return
}

func (p XMLPayload) SetData(data interface{}) (err error) {
	p["data"] = data
	return
}

func (p XMLPayload) GetError() (err error) {
	v, _ := p["error"]
	err = v.(error)
	return
}

func (p XMLPayload) SetError(err error) {
	p["error"] = err
	return
}

func (p XMLPayload) AppendMetadata(name string, values ...interface{}) (err error) {
	return
}

func (p XMLPayload) GetMetadata(name string) (values []interface{}, exist bool) {
	var metadata spirit.Metadata
	var ok bool
	if v, ex := p["metadata"]; !ex {
		return
	} else if metadata, ok = v.(spirit.Metadata); ok {
		values, exist = metadata[name]
	}

	return
}

func (p XMLPayload) GetContext(name string) (v interface{}, exist bool) {
	var contexts spirit.Contexts
	var ok bool
	if vContext, ex := p["contexts"]; !ex {
		return
	} else if contexts, ok = vContext.(spirit.Contexts); ok {
		v, exist = contexts[name]
	}

	return
}

func (p XMLPayload) Metadata() (metadata spirit.Metadata) {
	return
}

func (p XMLPayload) SetContext(name string, v interface{}) (err error) {
	return
}

func (p XMLPayload) DeleteContext(name string) (err error) {
	return
}

func (p XMLPayload) Contexts() (contexts spirit.Contexts) {
	return
}
