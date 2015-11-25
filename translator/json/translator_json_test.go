package json

import (
	"bytes"
	"testing"

	"github.com/gogap/spirit"
)

func TestTranslateRead(t *testing.T) {
	trans, err := NewJSONInputTranslator(spirit.Map{"labels": map[string]string{"a": "b"}})

	deliveryJSON := `{"payload": {"data":"hello"}}{"payload": {"data":"hello2"}}`
	testValue1 := "hello"
	testValue2 := "hello2"

	buf := bytes.NewBufferString(deliveryJSON)

	deliveries, err := trans.In(buf)

	if err != nil {
		t.Error(err)
		return
	}

	if len(deliveries) != 2 {
		t.Errorf("delivery count 2!=%d", len(deliveries))
		return
	}

	delivery1 := deliveries[0]
	delivery2 := deliveries[1]

	if delivery1.Labels()["a"] != "b" {
		t.Errorf("delivery labels are not correct")
		return
	}

	if delivery2.Labels()["a"] != "b" {
		t.Errorf("delivery labels are not correct")
		return
	}

	if err != nil {
		t.Error(err)
	}

	var vData1, vData2 interface{}

	vData1, err = delivery1.Payload().GetData()

	if data, ok := vData1.(string); !ok {
		t.Error("payload data is not string type")
	} else if data != testValue1 {
		t.Errorf("the data value is not '%s', it is %s", testValue1, data)
		return
	}

	vData2, err = delivery2.Payload().GetData()

	if data, ok := vData2.(string); !ok {
		t.Error("payload data is not string type")
	} else if data != testValue2 {
		t.Errorf("the data value is not '%s', it is %s", testValue2, data)
		return
	}
}
