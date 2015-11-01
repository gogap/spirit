package test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gogap/spirit"
)

import (
	_ "github.com/gogap/spirit/component/util"
	_ "github.com/gogap/spirit/inbox/classic"
	_ "github.com/gogap/spirit/io/std"
	_ "github.com/gogap/spirit/matcher/classic"
	_ "github.com/gogap/spirit/outbox/classic"
	_ "github.com/gogap/spirit/receiver/polling"
	_ "github.com/gogap/spirit/router/classic"
	_ "github.com/gogap/spirit/sender/polling"
	_ "github.com/gogap/spirit/translator/line"
)

var (
	testJSONConf = `{
    "compose": [{
        "name": "test",
        "label_matchers": {
            "component": "test",
            "outbox": "test"
        },
        "components": [],
        "inboxes": [{
            "name": "test",
            "receivers": [{
                "name": "test",
                "translator": "test",
                "reader": "test"
            }]
        }],
        "outboxes": [{
            "name": "test",
            "senders": [{
                "name": "test",
                "translator": "test",
                "writer": "test"
            }]
        }]
    }],
    "readers": [{
        "name": "test",
        "urn": "urn:spirit:reader:std",
        "options": {
            "name": "ping",
            "proc": "ping",
            "args": ["-c", "10", "baidu.com"],
            "envs": {},
            "delim": "\n"
        }
    }],
    "writers": [{
        "name": "test",
        "urn": "urn:spirit:writer:std",
        "options": {
            "name": "write",
            "proc": "my-program-w",
            "args": [],
            "envs": [],
            "delim": "\n"
        }
    }],
    "input_translators": [{
        "name": "test",
        "urn": "urn:spirit:input-translator:line",
        "options": {
            "urn": "componentA:handlerA",
            "labels": {
                "a": "b"
            }
        }
    }],
    "output_translators": [{
        "name": "test",
        "urn": "urn:spirit:output-translator:line",
        "options": {}
    }],
    "receivers": [{
        "name": "test",
        "urn": "urn:spirit:receiver:polling",
        "options": {
            "interval": 0,
            "buffer_size": 1,
            "timeout": 5000
        }
    }],
    "inboxes": [{
        "name": "test",
        "urn": "urn:spirit:inbox:classic",
        "options": {
            "size": 100,
            "put_timeout": 100,
            "get_timeout": 100
        }
    }],
    "routers": [{
        "name": "test",
        "urn": "urn:spirit:router:classic",
        "options": {
            "interval": 0,
            "buffer_size": 1,
            "timeout": 1000
        }
    }],
    "outboxes": [{
        "name": "test",
        "urn": "urn:spirit:outbox:classic",
        "options": {
            "size": 100,
            "get_timeout": -1,
            "labels": {
                "a": "b"
            }
        }
    }],
    "label_matchers": [{
        "name": "test",
        "urn": "urn:spirit:label-matcher:equal",
        "options": {}
    }],
    "components": [{
        "name": "test",
        "urn": "urn:spirit:component:util:base64",
        "options": {}
    }],
    "senders": [{
        "name": "test",
        "urn": "urn:spirit:sender:polling",
        "options": {
            "interval": 0
        }
    }]
}`
)

func TestValidateSpiritConfig(t *testing.T) {

	spiritConf := spirit.SpiritConfig{}

	err := json.Unmarshal([]byte(testJSONConf), &spiritConf)

	if err != nil {
		t.Error(err)
		return
	}

	if err = spiritConf.Validate(); err != nil {
		t.Errorf("spirit config validate failed, %s", err)
		return
	}
}

func TestClassicSpiritBuild(t *testing.T) {

	spiritConf := spirit.SpiritConfig{}

	err := json.Unmarshal([]byte(testJSONConf), &spiritConf)

	if err != nil {
		t.Error(err)
		return
	}

	if err = spiritConf.Validate(); err != nil {
		t.Errorf("spirit config validate failed, %s", err)
		return
	}

	var sp spirit.Spirit
	if sp, err = spirit.NewClassicSpirit(); err != nil {
		t.Errorf("new classic spirit error, %s", err)
		return
	}

	if err = sp.Build(spiritConf); err != nil {
		t.Errorf("build classic spirit error, %s", err)
		return
	}

	if err = sp.Start(); err != nil {
		return
	}

	time.Sleep(time.Second * 10)

}
