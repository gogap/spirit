package line

import (
	"bufio"
	"errors"
	"io"

	"github.com/gogap/spirit"
)

const (
	lineTranslatorInURN = "urn:spirit:translator:in:line"
)

var _ spirit.InputTranslator = new(LineInputTranslator)

var ErrLineInputTranslatorNeedDefaultURN = errors.New("line input translator need default urn")

type LineInputTranslatorConfig struct {
	BindURN string        `json:"bind_urn"`
	Labels  spirit.Labels `json:"labels"`
	Delim   string        `json:"delim"`
}

type LineInputTranslator struct {
	conf LineInputTranslatorConfig
}

func init() {
	spirit.RegisterInputTranslator(lineTranslatorInURN, NewLineInputTranslator)
}

func NewLineInputTranslator(options spirit.Options) (translator spirit.InputTranslator, err error) {
	conf := LineInputTranslatorConfig{}

	if err = options.ToObject(&conf); err != nil {
		return
	}

	if conf.BindURN == "" {
		err = ErrLineInputTranslatorNeedDefaultURN
		return
	}

	translator = &LineInputTranslator{
		conf: conf,
	}
	return
}

func (p *LineInputTranslator) In(r io.Reader) (deliveries []spirit.Delivery, err error) {
	reader := bufio.NewReader(r)

	txt := ""

	var delim byte = '\n'
	if len(p.conf.Delim) == 1 {
		delim = p.conf.Delim[0]
	}

	if txt, err = reader.ReadString(delim); err != nil {
		return
	}

	labels := spirit.Labels{}
	for k, v := range p.conf.Labels {
		labels[k] = v
	}

	delivery := &LineDelivery{
		urn:    p.conf.BindURN,
		labels: labels,
		payload: &LinePayload{
			data: txt,
		},
	}

	if err = delivery.Validate(); err != nil {
		return
	}

	deliveries = append(deliveries, delivery)

	return
}
