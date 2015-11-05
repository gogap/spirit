package json

import (
	"encoding/json"
	"io"

	"github.com/gogap/spirit"
)

const (
	jsonTranslatorOutURN = "urn:spirit:translator:out:json"
)

var _ spirit.OutputTranslator = new(JSONOutputTranslator)

type JSONOutputTranslator struct {
}

func init() {
	spirit.RegisterOutputTranslator(jsonTranslatorOutURN, NewJSONOutputTranslator)
}

func NewJSONOutputTranslator(options spirit.Options) (translator spirit.OutputTranslator, err error) {
	translator = &JSONOutputTranslator{}
	return
}

func (p *JSONOutputTranslator) Out(w io.WriteCloser, delivery spirit.Delivery) (err error) {

	var data interface{}
	if data, err = delivery.Payload().GetData(); err != nil {
		return
	}

	payload := _JSONPayload{
		Id:       delivery.Payload().Id(),
		Data:     data,
		Error:    delivery.Payload().GetError(),
		Metadata: delivery.Payload().Metadata(),
		Contexts: delivery.Payload().Contexts(),
	}

	jd := _JSONDelivery{
		Id:        delivery.Id(),
		URN:       delivery.URN(),
		SessionId: delivery.SessionId(),
		Payload:   payload,
		Timestamp: delivery.Timestamp(),
	}

	encoder := json.NewEncoder(w)
	err = encoder.Encode(jd)

	return
}
