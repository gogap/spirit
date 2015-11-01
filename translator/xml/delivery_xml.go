package xml

import (
	"encoding/xml"
	"time"

	"github.com/gogap/spirit"
)

var _ spirit.Delivery = new(XMLDelivery)

type _XMLDelivery struct {
	XMLName   xml.Name    `xml:"Delivery"`
	Id        string      `xml:"Id"`
	SessionId string      `xml:"SessionId"`
	Payload   _XMLPayload `xml:"Payload"`
	Timestamp time.Time   `xml:"Timestamp"`
}

type XMLDelivery struct {
	id        string
	urn       string
	sessionId string
	labels    map[string]string
	payload   spirit.Payload
	timestamp time.Time
}

func (p *XMLDelivery) Id() string {
	return p.id
}

func (p *XMLDelivery) Payload() spirit.Payload {
	return p.payload
}

func (p *XMLDelivery) URN() string {
	return p.urn
}

func (p *XMLDelivery) SessionId() string {
	return p.sessionId
}

func (p *XMLDelivery) Labels() spirit.Labels {
	return p.labels
}

func (p *XMLDelivery) Validate() (err error) {
	return
}

func (p *XMLDelivery) Timestamp() time.Time {
	return p.timestamp
}
