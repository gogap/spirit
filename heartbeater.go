package spirit

type Heartbeater interface {
	Name() string
	Start(configFile string) error
	Heartbeat(heartbeatMessage HeartbeatMessage)
}
