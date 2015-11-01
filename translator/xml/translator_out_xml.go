package xml

import (
	"encoding/xml"
	"io"

	"github.com/gogap/spirit"
)

const (
	xmlTranslatorOutURN = "urn:spirit:output-translator:xml"
)

var _ spirit.OutputTranslator = new(XMLOutputTranslator)

type XMLOutputTranslator struct {
}

func init() {
	spirit.RegisterOutputTranslator(xmlTranslatorOutURN, NewXMLOutputTranslator)
}

func NewXMLOutputTranslator(options spirit.Options) (translator spirit.OutputTranslator, err error) {
	return
}

func (p *XMLOutputTranslator) Out(w io.WriteCloser, delivery spirit.Delivery) (err error) {
	encoder := xml.NewEncoder(w)
	err = encoder.Encode(delivery)
	return
}
