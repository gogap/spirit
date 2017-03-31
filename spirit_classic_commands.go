package spirit

import (
	"github.com/urfave/cli"
)

func commandComponents(action cli.ActionFunc) cli.Command {
	return cli.Command{
		Name:   "components",
		Usage:  "List the hosting components",
		Action: action,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "v",
				Usage: "Show the component details",
			},
		},
	}
}

func commandRun(action cli.ActionFunc) cli.Command {
	return cli.Command{
		Name:   "run",
		Usage:  "Run the component",
		Action: action,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "name",
				Value: "",
				Usage: "the name of this spirit instance",
			}, cli.StringSliceFlag{
				Name:  "env, e",
				Value: new(cli.StringSlice),
				Usage: "set env to the process",
			}, cli.StringFlag{
				Name:  "config, c",
				Value: "",
				Usage: "the spirit components config",
			}, cli.BoolFlag{
				Name:  "v",
				Usage: "print more internal info",
			}, cli.BoolFlag{
				Name:  "watch, w",
				Usage: "watch all exported events",
			}, cli.StringFlag{
				Name:  "message, m",
				Usage: "Commit message",
			},
		},
	}
}

func commandCreate(action cli.ActionFunc) cli.Command {
	return cli.Command{
		Name:   "create",
		Usage:  "Create the component instance",
		Action: action,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "name",
				Value: "",
				Usage: "the name of this spirit instance",
			}, cli.StringSliceFlag{
				Name:  "env, e",
				Value: new(cli.StringSlice),
				Usage: "set env to the instance process",
			}, cli.StringFlag{
				Name:  "config, c",
				Value: "",
				Usage: "the spirit components config",
			}, cli.BoolFlag{
				Name:  "v",
				Usage: "print more internal info",
			}, cli.StringFlag{
				Name:  "message, m",
				Usage: "Commit message",
			},
		},
	}
}

func commandCall(action cli.ActionFunc) cli.Command {
	return cli.Command{
		Name:   "call",
		Usage:  "Call the resgistered function of component",
		Action: action,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "name, n",
				Value: "",
				Usage: "Component name",
			}, cli.StringFlag{
				Name:  "handler",
				Value: "",
				Usage: "The name of handler to be call",
			}, cli.StringFlag{
				Name:  "payload, l",
				Value: "",
				Usage: "The json data file path of spirit.Payload struct",
			}, cli.BoolFlag{
				Name:  "json, j",
				Usage: "Format the result into json",
			},
		},
	}
}

func commandRemoveInstance(action cli.ActionFunc) cli.Command {
	return cli.Command{
		Name:      "rmi",
		ShortName: "",
		Usage:     "Remove whole instance",
		Action:    action,
	}
}

func commandRemoveBinary(action cli.ActionFunc) cli.Command {
	return cli.Command{
		Name:      "rm",
		ShortName: "",
		Usage:     "Remove instance binary",
		Action:    action,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "hash",
				Usage: "Binary hash",
			},
		},
	}
}

func commandStart(action cli.ActionFunc) cli.Command {
	return cli.Command{
		Name:      "start",
		ShortName: "",
		Usage:     "Start a named spirit instance",
		Action:    action,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "v",
				Usage: "Show details",
			}, cli.StringFlag{
				Name:  "hash",
				Usage: "Binary hash",
			}, cli.BoolFlag{
				Name:  "interactive, i",
				Usage: "Attach instance's STDIN",
			}, cli.BoolFlag{
				Name:  "attach, a",
				Usage: "Attach STDOUT/STDERR and forward signals",
			},
		},
	}
}

func commandStop(action cli.ActionFunc) cli.Command {
	return cli.Command{
		Name:      "stop",
		ShortName: "",
		Usage:     "Stop a named spirit instance",
		Action:    action,
	}
}

func commandKill(action cli.ActionFunc) cli.Command {
	return cli.Command{
		Name:      "kill",
		ShortName: "",
		Usage:     "Kill a named spirit instance",
		Action:    action,
	}
}

func commandPause(action cli.ActionFunc) cli.Command {
	return cli.Command{
		Name:      "pause",
		ShortName: "",
		Usage:     "Pause or Resume a named spirit instance",
		Action:    action,
	}
}

func commandPS(action cli.ActionFunc) cli.Command {
	return cli.Command{
		Name:      "ps",
		ShortName: "",
		Usage:     "Show spirit instance",
		Action:    action,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "all, a",
				Usage: "Show all process, running and exited",
			}, cli.BoolFlag{
				Name:  "quiet, q",
				Usage: " Only display spirit instance name",
			},
		},
	}
}

func commandInspect(action cli.ActionFunc) cli.Command {
	return cli.Command{
		Name:      "inspect",
		ShortName: "",
		Usage:     "Inspect a named spirit instance",
		Action:    action,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "metadata",
				Usage: "Show the instance runtime metadata",
			}, cli.BoolFlag{
				Name:  "config",
				Usage: "Show the instance built configs",
			},
		},
	}
}

func commandCommit(action cli.ActionFunc) cli.Command {
	return cli.Command{
		Name:      "commit",
		ShortName: "",
		Usage:     "Commit binary into instance's cache",
		Action:    action,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "message, m",
				Usage: "Commit message",
			},
		},
	}
}

func commandCheckout(action cli.ActionFunc) cli.Command {
	return cli.Command{
		Name:      "checkout",
		ShortName: "",
		Usage:     "Checkout the instance binary by hash",
		Action:    action,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "v",
				Usage: "Show details",
			}, cli.StringFlag{
				Name:  "hash",
				Usage: "Binary hash",
			},
		},
	}
}

func commandInstanceBins(action cli.ActionFunc) cli.Command {
	return cli.Command{
		Name:      "bins",
		ShortName: "",
		Usage:     "List instance binary committed",
		Action:    action,
	}
}

func commandVersion(action cli.ActionFunc) cli.Command {
	return cli.Command{
		Name:      "version",
		ShortName: "",
		Usage:     "Get current binary version",
		Action:    action,
	}
}
