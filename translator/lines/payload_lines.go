package lines

import (
	"github.com/gogap/spirit"
)

var _ spirit.Payload = new(LinesPayload)

type LinesPayload struct {
	data string
}

func (p *LinesPayload) Id() (id string) {
	return
}

func (p *LinesPayload) GetData() (data interface{}, err error) {
	data = p.data
	return
}

func (p *LinesPayload) SetData(data interface{}) (err error) {
	p.data = data.(string)
	return
}

func (p *LinesPayload) GetError() (err error) {
	return
}

func (p *LinesPayload) SetError(err error) {
	return
}

func (p *LinesPayload) AppendMetadata(name string, values ...interface{}) (err error) {
	return
}

func (p *LinesPayload) GetMetadata(name string) (values []interface{}, exist bool) {
	return
}

func (p *LinesPayload) GetContext(name string) (v interface{}, exist bool) {
	return
}

func (p *LinesPayload) Metadata() (metadata spirit.Metadata) {
	return
}

func (p *LinesPayload) SetContext(name string, v interface{}) (err error) {
	return
}

func (p *LinesPayload) DeleteContext(name string) (err error) {
	return
}

func (p *LinesPayload) Context() (context spirit.Context) {
	return
}
