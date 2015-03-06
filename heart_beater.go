package spirit

type HeartBeater interface {
	Name() string
	HeartBeat(heartBeatMessage HeartBeatMessage)
}
