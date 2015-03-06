package spirit

import (
	"time"
)

type HeartBeatMessage struct {
	PID            int
	HostName       string
	Component      string
	StartTime      time.Time
	CurrentTime    time.Time
	HeartBeatCount int64
}
