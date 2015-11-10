package classic

import (
	"strings"
	"testing"

	"github.com/gogap/spirit"

	"github.com/gogap/spirit/io/std"
	"github.com/gogap/spirit/receiver/polling"
	"github.com/gogap/spirit/translator/lines"
)

func TestInboxReceive(t *testing.T) {

	opts := spirit.Options{"interval": 0, "buffer_size": 1, "timeout": 1000}

	var err error
	var receiver spirit.Receiver

	receiver, err = polling.NewPollingReceiver(opts)

	readerOpts := spirit.Options{
		"name":  "ping",
		"proc":  "ping",
		"args":  []string{"-c", "10", "baidu.com"},
		"envs":  []string{},
		"delim": "\n",
	}

	receiver.SetNewReaderFunc(std.NewStdout, readerOpts)

	var translator spirit.InputTranslator
	if translator, err = lines.NewLinesInputTranslator(spirit.Options{}); err != nil {
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

	inboxOpts := spirit.Options{
		"size":        100,
		"put_timeout": -1,
		"get_timeout": -1,
	}

	var box spirit.Inbox
	if box, err = NewClassicInbox(inboxOpts); err != nil {
		t.Errorf("create new classic inbox error, %s", err.Error())
		return
	}

	if err = box.AddReceiver(receiver); err != nil {
		t.Errorf("add receiver to inbox error, %s", err.Error())
		return
	}

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
		t.Errorf("deliveries is empty, %s", err.Error())
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
