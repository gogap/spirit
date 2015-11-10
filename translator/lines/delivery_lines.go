package lines

import (
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
}

func (p *LinesDelivery) Id() string {
	return ""
}

func (p *LinesDelivery) URN() string {
	return p.urn
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

func (p *LinesDelivery) Validate() (err error) {
	return
}

func (p *LinesDelivery) Timestamp() time.Time {
	return p.timestamp
}
