package xml

import (
	"encoding/xml"
	"io"

	"github.com/gogap/spirit"
)

const (
	xmlTranslatorInURN = "urn:spirit:input-translator:xml"
)

var _ spirit.InputTranslator = new(XMLInputTranslator)

type XMLInputTranslatorConfig struct {
	Labels spirit.Labels `json:"labels"`
}

type XMLInputTranslator struct {
	conf XMLInputTranslatorConfig
}

func init() {
	spirit.RegisterInputTranslator(xmlTranslatorInURN, NewXMLInputTranslator)
}

func NewXMLInputTranslator(options spirit.Options) (translator spirit.InputTranslator, err error) {
	conf := XMLInputTranslatorConfig{}

	if err = options.ToObject(&conf); err != nil {
		return
	}
	translator = &XMLInputTranslator{conf: conf}
	return
}

func (p *XMLInputTranslator) In(r io.Reader) (deliveries []spirit.Delivery, err error) {
	decoder := xml.NewDecoder(r)

	var tmpDeliveries []spirit.Delivery
	for err == nil {
		jd := _XMLDelivery{}

		if err = decoder.Decode(&jd); err != nil {
			if err == io.EOF {
				err = nil
				deliveries = tmpDeliveries
			}
			return
		}

		jp := XMLPayload{}

		jp["id"] = jd.Payload.Id
		jp["data"] = jd.Payload.Data
		jp["error"] = jd.Payload.Error
		jp["metadata"] = jd.Payload.Metadata
		jp["contexts"] = jd.Payload.Contexts

		labels := spirit.Labels{}
		for k, v := range p.conf.Labels {
			labels[k] = v
		}

		delivery := &XMLDelivery{
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
