package spirit

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type Config map[string]interface{}

func (p Config) String(key string) (val string, err error) {
	if v, exist := p[key]; !exist {
		err = fmt.Errorf("config of %s not exist", key)
		return
	} else if strV, ok := v.(string); ok {
		val = strV
		return
	} else {
		err = fmt.Errorf("config of %s's value is not type of %s", key, reflect.TypeOf(val).Name())
		return
	}
	return
}

func (p Config) Int(key string) (val int, err error) {
	if v, exist := p[key]; !exist {
		err = fmt.Errorf("config of %s not exist", key)
		return
	} else if intV, ok := v.(int); ok {
		val = intV
		return
	} else {
		err = fmt.Errorf("config of %s's value is not type of %s", key, reflect.TypeOf(val).Name())
		return
	}
	return
}

func (p Config) Bool(key string) (val bool, err error) {
	if v, exist := p[key]; !exist {
		err = fmt.Errorf("config of %s not exist", key)
		return
	} else if boolV, ok := v.(bool); ok {
		val = boolV
		return
	} else {
		err = fmt.Errorf("config of %s's value is not type of %s", key, reflect.TypeOf(val).Name())
		return
	}
	return
}

func (p Config) Float64(key string) (val float64, err error) {
	if v, exist := p[key]; !exist {
		err = fmt.Errorf("config of %s not exist", key)
		return
	} else if floatV, ok := v.(float64); ok {
		val = floatV
		return
	} else {
		err = fmt.Errorf("config of %s's value is not type of %s", key, reflect.TypeOf(val).Name())
		return
	}
	return
}

func (p Config) Object(key string, v interface{}) (err error) {
	var obj interface{}
	if val, exist := p[key]; !exist {
		err = fmt.Errorf("config of %s not exist", key)
		return
	} else {
		obj = val
	}

	if obj == nil {
		v = nil
		return
	}

	var data []byte
	if data, err = json.Marshal(obj); err != nil {
		return
	}

	err = json.Unmarshal(data, v)

	return
}

func (p Config) ToObject(v interface{}) (err error) {
	var data []byte
	if data, err = json.Marshal(p); err != nil {
		return
	}

	err = json.Unmarshal(data, v)

	return
}
