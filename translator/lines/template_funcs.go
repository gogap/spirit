package lines

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
	"text/template"
	"time"
)

//some function copied from https://github.com/spf13/hugo

var funcMap template.FuncMap

func init() {
	funcMap = template.FuncMap{
		"isNil":        isNil,
		"getJSON":      getJSON,
		"getenv":       func(varName string) string { return os.Getenv(varName) },
		"replaceMaps":  replaceMaps,
		"replace":      replace,
		"regexCompile": regexp.Compile,
		"substr":       substr,
		"split":        split,
		"in":           in,
		"toStr":        toStr,
		"dateFormat":   dateFormat,
		"now":          time.Now,
		"localtime":    time.Now().Local,
		"newDict":      newDict,
		"newArray":     newArray,
		"scanf":        fmt.Sscanf,
		"exist":        exist,
		"add":          func(a, b interface{}) (interface{}, error) { return doArithmetic(a, b, '+') },
		"sub":          func(a, b interface{}) (interface{}, error) { return doArithmetic(a, b, '-') },
		"div":          func(a, b interface{}) (interface{}, error) { return doArithmetic(a, b, '/') },
	}
}

type dict map[string]interface{}

func newDict() dict {
	return dict{}
}

func (p dict) put(key string, val interface{}) bool {
	p[key] = val
	return true
}

func (p dict) get(key string) (val interface{}) {
	val, _ = p[key]
	return
}

func (p dict) exist(key string) (exist bool) {
	_, exist = p[key]
	return
}

func (p dict) del(key string) bool {
	if _, exist := p[key]; exist {
		delete(p, key)
		return true
	}
	return false
}

type array struct {
	items []interface{}
}

func newArray() *array {
	return &array{
		items: make([]interface{}, 0),
	}
}

func (p *array) append(v interface{}) bool {
	p.items = append(p.items, v)
	return true
}

func (p array) join(sep string) string {
	strs := []string{}

	for _, item := range p.items {
		switch val := item.(type) {
		case string:
			{
				strs = append(strs, val)
			}
		default:
			{
				strv := fmt.Sprintf("%v", item)
				strs = append(strs, strv)
			}
		}
	}
	return strings.Join(strs, sep)
}

func toStr(v interface{}) (str string) {
	return fmt.Sprintf("%v", v)
}

func getJSON(v interface{}) (jsonStr string, err error) {
	var data []byte

	if data, err = json.Marshal(v); err != nil {
		return
	}

	jsonStr = string(data)

	return
}

func replaceMaps(v interface{}, key string, expr string, repl string) (ret interface{}, err error) {
	kind := reflect.TypeOf(v).Kind()
	switch kind {
	case reflect.Slice:
		{
			items, _ := v.([]interface{})
			mapItems := []map[string]interface{}{}
			for _, item := range items {
				if mapItem, ok := item.(map[string]interface{}); ok {
					mapItems = append(mapItems, mapItem)
				} else {
					err = fmt.Errorf("the map's values type is not interface{}")
					return
				}
			}
			return replaceMapSlice(mapItems, key, expr, repl)
		}
	case reflect.Map:
		{
			if m, ok := v.(map[string]interface{}); ok {
				return replaceMap(m, key, expr, repl)
			} else {
				err = fmt.Errorf("the map's values type is not interface{}")
				return
			}
		}
	default:
		{
			err = fmt.Errorf("unsupport Kind of %v", kind)
			return
		}
	}
	return
}

func replace(v interface{}, expr string, repl string) (ret interface{}, err error) {
	kind := reflect.TypeOf(v).Kind()
	switch kind {
	case reflect.Slice:
		{
			strs, _ := v.([]interface{})
			strItems := []string{}
			for _, strV := range strs {
				if str, ok := strV.(string); ok {
					strItems = append(strItems, str)
				} else {
					err = fmt.Errorf("the map's values type is not string")
					return
				}
			}
			return replaceStringSlice(strItems, expr, repl)
		}
	case reflect.String:
		{
			str, _ := v.(string)
			return replaceString(str, expr, repl)
		}
	default:
		{
			err = fmt.Errorf("unsupport Kind of %v", kind)
			return
		}
	}
}

func replaceStringSlice(strs []string, expr string, repl string) (ret []string, err error) {
	if strs != nil {
		rets := make([]string, 0)
		for _, src := range strs {
			if rStr, e := replaceString(src, expr, repl); e != nil {
				err = e
				return
			} else {
				rets = append(rets, rStr)
			}
		}
		ret = rets
	}
	return
}

func replaceString(src string, expr string, repl string) (ret string, err error) {
	var reg *regexp.Regexp
	if reg, err = regexp.Compile(expr); err != nil {
		return
	}

	ret = reg.ReplaceAllString(src, repl)

	return
}

