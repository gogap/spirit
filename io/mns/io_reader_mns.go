package mns

import (
	"io"

	"github.com/gogap/spirit"
)

const (
	msnReaderURN = "urn:spirit:io:reader:mns"
)

var _ io.ReadCloser = new(MNSReader)

type MNSReader struct {
}

func init() {
	spirit.RegisterReader(msnReaderURN, NewMNSReader)
}

func NewMNSReader(options spirit.Options) (r io.ReadCloser, err error) {
	return
}

func (p *MNSReader) Read(data []byte) (n int, err error) {
	return
}

func (p *MNSReader) Close() (err error) {
	return
}
