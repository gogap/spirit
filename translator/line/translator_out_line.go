package line

import (
	"bufio"
	"io"

	"github.com/gogap/spirit"
)

const (
	lineTranslatorOutURN = "urn:spirit:translator:out:line"
)

var _ spirit.OutputTranslator = new(LineOutputTranslator)

type LineOutputTranslator struct {
}

func init() {
	spirit.RegisterOutputTranslator(lineTranslatorOutURN, NewLineOutputTranslator)
}

func NewLineOutputTranslator(options spirit.Options) (translator spirit.OutputTranslator, err error) {
	translator = &LineOutputTranslator{}
	return
}

func (p *LineOutputTranslator) Out(w io.WriteCloser, delivery spirit.Delivery) (err error) {

	newWriter := bufio.NewWriter(w)
	var vData interface{}
	if vData, err = delivery.Payload().GetData(); err != nil {
		return
	}

	if data, ok := vData.(string); ok {
		if data[len(data)-1] != '\n' {
			data += "\n"
		}
		_, err = newWriter.Write([]byte(data))
	}

	err = newWriter.Flush()

	return
}