func replaceMapSlice(data []map[string]interface{}, key string, expr string, repl string) (ret []map[string]interface{}, err error) {
	if data != nil {
		retMap := make([]map[string]interface{}, 0)
		for _, item := range data {
			if rMap, e := replaceMap(item, key, expr, repl); e != nil {
				err = e
				return
			} else {
				retMap = append(retMap, rMap)
			}
		}
		ret = retMap
	}
	return
}

func replaceMap(data map[string]interface{}, key string, expr string, repl string) (ret map[string]interface{}, err error) {

	src := ""

	if val, exist := data[key]; !exist {
		return
	} else if strVal, ok := val.(string); ok {
		src = strVal
	} else {
		err = fmt.Errorf("value of %s's type is not string", key)
		return
	}

	retval := ""
	if retval, err = replaceString(src, expr, repl); err != nil {
		return
	}

	ret = make(map[string]interface{})

	for k, v := range data {
		ret[k] = v
	}

	ret[key] = retval

	return
}

func dateFormat(layout string, v interface{}) (string, error) {
	t, err := toTimeE(v)

	if err != nil {
		return "", err
	}
	return t.Format(layout), nil
}

func split(a interface{}, delimiter string) ([]string, error) {
	aStr := ""
	if str, ok := a.(string); !ok {
		return nil, fmt.Errorf("type of a is not string")
	} else {
		aStr = str
	}

	return strings.Split(aStr, delimiter), nil
}

func substr(a interface{}, nums ...interface{}) (string, error) {

	aStr := ""
	if str, ok := a.(string); !ok {
		return "", fmt.Errorf("type of a is not string")
	} else {
		aStr = str
	}

	var start, length int
	toInt := func(v interface{}, message string) (int, error) {
		switch i := v.(type) {
		case int:
			return i, nil
		case int8:
			return int(i), nil
		case int16:
			return int(i), nil
		case int32:
			return int(i), nil
		case int64:
			return int(i), nil
		default:
			return 0, fmt.Errorf(message)
		}
	}

	var err error
	switch len(nums) {
	case 0:
		return "", fmt.Errorf("too less arguments")
	case 1:
		if start, err = toInt(nums[0], "start argument must be integer"); err != nil {
			return "", err
		}
		length = len(aStr)
	case 2:
		if start, err = toInt(nums[0], "start argument must be integer"); err != nil {
			return "", err
		}
		if length, err = toInt(nums[1], "length argument must be integer"); err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("too many arguments")
	}

	if start < -len(aStr) {
		start = 0
	}
	if start > len(aStr) {
		return "", fmt.Errorf("start position out of bounds for %d-byte string", len(aStr))
	}

	var s, e int
	if start >= 0 && length >= 0 {
		s = start
		e = start + length
	} else if start < 0 && length >= 0 {
		s = len(aStr) + start - length + 1
		e = len(aStr) + start + 1
	} else if start >= 0 && length < 0 {
		s = start
		e = len(aStr) + length
	} else {
		s = len(aStr) + start
		e = len(aStr) + length
	}

	if s > e {
		return "", fmt.Errorf("calculated start position greater than end position: %d > %d", s, e)
	}
	if e > len(aStr) {
		e = len(aStr)
	}

	return aStr[s:e], nil
}

func in(l interface{}, v interface{}) bool {
	lv := reflect.ValueOf(l)
	vv := reflect.ValueOf(v)

	switch lv.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < lv.Len(); i++ {
			lvv := lv.Index(i)
			switch lvv.Kind() {
			case reflect.String:
				if vv.Type() == lvv.Type() && vv.String() == lvv.String() {
					return true
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				switch vv.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					if vv.Int() == lvv.Int() {
						return true
					}
				}
			case reflect.Float32, reflect.Float64:
				switch vv.Kind() {
				case reflect.Float32, reflect.Float64:
					if vv.Float() == lvv.Float() {
						return true
					}
				}
			}
		}
	case reflect.String:
		if vv.Type() == lv.Type() && strings.Contains(lv.String(), vv.String()) {
			return true
		}
	}
	return false
}

func toTimeE(i interface{}) (tim time.Time, err error) {

	if rv, isNil := indirect(reflect.ValueOf(i)); isNil {
		i = nil
	} else {
		i = rv.Interface()
	}

	switch s := i.(type) {
	case time.Time:
		return s, nil
	case string:
		d, e := stringToDate(s)
		if e == nil {
			return d, nil
		}
		return time.Time{}, fmt.Errorf("Could not parse Date/Time format: %v\n", e)
	default:
		return time.Time{}, fmt.Errorf("Unable to Cast %#v to Time\n", i)
	}
}

func stringToDate(s string) (time.Time, error) {
	return parseDateWith(s, []string{
		time.RFC3339,
		"2006-01-02T15:04:05", // iso8601 without timezone
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		"2006-01-02 15:04:05Z07:00",
		"02 Jan 06 15:04 MST",
		"2006-01-02",
		"02 Jan 2006",
	})
}

