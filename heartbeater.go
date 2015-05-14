package spirit

type Heartbeater interface {
	Name() string
	Start(options Options) error
	Heartbeat(heartbeatMessage HeartbeatMessage)
}
