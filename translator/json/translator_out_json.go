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
	return
}

func (p *JSONOutputTranslator) Out(w io.WriteCloser, delivery spirit.Delivery) (err error) {
	encoder := json.NewEncoder(w)
	err = encoder.Encode(delivery)
	return
}
