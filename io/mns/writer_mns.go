package mns

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gogap/ali_mns"

	"github.com/gogap/spirit"
)

const (
	msnWriterURN = "urn:spirit:io:writer:mns"
)

var (
	ErrMNSWriteMessageError = errors.New("ali-mns write message error")
)

var _ io.WriteCloser = new(MNSWriter)

type MNSWriterConfig struct {
	URL             string `json:"url"`
	AccessKeyId     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	Queue           string `json:"queue"`
	DelaySeconds    int    `json:"delay_seconds"`
	Priority        int    `json:"priority"`
	ProxyAddress    string `json:"proxy_address"`
}

type MNSWriter struct {
	conf MNSWriterConfig

	client ali_mns.MNSClient
}

func init() {
	spirit.RegisterWriter(msnWriterURN, NewMNSWriter)
}

func NewMNSWriter(config spirit.Config) (w io.WriteCloser, err error) {
	conf := MNSWriterConfig{}
	if err = config.ToObject(&conf); err != nil {
		return
	}

	if conf.Queue == "" {
		err = ErrAliMNSQueueIsEmpty
		return
	}

	if conf.URL == "" {
		err = ErrAliURLIsEmpty
		return
	}

	if conf.AccessKeyId == "" {
		err = ErrAliAccessKeyIdIsEmpty
		return
	}

	if conf.AccessKeySecret == "" {
		err = ErrAliAccessKeySecretIsEmpty
		return
	}

	if conf.Priority == 0 {
		conf.Priority = 8
	}

	if conf.DelaySeconds > 604800 {
		conf.DelaySeconds = 604800
	}

	if conf.DelaySeconds < 0 {
		conf.DelaySeconds = 0
	}

	client := ali_mns.NewAliMNSClient(conf.URL, conf.AccessKeyId, conf.AccessKeySecret)

	if conf.ProxyAddress != "" {
		client.SetProxy(conf.ProxyAddress)
	}

	w = &MNSWriter{conf: conf, client: client}

	return
}

func (p *MNSWriter) Write(data []byte) (n int, err error) {
	resource := fmt.Sprintf("queues/%s/%s", p.conf.Queue, "messages")

	reqData := ali_mns.MessageSendRequest{
		MessageBody:  ali_mns.Base64Bytes(data),
		DelaySeconds: int64(p.conf.DelaySeconds),
		Priority:     int64(p.conf.Priority),
	}

	var resp *http.Response
	if resp, err = p.client.Send(ali_mns.POST, nil, reqData, resource); err != nil {
		return
	}

	if resp != nil {
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated &&
			resp.StatusCode != http.StatusOK &&
			resp.StatusCode != http.StatusNoContent {
			err = ErrMNSWriteMessageError

			if body, e := ioutil.ReadAll(resp.Body); e == nil {
				spirit.Logger().WithField("actor", spirit.ActorWriter).
					WithField("urn", msnWriterURN).
					WithField("event", "write").
					WithField("url", p.conf.URL).
					WithField("queue", p.conf.Queue).
					WithField("status_code", resp.StatusCode).
					WithField("response", string(body)).
					Debugln(err)
			}
			return
		}
	}

	return
}

func (p *MNSWriter) Close() (err error) {
	return
}
