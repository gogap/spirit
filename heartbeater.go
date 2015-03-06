package spirit

type Heartbeater interface {
	Name() string
	Start() error
	Heartbeat(heartbeatMessage HeartbeatMessage)
}
