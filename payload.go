package spirit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/gogap/errors"
)

type ComponentCommands map[string][]interface{}
type ComponentContext map[string]interface{}

type Error struct {
	Id        string `json:"id,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Code      uint64 `json:"code,omitempty"`
	AddressId int32  `json:"address_id,omitempty"`
	Message   string `json:"message,omitempty"`
}

type Payload struct {
	id      string
	context ComponentContext
	command ComponentCommands
	content interface{}
	err     Error
}

func (p *Payload) CopyFrom(payload *Payload) {
	p.context = payload.context
	p.command = payload.command
	p.content = payload.content
	p.err = payload.err
}

func (p *Payload) Serialize() (data []byte, err error) {
	jsonMap := map[string]interface{}{
		"id":      p.id,
		"context": p.context,
		"command": p.command,
		"content": p.content,
		"error":   p.err,
	}

	if data, err = json.Marshal(&jsonMap); err != nil {
		err = ERR_PAYLOAD_SERIALIZE_FAILED.New(errors.Params{"err": err})
		return
	}
	return
}

func (p *Payload) UnSerialize(data []byte) (err error) {
	var tmp struct {
		Id      string            `json:"id,omitempty"`
		Context ComponentContext  `json:"context,omitempty"`
		Command ComponentCommands `json:"command,omitempty"`
		Content interface{}       `json:"content,omitempty"`
		Error   Error             `json:"error,omitempty"`
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	if err = decoder.Decode(&tmp); err != nil {
		return
	}

	p.id = tmp.Id
	p.context = tmp.Context
	p.command = tmp.Command
	p.content = tmp.Content
	p.err = tmp.Error

	return
}

func (p *Payload) Id() string {
	return p.id
}

func (p *Payload) IsCorrect() bool {
	return p.err.Code == 0
}

func (p *Payload) Error() Error {
	return p.err
}

func (p *Payload) GoBad(e error) {
	if errCode, ok := e.(errors.ErrCode); ok {
		p.err.Id = errCode.Id()
		p.err.Namespace = errCode.Namespace()
		p.err.Code = errCode.Code()
		p.err.Message = errCode.Error()
	} else {
		err := ERR_PAYLOAD_GO_BAD.New(errors.Params{"err": e})
		p.err.Id = err.Id()
		p.err.Namespace = err.Namespace()
		p.err.Code = err.Code()
		p.err.Message = err.Error()
	}
}

func (p *Payload) GetContent() interface{} {
	return p.content
}

func (p *Payload) FillContentToObject(v interface{}) (err error) {
	if p.content == nil {
		err = errors.New("content is nil")
		return
	}

	var data []byte
	if data, err = json.Marshal(p.content); err != nil {
		return
	} else {
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.UseNumber()

		if err = decoder.Decode(v); err != nil {
			return
		}
		return
	}
}

func (p *Payload) SetContent(content interface{}) {
	p.content = content
}

func (p *Payload) SetContext(key string, val interface{}) {
	if p.context == nil {
		p.context = make(map[string]interface{})
	}
	p.context[key] = val
}

func (p *Payload) GetContext(key string) (val interface{}, exist bool) {
	if p.context == nil {
		return nil, false
	}
	val, exist = p.context[key]
	return
}

func (p *Payload) GetContextString(key string) (val string, err error) {
	if p.context == nil {
		return "", fmt.Errorf("the context container is nil")
	}

	var v interface{}
	exist := false
	if v, exist = p.context[key]; !exist {
		err = fmt.Errorf("the context key of %s is not exist", key)
		return
	}

	if strVal, ok := v.(string); ok {
		val = strVal
		return
	}

	return
}

func (p *Payload) GetContextStringArray(key string) (vals []string, err error) {
	if p.context == nil {
		return nil, fmt.Errorf("the context container is nil")
	}

	var v interface{}
	exist := false
	if v, exist = p.context[key]; !exist {
		err = fmt.Errorf("the context key of %s is not exist", key)
		return
	}

	if vInterfaces, ok := v.([]interface{}); ok {
		tmpArray := []string{}
		for i, vStr := range vInterfaces {
			if str, ok := vStr.(string); ok {
				tmpArray = append(tmpArray, str)
			} else {
				err = fmt.Errorf("the context key of %s's value type at index of %d is not string", key, i)
				return
			}
		}
		vals = tmpArray
		return
	} else {
		err = fmt.Errorf("the type of context key %s is not array", key)
		return
	}
	return
}

func (p *Payload) GetContextInt(key string) (val int, err error) {
	if p.context == nil {
		return 0, fmt.Errorf("the context container is nil")
	}

	var v interface{}
	exist := false
	if v, exist = p.context[key]; !exist {
		err = fmt.Errorf("the context key of %s is not exist", key)
		return
	}

	if intVal, ok := v.(int); ok {
		val = intVal
		return
	} else {
		err = fmt.Errorf("the type of context key %s is not int", key)
	}
	return
}

func (p *Payload) GetContextInt32(key string) (val int32, err error) {
	if p.context == nil {
		return 0, fmt.Errorf("the context container is nil")
	}

	var v interface{}
	exist := false
	if v, exist = p.context[key]; !exist {
		err = fmt.Errorf("the context key of %s is not exist", key)
		return
	}

	if intVal, ok := v.(int32); ok {
		val = intVal
		return
	} else {
		err = fmt.Errorf("the type of context key %s is not int32", key)
	}
	return
}

func (p *Payload) GetContextInt64(key string) (val int64, err error) {
	if p.context == nil {
		return 0, fmt.Errorf("the context container is nil")
	}

	var v interface{}
	exist := false
	if v, exist = p.context[key]; !exist {
		err = fmt.Errorf("the context key of %s is not exist", key)
		return
	}

	if intVal, ok := v.(int64); ok {
		val = intVal
		return
	} else {
		err = fmt.Errorf("the type of context key %s is not int64", key)
	}
	return
}

func (p *Payload) GetContextObject(key string, v interface{}) (err error) {
	if v == nil {
		err = fmt.Errorf("the v should not be nil, it should be a Pointer")
		return
	}

	if p.context == nil {
		return fmt.Errorf("the context container is nil")
	}

	if val, exist := p.context[key]; !exist {
		err = fmt.Errorf("the context key of %s is not exist", key)
		return
	} else if val == nil {
		err = fmt.Errorf("the context key of %s is exist, but the value is nil", key)
		return
	} else {
		if bJson, e := json.Marshal(val); e != nil {
			err = fmt.Errorf("marshal object of %s to json failed, error is:%v", key, e)
			return
		} else {
			decoder := json.NewDecoder(bytes.NewReader(bJson))
			decoder.UseNumber()

			if e := decoder.Decode(v); e != nil {
				err = fmt.Errorf("unmarshal json to object %s failed, error is:%v", key, e)
				return
			}
		}
	}
	return
}

func (p *Payload) SetCommand(command string, values []interface{}) {
	if p.command == nil {
		p.command = make(map[string][]interface{})
	}
	p.command[command] = values
}

func (p *Payload) AppendCommand(command string, value interface{}) {
	if p.command == nil {
		p.command = make(map[string][]interface{})
	}

	if values, ok := p.command[command]; !ok {
		p.command[command] = []interface{}{value}
	} else {
		values = append(values, value)
		p.command[command] = values
	}
	return
}

func (p *Payload) GetCommand(key string) (val []interface{}, exist bool) {
	if p.command == nil {
		return nil, false
	}
	val, exist = p.command[key]
	return
}

func (p *Payload) GetCommandValueSize(key string) int {
	if p.command == nil {
		return 0
	} else {
		if vals, exist := p.command[key]; exist {
			if vals != nil {
				return len(vals)
			}
			return 0
		}
	}
	return 0
}

func (p *Payload) GetCommandStringArray(command string) (vals []string, err error) {
	if size := p.GetCommandValueSize(command); size > 0 {
		values, _ := p.GetCommand(command)
		tmpVals := []string{}
		for _, iStr := range values {
			if strV, ok := iStr.(string); ok {
				tmpVals = append(tmpVals, strV)
			} else {
				err = fmt.Errorf("the value of %v (%s) are not string type", iStr, reflect.TypeOf(iStr).String())
				return
			}
		}
		vals = tmpVals
		return
	}
	err = fmt.Errorf("command values is nil or command not exist")
	return
}

func (p *Payload) GetCommandObjectArray(command string, values []interface{}) (err error) {

	if values == nil {
		err = fmt.Errorf("the values should not be nil, it should be a interface{}")
		return
	}

	if len(values) == 0 {
		return
	}

	if p.GetCommandValueSize(command) < len(values) {
		err = fmt.Errorf("the command of %s is exist, but the recv values length is greater than command values", command)
		return
	}

	vals, _ := p.GetCommand(command)

	for i, objVal := range vals {
		var bJson []byte
		var e error
		if bJson, e = json.Marshal(objVal); e != nil {
			err = fmt.Errorf("marshal object of %s to json failed, error is:%v", command, e)
			return
		}

		decoder := json.NewDecoder(bytes.NewReader(bJson))
		decoder.UseNumber()

		if e = decoder.Decode(&values[i]); e != nil {
			err = fmt.Errorf("unmarshal json to object %s failed, error is:%v", command, e)
			return
		}
	}

	return
}
