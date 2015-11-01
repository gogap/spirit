package line

import (
	"time"

	"github.com/gogap/spirit"
)

var _ spirit.Delivery = new(LineDelivery)

type LineDelivery struct {
	urn       string
	payload   spirit.Payload
	labels    spirit.Labels
	sessionId string
	timestamp time.Time
}

func (p *LineDelivery) Id() string {
	return ""
}

func (p *LineDelivery) URN() string {
	return p.urn
}

func (p *LineDelivery) SessionId() string {
	return p.sessionId
}

func (p *LineDelivery) Payload() spirit.Payload {
	return p.payload
}

func (p *LineDelivery) Labels() spirit.Labels {
	return p.labels
}

func (p *LineDelivery) Validate() (err error) {
	return
}

func (p *LineDelivery) Timestamp() time.Time {
	return p.timestamp
}
