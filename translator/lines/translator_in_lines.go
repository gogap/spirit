package lines

import (
	"bufio"
	"errors"
	"io"

	"github.com/gogap/spirit"
)

const (
	linesTranslatorInURN = "urn:spirit:translator:in:lines"
)

var _ spirit.InputTranslator = new(LinesInputTranslator)

var ErrLinesInputTranslatorNeedDefaultURN = errors.New("lines input translator need default urn")

type LinesInputTranslatorConfig struct {
	BindURN string        `json:"bind_urn"`
	Labels  spirit.Labels `json:"labels"`
	Delim   string        `json:"delim"`
}

type LinesInputTranslator struct {
	conf LinesInputTranslatorConfig
}

func init() {
	spirit.RegisterInputTranslator(linesTranslatorInURN, NewLinesInputTranslator)
}

func NewLinesInputTranslator(config spirit.Map) (translator spirit.InputTranslator, err error) {
	conf := LinesInputTranslatorConfig{}

	if err = config.ToObject(&conf); err != nil {
		return
	}

	if conf.BindURN == "" {
		err = ErrLinesInputTranslatorNeedDefaultURN
		return
	}

	translator = &LinesInputTranslator{
		conf: conf,
	}
	return
}

func (p *LinesInputTranslator) In(r io.Reader) (deliveries []spirit.Delivery, err error) {
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

	delivery := &LinesDelivery{
		urn:    p.conf.BindURN,
		labels: labels,
		payload: &LinesPayload{
			data: txt,
		},
	}

	if err = delivery.Validate(); err != nil {
		return
	}

	deliveries = append(deliveries, delivery)

	return
}
