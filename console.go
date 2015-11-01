package spirit

type NewConsoleFunc func(options Options) (console Console, err error)

type Console interface {
	StartStoper
}

func RegisterConsole(urn string, newConsoleFunc NewConsoleFunc) (err error) {

	logger.WithField("module", "spirit").
		WithField("event", "register console").
		WithField("urn", urn).
		Debugln("console registered")
	return
}
