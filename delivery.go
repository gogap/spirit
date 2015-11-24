package spirit

import (
	"time"
)

type Delivery interface {
	Id() string
	URN() string
	SessionId() string
	Labels() Labels
	Payload() Payload
	Validate() (err error)
	Timestamp() time.Time

	GetMetadata(name string) (v interface{}, exist bool)
	SetMetadata(name string, v interface{}) (err error)
	Metadata() (metadata Metadata)
	DeleteContext(name string) (err error)
}
