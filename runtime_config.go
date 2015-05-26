package spirit

import (
	"github.com/gogap/env_json"
)

type runComponentConf struct {
	Name      string         `json:"name"`
	Address   []addressConf  `json:"address,omitempty"`
	PortHooks []portHookConf `json:"port_hooks,omitempty"`
}

type addressConf struct {
	PortName    string                 `json:"port_name"`
	HandlerName string                 `json:"handler_name"`
	Type        string                 `json:"type"`
	Url         string                 `json:"url"`
	Options     map[string]interface{} `json:"options"`
}

type heartbeatConf struct {
	Heart    string                 `json:"heart"`
	Interval int64                  `json:"interval"`
	Options  map[string]interface{} `json:"options"`
}

type portHookConf struct {
	Type    string                 `json:"type"`
	Name    string                 `json:"name"`
	Port    string                 `json:"port_name"`
	Options map[string]interface{} `json:"options"`
}

type globalHookConf struct {
	Type    string                 `json:"type"`
	Name    string                 `json:"name"`
	Options map[string]interface{} `json:"options"`
}

type runtimeConfig struct {
	Components   []runComponentConf `json:"components"`
	Heartbeat    []heartbeatConf    `json:"heartbeat,omitempty"`
	GlobalHooks  []globalHookConf   `json:"global_hooks,omitempty"`
	StoreConfigs []string           `json:"store_conf,omitempty"`
}

func (p *runtimeConfig) Serialize() (str string, err error) {
	var data []byte
	if data, err = env_json.MarshalIndent(p, "", "    "); err != nil {
		return
	}

	str = string(data)

	return
}
