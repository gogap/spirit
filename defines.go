package spirit

type Status int32

const (
	StatusStopped Status = 0
	StatusRunning        = 1
	StatusPaused         = 2
)

type Labels map[string]string
