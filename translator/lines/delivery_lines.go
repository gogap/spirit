package lines

import (
	"sync"
	"time"

	"github.com/gogap/spirit"
)

var _ spirit.Delivery = new(LinesDelivery)

type LinesDelivery struct {
	urn       string
	payload   spirit.Payload
	labels    spirit.Labels
	sessionId string
	timestamp time.Time
	metadata  spirit.Map

	labelsLocker sync.Mutex
}

func (p *LinesDelivery) Id() string {
	return ""
}

func (p *LinesDelivery) URN() string {
	return p.urn
}

func (p *LinesDelivery) SetURN(urn string) (err error) {
	p.urn = urn
	return
}

func (p *LinesDelivery) SessionId() string {
	return p.sessionId
}

func (p *LinesDelivery) Payload() spirit.Payload {
	return p.payload
}

func (p *LinesDelivery) Labels() spirit.Labels {
	return p.labels
}

func (p *LinesDelivery) SetLabel(label string, value string) (err error) {
	p.labelsLocker.Lock()
	p.labelsLocker.Unlock()

	if p.labels == nil {
		p.labels = make(spirit.Labels)
		return
	}

	p.labels[label] = value

	return
}
func (p *LinesDelivery) SetLabels(labels spirit.Labels) (err error) {
	p.labelsLocker.Lock()
	p.labelsLocker.Unlock()

	p.labels = labels
	return
}

func (p *LinesDelivery) Validate() (err error) {
	return
}

func (p *LinesDelivery) Timestamp() time.Time {
	return p.timestamp
}

func (p *LinesDelivery) GetMetadata(name string) (v interface{}, exist bool) {
	if p.metadata == nil {
		return
	}

	v, exist = p.metadata[name]

	return
}

func (p *LinesDelivery) SetMetadata(name string, v interface{}) (err error) {
	p.metadata[name] = v
	return
}

func (p *LinesDelivery) Metadata() (metadata spirit.Map) {
	metadata = p.metadata
	return
}

func (p *LinesDelivery) DeleteMetadata(name string) (err error) {
	if p.metadata != nil {
		delete(p.metadata, name)
	}
	return
}
