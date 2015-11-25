package spirit

import (
	"time"
)

type Delivery interface {
	Id() string
	SessionId() string

	URN() string
	SetURN(urn string) (err error)

	Labels() Labels
	SetLabel(label string, value string) (err error)
	SetLabels(labels Labels) (err error)

	Payload() Payload
	Validate() (err error)
	Timestamp() time.Time

	GetMetadata(name string) (v interface{}, exist bool)
	SetMetadata(name string, v interface{}) (err error)
	Metadata() (metadata Map)
	DeleteMetadata(name string) (err error)
}
