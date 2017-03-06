package spirit

import (
	"bytes"
	"encoding/json"
	"strconv"

	"github.com/gogap/errors"
	"github.com/nu7hatch/gouuid"
)

const (
	ERROR_MSG_ADDR           = "-100"
	ERROR_MSG_ADDR_INT int32 = -100
)

type MessageGraph map[string]MessageAddress

func (p MessageGraph) AddAddress(addrs ...MessageAddress) int {
	lenAddr := 0
	startIndex := len(p) + 1
	if addrs != nil {
		lenAddr = len(addrs)
		for i, addr := range addrs {
			p[strconv.Itoa(i+startIndex)] = addr
		}
	}
	return lenAddr
}

func (p MessageGraph) AddAddressToHead(addrs ...MessageAddress) int {

	lAddr := len(addrs)
	if lAddr == 0 {
		return 0
	}

	// move
	for k, v := range p {
		i, _ := strconv.Atoi(k)
		p[strconv.Itoa(i+lAddr)] = v
	}

	startIndex := 1
	if addrs != nil {
		for i, addr := range addrs {
			p[strconv.Itoa(i+startIndex)] = addr
		}
	}

	return lAddr
}

func (p MessageGraph) SetErrorAddress(address MessageAddress) {
	p[ERROR_MSG_ADDR] = address
}

func (p MessageGraph) ClearErrorAddress() {
	if _, exist := p[ERROR_MSG_ADDR]; exist {
		delete(p, ERROR_MSG_ADDR)
	}
}

type ComponentMessage struct {
	id                string
	graph             MessageGraph
	currentGraphIndex int32
	payload           Payload
	hooksMetaData     []MessageHookMetadata
}

func NewComponentMessage(graph MessageGraph, payload Payload) (message ComponentMessage, err error) {
	msgId := ""

	if payload.id == "" {
		if id, e := uuid.NewV4(); e != nil {
			err = ERR_UUID_GENERATE_FAILED.New(errors.Params{"err": e})
			return
		} else {
			msgId = id.String()
		}
		payload.id = msgId
	} else {
		msgId = payload.id
	}

	message = ComponentMessage{
		id:                msgId,
		graph:             graph,
		currentGraphIndex: 1,
		payload:           payload,
	}

	return
}

func (p *ComponentMessage) Serialize() (data []byte, err error) {
	jsonMap := map[string]interface{}{
		"id":                     p.id,
		"graph":                  p.graph,
		"current_graph_index":    p.currentGraphIndex,
		"message_hooks_metadata": p.hooksMetaData,
		"payload": map[string]interface{}{
			"id":      p.id,
			"context": p.payload.context,
			"command": p.payload.command,
			"content": p.payload.content,
			"error":   p.payload.err,
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
		Id                string                `json:"id"`
		Graph             MessageGraph          `json:"graph"`
		CurrentGraphIndex int32                 `json:"current_graph_index"`
		HooksMetaData     []MessageHookMetadata `json:"message_hooks_metadata"`
		Payload           struct {
			Id      string            `json:"id,omitempty"`
			Context ComponentContext  `json:"context,omitempty"`
			Command ComponentCommands `json:"command,omitempty"`
			Content interface{}       `json:"content,omitempty"`
			Error   Error             `json:"error,omitempty"`
		} `json:"payload"`
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	if err = decoder.Decode(&tmp); err != nil {
		return
	}

	p.id = tmp.Id
	p.graph = tmp.Graph
	p.currentGraphIndex = tmp.CurrentGraphIndex
	p.payload = Payload{
		id:      tmp.Payload.Id,
		context: tmp.Payload.Context,
		command: tmp.Payload.Command,
		content: tmp.Payload.Content,
		err:     tmp.Payload.Error,
	}
	p.hooksMetaData = tmp.HooksMetaData

	return
}

func (p *ComponentMessage) Id() string {
	return p.id
}
