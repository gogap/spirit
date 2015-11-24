package base64

import (
	"testing"

	"github.com/gogap/spirit"
)

type _MockPayload struct {
	data string
}

func (p *_MockPayload) Id() (id string) {
	return
}
func (p *_MockPayload) GetData() (data interface{}, err error) {
	return p.data, nil
}
func (p *_MockPayload) SetData(data interface{}) (err error) {
	return
}
func (p *_MockPayload) GetError() (err error) {
	return
}
func (p *_MockPayload) SetError(err error) {
	return
}
func (p *_MockPayload) AppendMetadata(name string, values ...interface{}) (err error) {
	return
}
func (p *_MockPayload) GetMetadata(name string) (values []interface{}, exist bool) {
	return
}
func (p *_MockPayload) Metadata() (metadata spirit.Metadata) {
	return
}
func (p *_MockPayload) GetContext(name string) (v interface{}, exist bool) {
	return
}
func (p *_MockPayload) SetContext(name string, v interface{}) (err error) {
	return
}
func (p *_MockPayload) Context() (context spirit.Context) {
	return
}
func (p *_MockPayload) DeleteContext(name string) (err error) {
	return
}

func TestBase64Encode(t *testing.T) {
	component, err := NewBase64Component(spirit.Config{})
	if err != nil {
		t.Error("create component error")
		return
	}

	handler, exist := component.Handlers()["encode"]
	if !exist {
		t.Error("handler of encode not exist")
		return
	}

	var ret interface{}
	ret, err = handler(&_MockPayload{data: "hello world"})
	if err != nil {
		t.Error(err)
		return
	}

	if ret != "aGVsbG8gd29ybGQ=" {
		t.Errorf("base64 encode result %s!=%s", ret, "aGVsbG8gd29ybGQ=")
	}
}

func TestBase64Decode(t *testing.T) {
	component, err := NewBase64Component(spirit.Config{})
	if err != nil {
		t.Error("create component error")
		return
	}

	handler, exist := component.Handlers()["decode"]
	if !exist {
		t.Error("handler of decode not exist")
		return
	}

	var ret interface{}
	ret, err = handler(&_MockPayload{data: "aGVsbG8gd29ybGQ="})
	if err != nil {
		t.Error(err)
		return
	}

	if ret != "hello world" {
		t.Errorf("base64 encode result %s!=%s", ret, "hello world")
	}
}
