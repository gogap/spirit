package json

import (
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
}

type JSONDelivery struct {
	id        string
	urn       string
	sessionId string
	labels    spirit.Labels
	payload   spirit.Payload
	timestamp time.Time
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
