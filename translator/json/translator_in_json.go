package json

import (
	"encoding/json"
	"io"

	"github.com/gogap/spirit"
)

const (
	jsonTranslatorInURN = "urn:spirit:input-translator:json"
)

var _ spirit.InputTranslator = new(JSONInputTranslator)

type JSONInputTranslatorConfig struct {
	Labels spirit.Labels `json:"labels"`
}

type JSONInputTranslator struct {
	conf JSONInputTranslatorConfig
}

func init() {
	spirit.RegisterInputTranslator(jsonTranslatorURN, NewJSONInputTranslator)
}

func NewJSONInputTranslator(options spirit.Options) (translator spirit.InputTranslator, err error) {
	conf := JSONInputTranslatorConfig{}

	if err = options.ToObject(&conf); err != nil {
		return
	}

	translator = &JSONInputTranslator{
		conf: conf,
	}
	return
}

func (p *JSONInputTranslator) In(r io.Reader) (deliveries []spirit.Delivery, err error) {
	decoder := json.NewDecoder(r)

	var tmpDeliveries []spirit.Delivery
	for err == nil {
		jd := _JSONDelivery{}

		if err = decoder.Decode(&jd); err != nil {
			if err == io.EOF {
				err = nil
				deliveries = tmpDeliveries
			}
			return
		}

		jp := JSONPayload{}

		jp["id"] = jd.Payload.Id
		jp["data"] = jd.Payload.Data
		jp["error"] = jd.Payload.Error
		jp["metadata"] = jd.Payload.Metadata
		jp["contexts"] = jd.Payload.Contexts

		labels := spirit.Labels{}
		for k, v := range p.conf.Labels {
			labels[k] = v
		}

		delivery := &JSONDelivery{
			id:        jd.Id,
			payload:   &jp,
			labels:    labels,
			timestamp: jd.Timestamp,
		}

		if err = delivery.Validate(); err != nil {
			return
		}

		tmpDeliveries = append(tmpDeliveries, delivery)
	}

	return
}
