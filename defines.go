package spirit

type Status int32

const (
	StatusStopped Status = 0
	StatusRunning Status = 1
	StatusPaused  Status = 2
)

type Labels map[string]string

const (
	SpiritInitialLogLevelEnvKey = "SPIRIT_INIT_LOG_LEVEL"
)
