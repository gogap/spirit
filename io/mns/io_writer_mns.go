package mns

import (
	"io"

	"github.com/gogap/spirit"
)

const (
	msnWriterURN = "urn:spirit:writer:mns"
)

var _ io.WriteCloser = new(MNSWriter)

type MNSWriter struct {
}

func init() {
	spirit.RegisterWriter(msnWriterURN, NewMNSWriter)
}

func NewMNSWriter(options spirit.Options) (w io.WriteCloser, err error) {
	return
}

func (p *MNSWriter) Write(data []byte) (n int, err error) {
	return
}

func (p *MNSWriter) Close() (err error) {
	return
}
