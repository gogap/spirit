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
}
