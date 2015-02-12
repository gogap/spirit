package spirit

import (
	"encoding/json"

	"github.com/gogap/errors"
)

type MessageGraph map[int32]MessageAddress

type ComponentMessage struct {
	id                string       `json:"id"`
	graph             MessageGraph `json:"graph"`
	currentGraphIndex int32        `json:"current_graph_index"`
	payload           Payload      `json:"payload"`
}

func NewComponentMessage(content interface{}) (message ComponentMessage, err error) {
	return
}

func (p *ComponentMessage) Serialize() (data []byte, err error) {
	jsonMap := map[string]interface{}{
		"id":                  p.id,
		"graph":               p.graph,
		"current_graph_index": p.currentGraphIndex,
		"payload": map[string]interface{}{
			"code":    p.payload.code,
			"message": p.payload.message,
			"context": p.payload.context,
			"command": p.payload.command,
			"content": p.payload.content,
		},
	}

	if data, err = json.Marshal(&jsonMap); err != nil {
		err = ERR_MESSAGE_SERIALIZE_FAILED.New(errors.Params{"err": err})
		return
	}
	return
}

func (p *ComponentMessage) UnSerialize(data []byte) (err error) {
	var tmp struct {
		Id                string       `json:"id"`
		Graph             MessageGraph `json:"graph"`
		CurrentGraphIndex int32        `json:"current_graph_index"`
		Payload           struct {
			Code    string            `json:"code"`
			Message string            `json:"message"`
			Context ComponentContext  `json:"context,omitempty"`
			Command ComponentCommands `json:"command,omitempty"`
			Content interface{}       `json:"content"`
		} `json:"payload"`
	}

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	p.id = tmp.Id
	p.graph = tmp.Graph
	p.currentGraphIndex = tmp.CurrentGraphIndex
	p.payload = Payload{
		code:    tmp.Payload.Code,
		message: tmp.Payload.Message,
		context: tmp.Payload.Context,
		command: tmp.Payload.Command,
		content: tmp.Payload.Content}

	return
}
