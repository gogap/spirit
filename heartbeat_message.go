package spirit

import (
	"time"
)

type HeartbeatMessage struct {
	PID            int
	HostName       string
	InstanceName   string
	StartTime      time.Time
	CurrentTime    time.Time
	HeartbeatCount int64
}
