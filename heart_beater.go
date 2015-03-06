package spirit

type HeartBeater interface {
	Name() string
	Start() error
	HeartBeat(heartBeatMessage HeartBeatMessage)
}
