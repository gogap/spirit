package mns

import (
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

	tmpReader io.ReadCloser
}

func init() {
	spirit.RegisterReader(mnsReaderURN, NewMNSReader)
}

func NewMNSReader(options spirit.Options) (r io.ReadCloser, err error) {
	conf := MNSReaderConfig{}
	if err = options.ToObject(&conf); err != nil {
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

		if resp.StatusCode != http.StatusOK {
			err = ErrAliMNSResponseError
			return
		}

		p.reading = true
		p.tmpReader = resp.Body
	}

	n, err = p.tmpReader.Read(data)
	if err == io.EOF {
		p.reading = false
		p.tmpReader.Close()
		p.tmpReader = nil
	}

	return n, err
}

func (p *MNSReader) Close() (err error) {
	return
}
