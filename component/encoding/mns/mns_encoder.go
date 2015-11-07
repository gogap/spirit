package mns

import (
	"encoding/xml"
	"errors"
	"io"
	"reflect"

	"github.com/gogap/ali_mns"

	"github.com/gogap/spirit"
)

var _ spirit.Component = new(MNSEncodingComponent)

const (
	mnsEncodingURN = "urn:spirit:component:encoding:mns"
)

var (
	ErrDataTypeIsNotStruct      = errors.New("payload data type is not a struct")
	ErrDataTypeCouldNotBeDecode = errors.New("payload data type could not be decode")
)

type MNSEncodingComponentConfig struct {
	SingleMessage bool `json:"single_message"`
}

type MNSEncodingComponent struct {
	conf MNSEncodingComponentConfig
}

func init() {
	spirit.RegisterComponent(mnsEncodingURN, NewMNSEncodingComponent)
}

func NewMNSEncodingComponent(options spirit.Options) (component spirit.Component, err error) {
	conf := MNSEncodingComponentConfig{}

	if err = options.ToObject(&conf); err != nil {
		return
	}

	component = &MNSEncodingComponent{conf: conf}
	return
}

func (p *MNSEncodingComponent) URN() string {
	return mnsEncodingURN
}

func (p *MNSEncodingComponent) Labels() spirit.Labels {
	return spirit.Labels{
		"version": "0.0.1",
	}
}

func (p *MNSEncodingComponent) Handlers() spirit.Handlers {
	return spirit.Handlers{
		"Encode": p.Encode,
		"Decode": p.Decode,
	}
}

func (p *MNSEncodingComponent) Encode(payload spirit.Payload) (result interface{}, err error) {
	var vData interface{}
	if vData, err = payload.GetData(); err != nil {
		return
	}

	dataKind := reflect.TypeOf(vData).Kind()
	if dataKind == reflect.Ptr ||
		dataKind == reflect.Struct {
		var data []byte
		data, result = xml.Marshal(vData)
		result = data
	} else {
		err = ErrDataTypeIsNotStruct
		return
	}

	return
}

func (p *MNSEncodingComponent) Decode(payload spirit.Payload) (result interface{}, err error) {
	var vData interface{}
	if vData, err = payload.GetData(); err != nil {
		return
	}

	var resp interface{}

	if p.conf.SingleMessage {
		resp = &ali_mns.MessageReceiveResponse{}
	} else {
		resp = &ali_mns.BatchMessageReceiveResponse{}
	}

	switch data := vData.(type) {
	case string:
		{
			err = xml.Unmarshal([]byte(data), &resp)
		}
	case []byte:
		{
			err = xml.Unmarshal(data, &resp)
		}
	case io.Reader:
		{
			reader := xml.NewDecoder(data)
			err = reader.Decode(&resp)
		}
	default:
		err = ErrDataTypeCouldNotBeDecode
		return
	}

	if err == nil {
		result = resp
	}

	return
}
