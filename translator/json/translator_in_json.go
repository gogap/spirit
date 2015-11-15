package json

import (
	"encoding/json"
	"github.com/nu7hatch/gouuid"
	"io"
	"io/ioutil"
	"time"

	"github.com/gogap/spirit"
)

const (
	jsonTranslatorInURN = "urn:spirit:translator:in:json"
)

var _ spirit.InputTranslator = new(JSONInputTranslator)

type JSONInputTranslatorConfig struct {
	DataOnly bool          `json:"data_only"`
	Labels   spirit.Labels `json:"labels"`
	BindURN  string        `json:"bind_urn"`
}

type JSONInputTranslator struct {
	conf JSONInputTranslatorConfig
}

func init() {
	spirit.RegisterInputTranslator(jsonTranslatorInURN, NewJSONInputTranslator)
}

func NewJSONInputTranslator(config spirit.Config) (translator spirit.InputTranslator, err error) {
	conf := JSONInputTranslatorConfig{}

	if err = config.ToObject(&conf); err != nil {
		return
	}

	translator = &JSONInputTranslator{
		conf: conf,
	}
	return
}

func (p *JSONInputTranslator) inDataOnly(r io.Reader) (deliveries []spirit.Delivery, err error) {
	var data []byte

	if data, err = ioutil.ReadAll(r); err != nil {
		return
	}

	jp := JSONPayload{"data": data}

	deliverId, _ := uuid.NewV4()

	delivery := &JSONDelivery{
		urn:       p.conf.BindURN,
		id:        deliverId.String(),
		payload:   &jp,
		labels:    p.conf.Labels,
		timestamp: time.Now(),
	}

	deliveries = append(deliveries, delivery)

	return
}
func (p *JSONInputTranslator) inDeliveryData(r io.Reader) (deliveries []spirit.Delivery, err error) {
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

		if jd.Id == "" {
			deliverId, _ := uuid.NewV4()
			jd.Id = deliverId.String()
		}

		delivery := &JSONDelivery{
			urn:       jd.URN,
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

func (p *JSONInputTranslator) In(r io.Reader) (deliveries []spirit.Delivery, err error) {
	if p.conf.DataOnly {
		return p.inDataOnly(r)
	}
	return p.inDeliveryData(r)
}