func parseDateWith(s string, dates []string) (d time.Time, e error) {
	for _, dateType := range dates {
		if d, e = time.Parse(dateType, s); e == nil {
			return
		}
	}
	return d, fmt.Errorf("Unable to parse date: %s", s)
}

// func indirect(a interface{}) interface{} {
// 	if a == nil {
// 		return nil
// 	}
// 	if t := reflect.TypeOf(a); t.Kind() != reflect.Ptr {
// 		// Avoid creating a reflect.Value if it's not a pointer.
// 		return a
// 	}
// 	v := reflect.ValueOf(a)
// 	for v.Kind() == reflect.Ptr && !v.IsNil() {
// 		v = v.Elem()
// 	}
// 	return v.Interface()
// }

func indirect(v reflect.Value) (rv reflect.Value, isNil bool) {
	for ; v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface; v = v.Elem() {
		if v.IsNil() {
			return v, true
		}
		if v.Kind() == reflect.Interface && v.NumMethod() > 0 {
			break
		}
	}
	return v, false
}

func exist(item interface{}, indices ...interface{}) bool {
	v := reflect.ValueOf(item)
	for _, i := range indices {
		index := reflect.ValueOf(i)
		var isNil bool
		if v, isNil = indirect(v); isNil {
			return false
		}
		switch v.Kind() {
		case reflect.Array, reflect.Slice, reflect.String:
			var x int64
			switch index.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				x = index.Int()
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
				x = int64(index.Uint())
			default:
				return false
			}
			if x < 0 || x >= int64(v.Len()) {
				return false
			}
			v = v.Index(int(x))
		case reflect.Map:
			if !index.IsValid() {
				index = reflect.Zero(v.Type().Key())
			}
			if !index.Type().AssignableTo(v.Type().Key()) {
				return false
			}
			if x := v.MapIndex(index); x.IsValid() {
				v = x
			} else {
				v = reflect.Zero(v.Type().Elem())
			}
		default:
			return false
		}
	}
	return !v.IsNil()
}

func doArithmetic(a, b interface{}, op rune) (interface{}, error) {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)
	var ai, bi int64
	var af, bf float64
	var au, bu uint64
	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ai = av.Int()
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			bi = bv.Int()
		case reflect.Float32, reflect.Float64:
			af = float64(ai) // may overflow
			ai = 0
			bf = bv.Float()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			bu = bv.Uint()
			if ai >= 0 {
				au = uint64(ai)
				ai = 0
			} else {
				bi = int64(bu) // may overflow
				bu = 0
			}
		default:
			return nil, fmt.Errorf("Can't apply the operator to the values")
		}
	case reflect.Float32, reflect.Float64:
		af = av.Float()
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			bf = float64(bv.Int()) // may overflow
		case reflect.Float32, reflect.Float64:
			bf = bv.Float()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			bf = float64(bv.Uint()) // may overflow
		default:
			return nil, fmt.Errorf("Can't apply the operator to the values")
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		au = av.Uint()
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			bi = bv.Int()
			if bi >= 0 {
				bu = uint64(bi)
				bi = 0
			} else {
				ai = int64(au) // may overflow
				au = 0
			}
		case reflect.Float32, reflect.Float64:
			af = float64(au) // may overflow
			au = 0
			bf = bv.Float()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			bu = bv.Uint()
		default:
			return nil, fmt.Errorf("Can't apply the operator to the values")
		}
	case reflect.String:
		as := av.String()
		if bv.Kind() == reflect.String && op == '+' {
			bs := bv.String()
			return as + bs, nil
		} else {
			return nil, fmt.Errorf("Can't apply the operator to the values")
		}
	default:
		return nil, fmt.Errorf("Can't apply the operator to the values")
	}

	switch op {
	case '+':
		if ai != 0 || bi != 0 {
			return ai + bi, nil
		} else if af != 0 || bf != 0 {
			return af + bf, nil
		} else if au != 0 || bu != 0 {
			return au + bu, nil
		} else {
			return 0, nil
		}
	case '-':
		if ai != 0 || bi != 0 {
			return ai - bi, nil
		} else if af != 0 || bf != 0 {
			return af - bf, nil
		} else if au != 0 || bu != 0 {
			return au - bu, nil
		} else {
			return 0, nil
		}
	case '*':
		if ai != 0 || bi != 0 {
			return ai * bi, nil
		} else if af != 0 || bf != 0 {
			return af * bf, nil
		} else if au != 0 || bu != 0 {
			return au * bu, nil
		} else {
			return 0, nil
		}
	case '/':
		if bi != 0 {
			return ai / bi, nil
		} else if bf != 0 {
			return af / bf, nil
		} else if bu != 0 {
			return au / bu, nil
		} else {
			return nil, fmt.Errorf("Can't divide the value by 0")
		}
	default:
		return nil, fmt.Errorf("There is no such an operation")
	}
}

func isNil(v interface{}) bool {
	return v == nil
}
