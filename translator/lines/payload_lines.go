package lines

import (
	"encoding/json"
	"github.com/gogap/spirit"
)

var _ spirit.Payload = new(LinesPayload)

type LinesPayload struct {
	data string
	errs []*spirit.Error
}

func (p *LinesPayload) Id() (id string) {
	return
}

func (p *LinesPayload) GetData() (data interface{}, err error) {
	if p.LastError() != nil {
		return p.LastError().Error(), nil
	}
	data = p.data
	return
}

func (p *LinesPayload) SetData(data interface{}) (err error) {
	p.data = data.(string)
	return
}

func (p *LinesPayload) DataToObject(v interface{}) (err error) {
	if err = json.Unmarshal([]byte(p.data), v); err != nil {
		return
	}
	return
}

func (p *LinesPayload) Errors() (err []*spirit.Error) {
	return p.errs
}

func (p *LinesPayload) AppendError(err ...*spirit.Error) {
	p.errs = append(p.errs, err...)
}

func (p *LinesPayload) LastError() (err *spirit.Error) {
	if len(p.errs) > 0 {
		err = p.errs[len(p.errs)-1]
	}
	return
}

func (p *LinesPayload) ClearErrors() {
	p.errs = nil
	return
}

func (p *LinesPayload) GetContext(name string) (v interface{}, exist bool) {
	return
}

func (p *LinesPayload) SetContext(name string, v interface{}) (err error) {
	return
}

func (p *LinesPayload) DeleteContext(name string) (err error) {
	return
}

func (p *LinesPayload) Context() (context spirit.Map) {
	return
}
