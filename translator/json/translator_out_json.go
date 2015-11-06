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

type JSONOutputTranslatorConfig struct {
	DataOnly bool `json:"data_only"`
}

type JSONOutputTranslator struct {
	conf JSONOutputTranslatorConfig
}

func init() {
	spirit.RegisterOutputTranslator(jsonTranslatorOutURN, NewJSONOutputTranslator)
}

func NewJSONOutputTranslator(options spirit.Options) (translator spirit.OutputTranslator, err error) {
	conf := JSONOutputTranslatorConfig{}

	if err = options.ToObject(&conf); err != nil {
		return
	}

	translator = &JSONOutputTranslator{
		conf: conf,
	}
	return
}

func (p *JSONOutputTranslator) outDataOnly(w io.WriteCloser, delivery spirit.Delivery) (err error) {

	var data interface{}
	if data, err = delivery.Payload().GetData(); err != nil {
		return
	}

	switch d := data.(type) {
	case []byte:
		{
			_, err = w.Write(d)
		}
	case string:
		{
			_, err = w.Write([]byte(d))
		}
	default:
		encode := json.NewEncoder(w)
		err = encode.Encode(data)
	}

	return
}

func (p *JSONOutputTranslator) outDeliveryData(w io.WriteCloser, delivery spirit.Delivery) (err error) {

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

func (p *JSONOutputTranslator) Out(w io.WriteCloser, delivery spirit.Delivery) (err error) {

	if p.conf.DataOnly {
		return p.outDataOnly(w, delivery)
	}

	return p.outDeliveryData(w, delivery)
}
