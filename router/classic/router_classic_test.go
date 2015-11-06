package classic

import (
	"testing"

	"github.com/gogap/spirit"

	classicInbox "github.com/gogap/spirit/inbox/classic"
	"github.com/gogap/spirit/io/pool"
	"github.com/gogap/spirit/io/std"
	"github.com/gogap/spirit/receiver/polling"
	"github.com/gogap/spirit/translator/line"

	"github.com/gogap/spirit/component/encoding/base64"
)

func TestRouteToHandler(t *testing.T) {

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

	readerPool, _ := pool.NewClassicReaderPool(nil)

	readerPool.SetNewReaderFunc(std.NewStdout, readerOpts)

	var translator spirit.InputTranslator
	if translator, err = line.NewLineInputTranslator(spirit.Options{"bind_urn": "test@urn:spirit:component:encoding:base64#encode"}); err != nil {
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
	if box, err = classicInbox.NewClassicInbox(inboxOpts); err != nil {
		t.Errorf("create new classic inbox error, %s", err.Error())
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

	var router spirit.Router
	if router, err = NewClassicRouter(spirit.Options{}); err != nil {
		t.Errorf("create router error: %s", err.Error())
		return
	}

	var component spirit.Component
	if component, err = base64.NewBase64Component(spirit.Options{}); err != nil {
		t.Errorf("create base64 component error: %s", err.Error())
		return
	}

	router.AddComponent("test", component)

	for _, delivery := range deliveries {
		var handlers []spirit.ComponentHandler
		if handlers, err = router.RouteToHandlers(delivery); err != nil {
			t.Errorf("route to handler failed: %s", err.Error())
			return
		} else if handlers == nil || len(handlers) != 1 {
			t.Errorf("handler geted, but is nil: %s", err.Error())
			return
		}
	}

}
