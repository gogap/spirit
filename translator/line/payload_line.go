package line

import (
	"github.com/gogap/spirit"
)

var _ spirit.Payload = new(LinePayload)

type LinePayload struct {
	data string
}

func (p *LinePayload) Id() (id string) {
	return
}

func (p *LinePayload) GetData() (data interface{}, err error) {
	data = p.data
	return
}

func (p *LinePayload) SetData(data interface{}) (err error) {
	p.data = data.(string)
	return
}

func (p *LinePayload) GetError() (err error) {
	return
}

func (p *LinePayload) SetError(err error) {
	return
}

func (p *LinePayload) AppendMetadata(name string, values ...interface{}) (err error) {
	return
}

func (p *LinePayload) GetMetadata(name string) (values []interface{}, exist bool) {
	return
}

func (p *LinePayload) GetContext(name string) (v interface{}, exist bool) {
	return
}

func (p *LinePayload) Metadata() (metadata spirit.Metadata) {
	return
}

func (p *LinePayload) SetContext(name string, v interface{}) (err error) {
	return
}

func (p *LinePayload) DeleteContext(name string) (err error) {
	return
}

func (p *LinePayload) Contexts() (contexts spirit.Contexts) {
	return
}
