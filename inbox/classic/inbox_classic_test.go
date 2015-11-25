package classic

import (
	"strings"
	"testing"

	"github.com/gogap/spirit"
	"github.com/gogap/spirit/io/pool"
	"github.com/gogap/spirit/io/std"
	"github.com/gogap/spirit/receiver/polling"
	"github.com/gogap/spirit/translator/lines"
)

func TestInboxReceive(t *testing.T) {

	opts := spirit.Map{"interval": 0, "buffer_size": 1, "timeout": 1000}

	var err error
	var receiver spirit.Receiver

	if receiver, err = polling.NewPollingReceiver(opts); err != nil {
		t.Error(err)
	}

	var readerPool spirit.ReaderPool
	if readerPool, err = pool.NewClassicReaderPool(spirit.Map{"max_size": 100}); err != nil {
		t.Error(err)
		return
	}

	readerConfig := spirit.Map{
		"name":  "ping",
		"proc":  "ping",
		"args":  []string{"-c", "1", "baidu.com"},
		"envs":  []string{},
		"delim": "\n",
	}

	if err = readerPool.SetNewReaderFunc(std.NewStdout, readerConfig); err != nil {
		t.Error(err)
		return
	}

	var translator spirit.InputTranslator
	if translator, err = lines.NewLinesInputTranslator(spirit.Map{}); err != nil {
		t.Errorf("create translator error, %s", err.Error())
		return
	}

	if recv, ok := receiver.(spirit.TranslatorReceiver); ok {
		if err = recv.SetTranslator(translator); err != nil {
			t.Errorf("set translator to receiver error, %s", err.Error())
			return
		}
	}

	if recv, ok := receiver.(spirit.ReadReceiver); ok {
		if err = recv.SetReaderPool(readerPool); err != nil {
			t.Errorf("set reader pool to receiver error, %s", err.Error())
			return
		}
	}

	if err != nil {
		t.Errorf("create polling receiver error, %s", err.Error())
		return
	}

	inboxOpts := spirit.Map{
		"size":        100,
		"put_timeout": 10000,
		"get_timeout": 10000,
	}

	var box spirit.Inbox
	if box, err = NewClassicInbox(inboxOpts); err != nil {
		t.Errorf("create new classic inbox error, %s", err.Error())
		return
	}

	receiver.SetDeliveryPutter(box)

	if err = box.Start(); err != nil {
		t.Errorf("start inbox error, %s", err.Error())
		return
	}

	var deliveries []spirit.Delivery

	if deliveries, err = box.Get(); err != nil {
		t.Errorf("get deliveries error, %s", err.Error())
		return
	}

	if len(deliveries) == 0 {
		t.Errorf("deliveries is empty")
		return
	}

	for _, delivery := range deliveries {
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
