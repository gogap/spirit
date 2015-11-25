package polling

import (
	"sync"
	"testing"
	"time"

	"github.com/gogap/spirit"

	"github.com/gogap/spirit/io/std"
	"github.com/gogap/spirit/translator/lines"
)

func TestSenderSend(t *testing.T) {

	opts := spirit.Map{"interval": 0, "disable_session": true}

	var err error
	var sender spirit.Sender

	sender, err = NewPollingSender(opts)

	optsW := spirit.Map{
		"name":  "write",
		"proc":  "my-program-w",
		"args":  []string{},
		"envs":  []string{},
		"delim": "\n",
	}

	sender.SetNewWriterFunc(std.NewStdin, optsW)

	var translator spirit.OutputTranslator
	if translator, err = lines.NewLinesOutputTranslator(spirit.Map{}); err != nil {
		t.Errorf("create translator error, %s", err.Error())
		return
	}

	if err = sender.SetTranslator(translator); err != nil {
		t.Errorf("set translator to sender error, %s", err.Error())
		return
	}

	if err != nil {
		t.Errorf("create polling sender error, %s", err.Error())
		return
	}

	payload1 := _MockPayload{data: "payload_1 hello\n"}
	payload2 := _MockPayload{data: "payload_2 hello\n"}

	delivery1 := _MockDelivery{payload: &payload1}
	delivery2 := _MockDelivery{payload: &payload2}

	getter := _MockGetter{
		deliveries: []spirit.Delivery{&delivery1, &delivery2},
	}

	if err = sender.SetDeliveryGetter(&getter); err != nil {
		t.Errorf("set delivery getter error, %s", err.Error())
		return
	}

	if err = sender.Start(); err != nil {
		t.Errorf("start sender error, %s", err.Error())
		return
	}

	time.Sleep(time.Second)

}

// Mocks

type _MockGetter struct {
	locker sync.Mutex

	deliveries []spirit.Delivery
}

func (p *_MockGetter) Get() (deliveries []spirit.Delivery, err error) {
	return p.deliveries, nil
}

type _MockDelivery struct {
	payload *_MockPayload
}

func (p *_MockDelivery) Id() string {
	return ""
}

func (p *_MockDelivery) URN() string {
	return ""
}

func (p *_MockDelivery) SessionId() string {
	return ""
}

func (p *_MockDelivery) Payload() spirit.Payload {
	return p.payload
}

func (p *_MockDelivery) Labels() spirit.Labels {
	return nil
}

func (p *_MockDelivery) Validate() (err error) {
	return
}

func (p *_MockDelivery) Timestamp() time.Time {
	return time.Now()
}

type _MockPayload struct {
	data string
}

func (p *_MockPayload) Id() (id string) {
	return
}

func (p *_MockPayload) GetData() (data interface{}, err error) {
	data = p.data
	return
}

func (p *_MockPayload) SetData(data interface{}) (err error) {
	p.data = data.(string)
	return
}

func (p *_MockPayload) SetError(err error) {
	return
}

func (p *_MockPayload) GetError() (err error) {
	return
}

func (p *_MockPayload) AppendMetadata(name string, values ...interface{}) (err error) {
	return
}

func (p *_MockPayload) GetMetadata(name string) (values []interface{}, exist bool) {
	return
}

func (p *_MockPayload) GetContext(name string) (v interface{}, exist bool) {
	return
}

func (p *_MockPayload) Metadata() (metadata spirit.Metadata) {
	return
}

func (p *_MockPayload) SetContext(name string, v interface{}) (err error) {
	return
}

func (p *_MockPayload) DeleteContext(name string) (err error) {
	return
}

func (p *_MockPayload) Context() (context spirit.Context) {
	return
}
