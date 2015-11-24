package json

import (
	"time"

	"github.com/gogap/spirit"
)

var _ spirit.Delivery = new(JSONDelivery)

type _JSONDelivery struct {
	Id        string          `json:"id"`
	URN       string          `json:"urn"`
	SessionId string          `json:"session_id"`
	Payload   _JSONPayload    `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
	Metadata  spirit.Metadata `json:"metadata"`
}

type JSONDelivery struct {
	id        string
	urn       string
	sessionId string
	labels    spirit.Labels
	payload   spirit.Payload
	timestamp time.Time
	metadata  spirit.Metadata
}

func (p *JSONDelivery) Id() string {
	return p.id
}

func (p *JSONDelivery) URN() string {
	return p.urn
}

func (p *JSONDelivery) SessionId() string {
	return p.sessionId
}

func (p *JSONDelivery) Labels() spirit.Labels {
	return p.labels
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

func (p *JSONDelivery) Metadata() (metadata spirit.Metadata) {
	metadata = p.metadata
	return
}

func (p *JSONDelivery) DeleteContext(name string) (err error) {
	if p.metadata != nil {
		delete(p.metadata, name)
	}
	return
}
