package mns

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/gogap/ali_mns"

	"github.com/gogap/spirit"
)

const (
	mnsReaderURN = "urn:spirit:io:reader:mns"
)

var (
	ErrAliMNSQueueIsEmpty        = errors.New("ali-mns queue is empty")
	ErrAliURLIsEmpty             = errors.New("ali-mns url is empty")
	ErrAliAccessKeyIdIsEmpty     = errors.New("ali-mns access_key_id is empty")
	ErrAliAccessKeySecretIsEmpty = errors.New("ali-mns access_key_secret is empty")

	ErrAliMNSResponseError = errors.New("ali-mns response error")
)

var _ io.ReadCloser = new(MNSReader)

type MNSReaderConfig struct {
	URL                string `json:"url"`
	AccessKeyId        string `json:"access_key_id"`
	AccessKeySecret    string `json:"access_key_secret"`
	Queue              string `json:"queue"`
	PollingWaitSeconds int    `json:"polling_wait_seconds"`
	MaxReadCount       int    `json:"max_read_count"`
	ProxyAddress       string `json:"proxy_address"`
}

type MNSReader struct {
	conf MNSReaderConfig

	client ali_mns.MNSClient

	readLocker sync.Mutex
	reading    bool

	tmpReader io.Reader
}

func init() {
	spirit.RegisterReader(mnsReaderURN, NewMNSReader)
}

func NewMNSReader(config spirit.Config) (r io.ReadCloser, err error) {
	conf := MNSReaderConfig{}
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

	if conf.MaxReadCount <= 0 || conf.MaxReadCount > 16 {
		conf.MaxReadCount = int(ali_mns.DefaultNumOfMessages)
	}

	if conf.PollingWaitSeconds <= 0 || conf.PollingWaitSeconds > 30 {
		conf.PollingWaitSeconds = int(ali_mns.DefaultTimeout)
	}

	client := ali_mns.NewAliMNSClient(conf.URL, conf.AccessKeyId, conf.AccessKeySecret)

	if conf.ProxyAddress != "" {
		client.SetProxy(conf.ProxyAddress)
	}

	r = &MNSReader{conf: conf, client: client}

	return
}

func (p *MNSReader) Read(data []byte) (n int, err error) {
	p.readLocker.Lock()
	defer p.readLocker.Unlock()

	if !p.reading {
		resource := fmt.Sprintf("queues/%s/%s?numOfMessages=%d", p.conf.Queue, "messages", p.conf.MaxReadCount)

		if p.conf.PollingWaitSeconds > 0 {
			resource = fmt.Sprintf("queues/%s/%s?numOfMessages=%d&waitseconds=%d", p.conf.Queue, "messages", p.conf.MaxReadCount, p.conf.PollingWaitSeconds)
		}

		var resp *http.Response
		if resp, err = p.client.Send(ali_mns.GET, nil, nil, resource); err != nil {
			return
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			if resp.StatusCode == http.StatusNoContent {

				spirit.Logger().WithField("actor", spirit.ActorReader).
					WithField("urn", mnsReaderURN).
					WithField("url", p.conf.URL).
					WithField("queue", p.conf.Queue).
					WithField("status_code", resp.StatusCode).
					Debugln("no content")

				return
			}

			mnsDecoder := ali_mns.NewAliMNSDecoder()
			errResp := ali_mns.ErrorMessageResponse{}

			if e := mnsDecoder.Decode(resp.Body, &errResp); e == nil {
				spirit.Logger().WithField("actor", spirit.ActorReader).
					WithField("urn", mnsReaderURN).
					WithField("event", "read").
					WithField("url", p.conf.URL).
					WithField("queue", p.conf.Queue).
					WithField("status_code", resp.StatusCode).
					WithField("code", errResp.Code).
					WithField("host_id", errResp.HostId).
					WithField("request_id", errResp.RequestId).
					Errorln(errors.New(errResp.Message))
			} else {
				spirit.Logger().WithField("actor", spirit.ActorReader).
					WithField("urn", mnsReaderURN).
					WithField("event", "decode mns error response").
					WithField("url", p.conf.URL).
					WithField("queue", p.conf.Queue).
					WithField("status_code", resp.StatusCode).
					Errorln(e)
			}

			err = ErrAliMNSResponseError
			return
		}

		decoder := ali_mns.NewAliMNSDecoder()
		batchMessages := ali_mns.BatchMessageReceiveResponse{}
		if err = decoder.Decode(resp.Body, &batchMessages); err != nil {
			return
		}

		var buf bytes.Buffer
		for _, message := range batchMessages.Messages {
			if _, err = buf.Write(message.MessageBody); err != nil {
				return
			}
		}

		p.reading = true
		p.tmpReader = &buf
	}

	n, err = p.tmpReader.Read(data)

	if err != nil {
		p.reading = false
		p.tmpReader = nil
	}

	return n, err
}

func (p *MNSReader) Close() (err error) {
	p.readLocker.Lock()
	defer p.readLocker.Unlock()

	if p.tmpReader != nil {
		p.reading = false
		p.tmpReader = nil
	}
	return
}
