package json

import (
	"sync"
	"time"

	"github.com/gogap/spirit"
)

var _ spirit.Delivery = new(JSONDelivery)

type _JSONDelivery struct {
	Id        string       `json:"id"`
	URN       string       `json:"urn"`
	SessionId string       `json:"session_id"`
	Payload   _JSONPayload `json:"payload"`
	Timestamp time.Time    `json:"timestamp"`
	Metadata  spirit.Map   `json:"metadata"`
}

type JSONDelivery struct {
	id        string
	urn       string
	sessionId string
	labels    spirit.Labels
	payload   spirit.Payload
	timestamp time.Time
	metadata  spirit.Map

	labelsLocker sync.Mutex
}

func (p *JSONDelivery) Id() string {
	return p.id
}

func (p *JSONDelivery) URN() string {
	return p.urn
}

func (p *JSONDelivery) SetURN(urn string) (err error) {
	p.urn = urn
	return
}

func (p *JSONDelivery) SessionId() string {
	return p.sessionId
}

func (p *JSONDelivery) Labels() spirit.Labels {
	return p.labels
}

func (p *JSONDelivery) SetLabel(label string, value string) (err error) {
	p.labelsLocker.Lock()
	p.labelsLocker.Unlock()

	if p.labels == nil {
		p.labels = make(spirit.Labels)
		return
	}

	p.labels[label] = value

	return
}
func (p *JSONDelivery) SetLabels(labels spirit.Labels) (err error) {
	p.labelsLocker.Lock()
	p.labelsLocker.Unlock()

	p.labels = labels
	return
}

func (p *JSONDelivery) Payload() spirit.Payload {
	return p.payload
}

func (p *JSONDelivery) Validate() (err error) {
	return
}

func (p *JSONDelivery) Timestamp() time.Time {
	return p.timestamp
}

func (p *JSONDelivery) GetMetadata(name string) (v interface{}, exist bool) {
	if p.metadata == nil {
		return
	}

	v, exist = p.metadata[name]

	return
}

func (p *JSONDelivery) SetMetadata(name string, v interface{}) (err error) {
	p.metadata[name] = v
	return
}

func (p *JSONDelivery) Metadata() (metadata spirit.Map) {
	metadata = p.metadata
	return
}

func (p *JSONDelivery) DeleteMetadata(name string) (err error) {
	if p.metadata != nil {
		delete(p.metadata, name)
	}
	return
}
