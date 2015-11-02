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

type LineOutputTranslatorConfig struct {
	Delim string `json:"delim"`
}

type LineOutputTranslator struct {
	conf LineOutputTranslatorConfig
}

func init() {
	spirit.RegisterOutputTranslator(lineTranslatorOutURN, NewLineOutputTranslator)
}

func NewLineOutputTranslator(options spirit.Options) (translator spirit.OutputTranslator, err error) {
	conf := LineOutputTranslatorConfig{}

	if err = options.ToObject(&conf); err != nil {
		return
	}

	translator = &LineOutputTranslator{
		conf: conf,
	}
	return
}

func (p *LineOutputTranslator) Out(w io.WriteCloser, delivery spirit.Delivery) (err error) {

	newWriter := bufio.NewWriter(w)
	var vData interface{}
	if vData, err = delivery.Payload().GetData(); err != nil {
		return
	}

	var delim byte = '\n'
	if len(p.conf.Delim) == 1 {
		delim = p.conf.Delim[0]
	}

	if data, ok := vData.(string); ok {
		if data[len(data)-1] != delim {
			data += string(delim)
		}
		_, err = newWriter.Write([]byte(data))
	}

	err = newWriter.Flush()

	return
}
