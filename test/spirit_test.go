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
	_ "github.com/gogap/spirit/io/pool"
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
        "components": ["std_base64"],
        "inboxes": [{
            "name": "test",
            "receivers": [{
                "name": "test",
                "translator": "test",
                "reader_pool": "test"
            }]
        }],
        "outboxes": [{
            "name": "test",
            "senders": [{
                "name": "test",
                "translator": "test",
                "writer_pool": "test"
            }]
        }]
    }],
    "reader_pools": [{
        "name": "test",
        "urn": "urn:spirit:io:pool:reader:classic",
        "options": {
            "max_size": 100
        },
        "reader": {
            "name": "test",
            "urn": "urn:spirit:io:reader:std",
            "options": {
                "name": "ping",
                "proc": "ping",
                "args": ["-c", "10", "baidu.com"],
                "envs": {},
                "delim": "\n"
            }
        }
    }],
    "writer_pools": [{
        "name": "test",
        "urn": "urn:spirit:io:pool:writer:classic",
        "options": {
            "enable_session":true
            },
        "writer": {
            "name": "test",
            "urn": "urn:spirit:io:writer:std",
            "options": {
                "name": "write",
                "proc": "my-program-w",
                "args": [],
                "envs": []
            }
        }
    }],
    "input_translators": [{
        "name": "test",
        "urn": "urn:spirit:translator:in:line",
        "options": {
            "bind_urn": "std_base64@urn:spirit:component:util:base64#encode|std_base64@urn:spirit:component:util:base64#decode",
            "labels": {
                "version": "0.0.1"
            }
        }
    }],
    "output_translators": [{
        "name": "test",
        "urn": "urn:spirit:translator:out:line",
        "options": {
            "delim": "\n"
        }
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
            "put_timeout": 1000,
            "get_timeout": 1000
        }
    }],
    "routers": [{
        "name": "test",
        "urn": "urn:spirit:router:classic",
        "options": {
            "allow_no_component":true
        }
    }],
    "outboxes": [{
        "name": "test",
        "urn": "urn:spirit:outbox:classic",
        "options": {
            "size": 100,
            "get_timeout": -1,
            "labels": {"version":"0.0.1"}
        }
    }],
    "label_matchers": [{
        "name": "test",
        "urn": "urn:spirit:matcher:label:equal",
        "options": {}
    }],
    "components": [{
        "name": "std_base64",
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
