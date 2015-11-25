package lines

import (
	"bytes"
	"testing"

	"github.com/gogap/spirit"
)

func TestTranslateRead(t *testing.T) {
	trans, err := NewLinesInputTranslator(spirit.Map{"urn": "componentA:handlerA", "labels": map[string]string{"a": "b"}})

	if err != nil {
		t.Error(err)
		return
	}

	testValue := "hello\n"

	buf := bytes.NewBufferString(testValue)

	deliveries, err := trans.In(buf)

	if err != nil {
		t.Error(err)
		return
	}

	delivery := deliveries[0]

	if delivery.Labels()["a"] != "b" {
		t.Errorf("delivery labels are not correct")
		return
	}

	var vData interface{}

	vData, err = delivery.Payload().GetData()

	if data, ok := vData.(string); !ok {
		t.Error("payload data is not string type")
	} else if data != testValue {
		t.Errorf("the data value is not '%s', it is %s", testValue, data)
		return
	}
}
