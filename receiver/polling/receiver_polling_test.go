package polling

import (
	"strings"
	"sync"
	"testing"

	"github.com/gogap/spirit"

	"github.com/gogap/spirit/io/std"
	"github.com/gogap/spirit/translator/lines"
)

type _MockPutter struct {
	locker sync.Mutex

	deliveries []spirit.Delivery
}

func (p *_MockPutter) Put(deliveries []spirit.Delivery) (err error) {
	p.locker.Lock()
	defer p.locker.Unlock()

	p.deliveries = append(p.deliveries, deliveries...)

	return
}

func TestReceiverReceive(t *testing.T) {

	opts := spirit.Map{"interval": 0, "buffer_size": 1, "timeout": 5000}

	var err error
	var receiver spirit.Receiver

	receiver, err = NewPollingReceiver(opts)

	readerOpts := spirit.Map{
		"name":  "ping",
		"proc":  "ping",
		"args":  []string{"-c", "1", "baidu.com"},
		"envs":  []string{},
		"delim": "\n",
	}

	receiver.SetNewReaderFunc(std.NewStdout, readerOpts)

	var translator spirit.InputTranslator
	if translator, err = lines.NewLinesInputTranslator(spirit.Map{}); err != nil {
		t.Errorf("create translator error, %s", err.Error())
		return
	}

	if err = receiver.SetTranslator(translator); err != nil {
		t.Errorf("set translator to receiver error, %s", err.Error())
		return
	}

	if err != nil {
		t.Errorf("create polling receiver error, %s", err.Error())
		return
	}

	putter := _MockPutter{}

	if err = receiver.SetDeliveryPutter(&putter); err != nil {
		t.Errorf("set delivery putter error, %s", err.Error())
		return
	}

	if err = receiver.Start(); err != nil {
		t.Errorf("start receiver error, %s", err.Error())
		return
	}

	for _, delivery := range putter.deliveries {
		if data, e := delivery.Payload().GetData(); e != nil {
			t.Errorf("get delivery payload error, %s", e.Error())
			return
		} else if strData, ok := data.(string); !ok {
			t.Errorf("payload data is not string type")
			return
		} else if !strings.Contains(strData, "baidu.com") &&
			!strings.Contains(strData, "icmp") {
			t.Errorf("payload data content error")
			return
		}
	}
}
