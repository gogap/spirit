package spirit

import (
	"reflect"
	"strconv"

	"github.com/gogap/env_json"
	"github.com/gogap/errors"
)

type Options map[string]interface{}

func (p Options) Serialize() (str string, err error) {
	var data []byte
	if data, err = env_json.MarshalIndent(&p, "", "    "); err != nil {
		return
	}

	str = string(data)

	return
}

func (p Options) GetStringValue(key string) (value string, err error) {
	if val, exist := p[key]; !exist {
		err = ERR_OPTIONS_KEY_NOT_EXIST.New(errors.Params{"key": key})
		return
	} else if strVal, ok := val.(string); !ok {
		err = ERR_OPTIONS_VALUE_TYPE_ERROR.New(errors.Params{"key": key, "value": val, "type": "string", "realType": getValueType(val)})
		return
	} else {
		value = strVal
	}
	return
}

func (p Options) GetInt64Value(key string) (value int64, err error) {
	if val, exist := p[key]; !exist {
		err = ERR_OPTIONS_KEY_NOT_EXIST.New(errors.Params{"key": key})
		return
	} else if intVal, ok := val.(int64); !ok {
		switch typedVal := val.(type) {
		case float64:
			value = int64(typedVal)
		case int:
			value = int64(typedVal)
		case int32:
			value = int64(typedVal)
		case string:
			if intV, e := strconv.Atoi(typedVal); e != nil {
				err = ERR_OPTIONS_VAL_TYPE_CONV_FAILED.New(errors.Params{"key": key, "value": val, "type": "int64", "realType": getValueType(val), "err": e})
				return
			} else {
				value = int64(intV)
			}
		default:
			err = ERR_OPTIONS_VALUE_TYPE_ERROR.New(errors.Params{"key": key, "value": val, "type": "int64", "realType": getValueType(val)})
			return
		}
		return
	} else {
		value = intVal
	}
	return
}

func (p Options) GetFloat64Value(key string) (value float64, err error) {
	if val, exist := p[key]; !exist {
		err = ERR_OPTIONS_KEY_NOT_EXIST.New(errors.Params{"key": key})
		return
	} else if floatVal, ok := val.(float64); !ok {
		err = ERR_OPTIONS_VALUE_TYPE_ERROR.New(errors.Params{"key": key, "value": val, "type": "float64", "realType": getValueType(val)})
		return
	} else {
		value = floatVal
	}
	return
}

func getValueType(v interface{}) string {
	return reflect.TypeOf(v).Name()
}
