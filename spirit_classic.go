package spirit

import (
	"encoding/json"
	"fmt"
	"github.com/gogap/logs"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/codegangsta/cli"
	"github.com/gogap/env_strings"
)

const (
	EXT_SPIRIT = ".spirit"

	ENV_NAME = "SPIRIT_ENV"
	ENV_EXT  = ".env"
)

type ClassicSpirit struct {
	cliApp *cli.App

	receiverFactory MessageReceiverFactory
	senderFactory   MessageSenderFactory

	components           map[string]Component
	runningComponent     Component
	runningComponentConf string

	heartbeaters       map[string]Heartbeater
	heartbeatSleepTime time.Duration
	heartbeatersToRun  map[string]bool
	heartbeaterConfig  string
	alias              string
	envs               []string

	isBuilt                       bool
	isRunCommand                  bool
	isBuildCheckOnly              bool
	isCreatAliasdSpirtContextOnly bool
	isRunCheckedCorrect           bool

	lockfile *LockFile
}

func NewClassicSpirit(name, description, version string) Spirit {
	newSpirit := new(ClassicSpirit)

	app := cli.NewApp()
	app.Name = name
	app.Usage = description
	app.Commands = newSpirit.commands()
	app.Email = ""
	app.Author = ""
	app.Version = version
	app.EnableBashCompletion = true

	newSpirit.cliApp = app
	newSpirit.components = make(map[string]Component, 0)
	newSpirit.heartbeaters = make(map[string]Heartbeater, 0)
	newSpirit.heartbeatersToRun = make(map[string]bool, 0)

	receiverFactory := NewDefaultMessageReceiverFactory()
	receiverFactory.RegisterMessageReceivers(new(MessageReceiverMQS))

	senderFactory := NewDefaultMessageSenderFactory()
	senderFactory.RegisterMessageSenders(new(MessageSenderMQS))

	newSpirit.receiverFactory = receiverFactory
	newSpirit.senderFactory = senderFactory

	newSpirit.RegisterHeartbeaters(new(AliJiankong))

	return newSpirit
}

func (p *ClassicSpirit) commands() []cli.Command {

	return []cli.Command{
		{
			Name:      "component",
			ShortName: "",
			Usage:     "( *** this command obsolete next version) options for component",
			Subcommands: []cli.Command{
				{
					Name:   "list",
					Usage:  "list the hosting components",
					Action: p.cmdListComponent,
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "func, f",
							Usage: "show the component functions",
						},
					},
				},
				{
					Name:   "run",
					Usage:  "run the component",
					Action: p.cmdRunComponent,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "name, n",
							Value: "",
							Usage: "the name of component to run",
						}, cli.StringSliceFlag{
							Name:  "address, a",
							Value: new(cli.StringSlice),
							Usage: "the address of component receiver, format: receiverType|url|configFile",
						}, cli.StringSliceFlag{
							Name:  "heartbeat, hb",
							Value: new(cli.StringSlice),
							Usage: "run the register heatbeater while heartbeat-timeout greater than 0",
						}, cli.IntFlag{
							Name:  "heartbeat-sleeptime, hbs",
							Value: 0,
							Usage: "the heartbeat sleep time, default = 0 (Millisecond), disabled",
						}, cli.StringFlag{
							Name:  "heartbeat-config, hbc",
							Value: "",
							Usage: "the config file of heartbeat",
						}, cli.StringFlag{
							Name:  "alias",
							Value: "",
							Usage: "if the alias did not empty, it will be singleton process by alias",
						}, cli.BoolFlag{
							Name:  "build-check, bc",
							Usage: "build check only, it won't really run",
						}, cli.BoolFlag{
							Name:  "create-only, co",
							Usage: "create aliasd component of spirit only",
						}, cli.StringFlag{
							Name:  "config",
							Value: "",
							Usage: "the config file path for inital func to use",
						},
					},
				},
				{
					Name:   "call",
					Usage:  "call the component in port handler",
					Action: p.cmdCallHandler,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "name, n",
							Value: "",
							Usage: "component name",
						}, cli.StringFlag{
							Name:  "handler",
							Value: "",
							Usage: "the name of handler to be call",
						}, cli.StringFlag{
							Name:  "payload, l",
							Value: "",
							Usage: "the json data file path of spirit.Payload struct",
						}, cli.BoolFlag{
							Name:  "json, j",
							Usage: "format the result into json",
						},
					},
				},
			},
		}, {
			Name:   "run",
			Usage:  "Run the component",
			Action: p.cmdRunComponent,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name, n",
					Value: "",
					Usage: "the name of component to run",
				}, cli.StringSliceFlag{
					Name:  "address, a",
					Value: new(cli.StringSlice),
					Usage: "the address of component receiver, format: receiverType|url|configFile",
				}, cli.StringSliceFlag{
					Name:  "heartbeat, hb",
					Value: new(cli.StringSlice),
					Usage: "run the register heatbeater while heartbeat-timeout greater than 0",
				}, cli.IntFlag{
					Name:  "heartbeat-sleeptime, hbs",
					Value: 0,
					Usage: "the heartbeat sleep time, default = 0 (Millisecond), disabled",
				}, cli.StringFlag{
					Name:  "heartbeat-config, hbc",
					Value: "",
					Usage: "the config file of heartbeat",
				}, cli.StringFlag{
					Name:  "alias",
					Value: "",
					Usage: "if the alias did not empty, it will be singleton process by alias",
				}, cli.BoolFlag{
					Name:  "build-check, bc",
					Usage: "build check only, it won't really run",
				}, cli.BoolFlag{
					Name:  "create-only, co",
					Usage: "create aliasd component of spirit only",
				}, cli.StringFlag{
					Name:  "config",
					Value: "",
					Usage: "the config file path for inital func to use",
				}, cli.StringSliceFlag{
					Name:  "env, e",
					Value: new(cli.StringSlice),
					Usage: "set env to the process",
				},
			},
		},
		{
			Name:   "components",
			Usage:  "List the hosting components",
			Action: p.cmdListComponent,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "func, f",
					Usage: "Show the component functions",
				},
			},
		}, {
			Name:   "call",
			Usage:  "Call the resgistered function of component",
			Action: p.cmdCallHandler,
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
		},
		{
			Name:      "ps",
			ShortName: "",
			Usage:     "Show spirit process which own alias",
			Action:    p.showProcess,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "all, a",
					Usage: "Show all process, running and exited",
				},
			},
		}, {
			Name:      "start",
			ShortName: "",
			Usage:     "Start the process created by run command and own alias",
			Action:    p.startProcess,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "std",
					Usage: "use the current std to the process",
				}, cli.StringSliceFlag{
					Name:  "env, e",
					Value: new(cli.StringSlice),
					Usage: "set env to the process",
				},
			},
		}, {
			Name:      "stop",
			ShortName: "",
			Usage:     "Stop the running process created by run command and own alias",
			Action:    p.stopProcess,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "alias, a",
					Usage: "Running component of spirit with alias",
				},
			},
		}, {
			Name:      "restart",
			ShortName: "",
			Usage:     "Restart the running process created by run command and own alias",
			Action:    p.restartProcess,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "alias, a",
					Usage: "Restart the running component of spirit with alias",
				}, cli.BoolFlag{
					Name:  "std",
					Usage: "use the current std to the process",
				}, cli.StringSliceFlag{
					Name:  "env, e",
					Value: new(cli.StringSlice),
					Usage: "set env to the process",
				},
			},
		},
	}
}

func (p *ClassicSpirit) cmdRunComponent(c *cli.Context) {
	p.isRunCommand = true
	componentName := c.String("name")
	p.envs = c.StringSlice("env")

	if p.envs != nil {
		for _, env := range p.envs {
			kv := strings.Split(env, "=")
			if len(kv) != 2 {
				fmt.Printf("[spirit] component %s - %s env %s error\n", p.runningComponent.Name(), p.alias, env)
				return
			} else {
				os.Setenv(kv[0], kv[1])
			}
		}
	}

	receiverAddrs := []string{}

	if tmpAddrs := c.StringSlice("address"); tmpAddrs != nil {
		envStrings := env_strings.NewEnvStrings(ENV_NAME, ENV_EXT)
		for _, addr := range tmpAddrs {
			if recvAddr, e := envStrings.Execute(addr); e != nil {
				fmt.Printf("[spirit] could not execute address with env config of %s, original addr: %s, error: %s \n", ENV_NAME, addr, e)
				return
			} else {
				receiverAddrs = append(receiverAddrs, recvAddr)
			}
		}
	}

	if len(receiverAddrs) == 0 {
		fmt.Println("[spirit] receiver address list not set")
		return
	}

	heartbeatersToRun := c.StringSlice("heartbeat")
	heartbeatSleepTime := c.Int("heartbeat-sleeptime")
	heartbeaterConfig := c.String("heartbeat-config")

	p.alias = c.String("alias")
	p.isBuildCheckOnly = c.Bool("build-check")
	p.isCreatAliasdSpirtContextOnly = c.Bool("create-only")
	p.runningComponentConf = c.String("config")

	p.heartbeatSleepTime = time.Millisecond * time.Duration(heartbeatSleepTime)

	if heartbeatersToRun != nil && p.heartbeatSleepTime > 0 {
		for _, heartbeaterName := range heartbeatersToRun {
			heartbeaterName = strings.TrimSpace(heartbeaterName)
			if _, exist := p.heartbeaters[heartbeaterName]; exist {
				p.heartbeatersToRun[heartbeaterName] = true
			}
		}
	}

	p.heartbeaterConfig = heartbeaterConfig

	tmpUrlUsed := map[string]string{} //one type:url could only use by one component-port

	var component Component
	for _, receiverAddr := range receiverAddrs {
		receiverType := ""
		receiverConfig := ""
		receiverUrl := ""
		portName := ""
		handlerName := ""

		if componentName == "" {
			fmt.Println("[spirit] component name is empty.")
			return
		}

		if receiverAddr == "" {
			fmt.Println("[spirit] address is empty.")
			return
		}

		addr := strings.Split(receiverAddr, "|")

		if len(addr) == 4 {
			portName = addr[0]
			handlerName = addr[1]
			receiverType = addr[2]
			receiverUrl = addr[3]
		} else if len(addr) == 5 {
			portName = addr[0]
			handlerName = addr[1]
			receiverType = addr[2]
			receiverUrl = addr[3]
			receiverConfig = addr[4]
		} else {
			fmt.Println("[spirit] address format error. example: port.in|delete|mqs|http://xxxx.com/queue?param=1|/etc/a.conf")
			return
		}

		if portName == "" {
			fmt.Println("[spirit] receiver port name is empty.")
			return
		}

		if handlerName == "" {
			fmt.Println("[spirit] handler name is empty.")
			return
		}

		if receiverType == "" {
			fmt.Println("[spirit] receiver type is empty.")
			return
		}

		if receiverUrl == "" {
			fmt.Println("[spirit] receiver url is empty.")
			return
		}

		if comp, exist := p.components[componentName]; !exist {
			fmt.Printf("[spirit] component %s does not hosting.\n", componentName)
			return
		} else {
			component = comp
		}

		usedItemValue := component.Name() + ":" + portName
		usedItemKey := receiverType + "|" + receiverUrl

		if v, exist := tmpUrlUsed[usedItemKey]; exist {
			if usedItemValue != v {
				fmt.Printf("[spirit] one address url only could be used by one component port, the used component is: %s\nurl:%s\n", usedItemValue, receiverUrl)
				return
			}
		} else {
			tmpUrlUsed[usedItemKey] = v
		}

		if !p.receiverFactory.IsExist(receiverType) {
			fmt.Printf("[spirit] the receiver type of %s does not registered.", receiverType)
			return
		}

		if receiver, e := p.receiverFactory.NewReceiver(receiverType, receiverUrl, receiverConfig); e != nil {
			fmt.Println(e)
			return
		} else {
			component.BindHandler(portName, handlerName).
				BindReceiver(portName, receiver)
		}
	}
	component.SetMessageSenderFactory(p.senderFactory).Build()

	p.runningComponent = component

	p.isRunCheckedCorrect = true
}

func (p *ClassicSpirit) cmdListComponent(c *cli.Context) {
	showDetails := c.Bool("func")

	for _, component := range p.components {
		fmt.Println(component.Name())
		if showDetails {
			if handlers, e := component.ListHandlers(); e != nil {
				fmt.Println("[spirit] "+component.Name()+":", e.Error())
			} else {
				for name, _ := range handlers {
					fmt.Println("\t", name)
				}
			}
		}
	}
}

func (p *ClassicSpirit) cmdCallHandler(c *cli.Context) {
	componentName := c.String("name")
	handlerName := c.String("handler")
	payloadFile := c.String("payload")
	toJson := c.Bool("json")

	if componentName == "" {
		fmt.Println("[spirit] component name is empty.")
		return
	}

	if handlerName == "" {
		fmt.Println("[spirit] handler name is empty.")
		return
	}

	var component Component
	if comp, exist := p.components[componentName]; !exist {
		fmt.Printf("[spirit] component %s does not hosting.\n", componentName)
		return
	} else {
		component = comp
	}

	var bPayload []byte
	if payloadFile != "" {
		if data, e := ioutil.ReadFile(payloadFile); e != nil {
			fmt.Println("[spirit] reade payload file error.", e)
			return
		} else {
			bPayload = data
		}
	}

	payload := Payload{}

	if payloadFile != "" {
		if e := payload.UnSerialize(bPayload); e != nil {
			fmt.Println("[spirit] parse payload file failed, please make sure it is json format", e)
			return
		}
	}

	if result, e := component.CallHandler(handlerName, &payload); e != nil {
		fmt.Println(e)
	} else {
		if toJson {
			if result != nil {
				if b, e := json.MarshalIndent(result, "", " "); e != nil {
					fmt.Println("[spirit] format result to json failed.", e)
				} else {
					fmt.Println(string(b))
				}
			} else {
				fmt.Println(result)
			}
		} else {
			fmt.Println(result)
		}
	}
}

func (p *ClassicSpirit) SetMessageReceiverFactory(factory MessageReceiverFactory) {
	if factory == nil {
		panic("message receiver factory could not be nil")
	}
	p.receiverFactory = factory
}

func (p *ClassicSpirit) GetMessageReceiverFactory() MessageReceiverFactory {
	return p.receiverFactory
}

func (p *ClassicSpirit) SetMessageSenderFactory(factory MessageSenderFactory) {
	if factory == nil {
		panic("message sender factory could not be nil")
	}
	p.senderFactory = factory
}

func (p *ClassicSpirit) GetMessageSenderFactory() MessageSenderFactory {
	return p.senderFactory
}

func (p *ClassicSpirit) Hosting(components ...Component) Spirit {
	if components == nil || len(components) == 0 {
		panic("components is nil or empty")
	}

	for _, component := range components {
		if component == nil {
			panic("component is nil")
		}
		p.components[component.Name()] = component
	}
	return p
}

func (p *ClassicSpirit) RegisterHeartbeaters(beaters ...Heartbeater) Spirit {
	if beaters == nil || len(beaters) == 0 {
		return p
	}

	for _, beater := range beaters {
		if _, exist := p.heartbeaters[beater.Name()]; exist {
			panic(fmt.Sprintf("heart beater %s already exist", beater.Name()))
		}
		p.heartbeaters[beater.Name()] = beater
	}
	return p
}

func (p *ClassicSpirit) RemoveHeartBeaters(names ...string) Spirit {
	if names == nil || len(names) == 0 {
		return p
	}

	if p.heartbeaters == nil {
		return p
	}

	for _, name := range names {
		delete(p.heartbeaters, name)
	}

	return p
}

func (p *ClassicSpirit) Build() Spirit {
	p.cliApp.Run(os.Args)
	p.isBuilt = true
	return p
}

func (p *ClassicSpirit) GetComponent(name string) Component {
	if !p.isBuilt {
		panic("please build components first")
	}

	if component, exist := p.components[name]; !exist {
		panic(fmt.Sprintf("component of %s did not exist.", name))
	} else {
		return component
	}
}

func (p *ClassicSpirit) getHeartbeatMessage() (message HeartbeatMessage) {
	hostName := ""
	if name, e := os.Hostname(); e != nil {
		panic(e)
	} else {
		hostName = name
	}

	message.Component = p.runningComponent.Name()
	message.HostName = hostName
	message.StartTime = time.Now()
	message.PID = syscall.Getpid()

	return
}

func (p *ClassicSpirit) Run(initalFuncs ...InitalFunc) {
	if p.isRunCommand && p.isRunCheckedCorrect {
		if !p.isBuilt {
			fmt.Println("[spirit] spirit should build first")
			return
		}

		if p.isBuildCheckOnly {
			fmt.Println("[spirit] it seems all correct")
			return
		}

		//if alias did not empty ,it will enter singleton mode
		if e := p.lock(); e != nil && p.alias != "" {
			fmt.Printf("[spirit] component %s - %s already running, pid: %d\n", p.runningComponent.Name(), p.alias, p.getPID())
			return
		}

		if p.isCreatAliasdSpirtContextOnly {
			if p.alias == "" {
				fmt.Printf("[spirit] please input alias first\n")
				return
			}
			fmt.Printf("[spirit] the context of spirit component %s alias named %s was created\n", p.runningComponent.Name(), p.alias)
			return
		}

		//run inital funcs
		if initalFuncs != nil {
			for _, initFunc := range initalFuncs {
				if e := initFunc(p.runningComponentConf); e != nil {
					panic(e)
				}
			}
		}

		//start heartbeaters
		if p.heartbeatSleepTime > 0 {
			heartbeatMessage := p.getHeartbeatMessage()
			for name, _ := range p.heartbeatersToRun {
				if heartbeater, exist := p.heartbeaters[name]; exist {
					if e := heartbeater.Start(p.heartbeaterConfig); e != nil {
						panic(e)
					}
					fmt.Printf("[spirit] heartbeater %s running\n", heartbeater.Name())
					go func(beater Heartbeater, msg HeartbeatMessage, sleepTime time.Duration) {
						for {
							time.Sleep(sleepTime)
							msg.CurrentTime = time.Now()
							msg.HeartbeatCount += 1
							beater.Heartbeat(msg)
						}
					}(heartbeater, heartbeatMessage, p.heartbeatSleepTime)
				} else {
					fmt.Printf("[spirit] heartbeater %s not exist\n", name)
					return
				}
			}
		}

		p.runningComponent.Run()
		fmt.Printf("[spirit] component %s running\n", p.runningComponent.Name())

		p.waitSignal()
	}
}

func (p *ClassicSpirit) waitSignal() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGUSR1)

	for {
		select {
		case signal := <-interrupt:
			logs.Info("singal received", signal)
			switch signal {
			case os.Interrupt, syscall.SIGTERM:
				{
					go p.runningComponent.Stop()
					time.Sleep(time.Second)

					i := 0
					for p.runningComponent.Status() != STATUS_STOPED {
						i++
						fmt.Printf("\r[spirit] waiting for running component %s to stop, %d sec", p.runningComponent.Name(), i)
						time.Sleep(time.Second)
					}
					logs.Info(fmt.Sprintf("[spirit] component %s was gracefully stoped\n", p.runningComponent.Name()))
					time.Sleep(time.Second)
					os.Exit(0)
				}
			case os.Kill:
				{
					logs.Info(fmt.Sprintf("[spirit] component %s was killed\n", p.runningComponent.Name()))
					time.Sleep(time.Second)
					os.Exit(128)
				}
			case syscall.SIGUSR1:
				{
					p.runningComponent.PauseOrResume()
					if p.runningComponent.Status() == STATUS_PAUSED {
						logs.Info(fmt.Sprintf("[spirit] component %s was paused\n", p.runningComponent.Name()))
					} else {
						logs.Info(fmt.Sprintf("[spirit] component %s was resumed\n", p.runningComponent.Name()))
					}
				}
			}
		}
	}
}

func (p *ClassicSpirit) showProcess(c *cli.Context) {
	showAll := c.Bool("all")

	contents := []LockFileContent{}

	home := GetComponentHome(p.cliApp.Name)

	if !IsFileOrDir(home, true) {
		return
	}

	if f, e := os.Open(home); e != nil {
		return
	} else if names, e := f.Readdirnames(-1); e == nil {
		for _, name := range names {
			if filepath.Ext(name) == EXT_SPIRIT {
				if lockfile, e := OpenLockFile(home+"/"+name, 0640); e != nil {
					fmt.Println("[spirit] open spirit context file failed, error:", e)
				} else if content, e := lockfile.ReadContent(); e != nil {
					fmt.Println("[spirit] error context: ", home+"/"+name)
				} else {
					if IsProcessAlive(content.PID) || showAll {
						contents = append(contents, content)
					}
				}
			}
		}
	}

	if data, e := json.MarshalIndent(contents, " ", "  "); e != nil {
		fmt.Printf("[spirit] format contents to json failed, error:\n", e)
		return
	} else {
		fmt.Println(string(data))
	}
}

func (p *ClassicSpirit) startProcess(c *cli.Context) {
	//TODO improve process logic, it was dirty implament currently

	if args := c.Args(); args == nil || len(args) == 0 {
		fmt.Println("[spirit] please input alias name first")
		return
	} else {
		p.alias = strings.TrimSpace(args[0])
	}

	if p.alias == "" {
		fmt.Println("[spirit] please input alias name first")
		return
	}

	useSTD := c.Bool("std")
	extEnvs := c.StringSlice("env")

	for _, env := range extEnvs {
		if v := strings.Split(env, "="); len(v) != 2 {
			fmt.Println("[spirit] env params format error, e.g.: ENV_KEY='value'")
			return
		}
	}

	home := GetComponentHome(p.cliApp.Name)

	if !IsFileOrDir(home, true) {
		return
	}

	lockfilePath := p.getLockeFileName()

	if _, err := os.Stat(lockfilePath); os.IsNotExist(err) {
		fmt.Printf("[spirit] component of spirit %s not created before\n", p.alias)
		return
	}

	if lockfile, e := OpenLockFile(lockfilePath, 0640); e != nil {
		fmt.Printf("[spirit] open spirit of %s context failed, error: %s\n", p.alias, e)
		return
	} else if content, e := lockfile.ReadContent(); e != nil {
		fmt.Printf("[spirit] read spirit of %s context failed, path: %s, error: %s\n", p.alias, lockfilePath, e)
		return
	} else {
		if IsProcessAlive(content.PID) {
			fmt.Printf("[spirit] spirit of %s already running, pid: %d\n", p.alias, content.PID)
			return
		}

		newProcessArgs := []string{}

		if argsI, exist := content.Context["args"]; !exist {
			fmt.Printf("[spirit] spirit of %s's context damage please use command of component run to recreate instance\n", p.alias)
			return
		} else if args, ok := argsI.([]interface{}); ok {
			for i := 0; i < len(args); i++ {
				if strArg, ok := args[i].(string); ok {
					newProcessArgs = append(newProcessArgs, strArg)
				} else {
					fmt.Printf("[spirit] spirit of %s's context damage please use command of component run to recreate instance\n", p.alias)
					return
				}
			}
		}

		originEnvs := []string{}
		if envI, exist := content.Context["envs"]; exist && envI != nil {
			if envs, ok := envI.([]interface{}); ok {
				for i := 0; i < len(envs); i++ {
					if env, ok := envs[i].(string); ok {
						originEnvs = append(originEnvs, env)
					} else {
						fmt.Printf("[spirit] spirit of %s's context damage please use command of component run to recreate instance\n", p.alias)
						return
					}
				}
			} else {
				fmt.Printf("[spirit] spirit of %s's context damage please use command of component run to recreate instance\n", p.alias)
				return
			}
		}

		absPath := ""
		if path, e := filepath.Abs(os.Args[0]); e != nil {
			fmt.Printf("[spirit] start spirit of %s failed, error: %s\n", p.alias, e)
			return
		} else {
			absPath = path
		}

		if pid, e := StartProcess(absPath, newProcessArgs, originEnvs, useSTD, extEnvs...); e != nil {
			fmt.Printf("[spirit] start spirit of %s failed, error: %s\n", p.alias, e)
			return
		} else {
			fmt.Printf("[spirit] start spirit of %s success, pid: %d\n", p.alias, pid)
			return
		}
	}
}

func (p *ClassicSpirit) stopProcess(c *cli.Context) {
	//TODO improve process logic, it was dirty implament currently
	if args := c.Args(); args == nil || len(args) == 0 {
		fmt.Println("[spirit] please input alias name first")
		return
	} else {
		p.alias = strings.TrimSpace(args[0])
	}

	if p.alias == "" {
		fmt.Println("[spirit] please input alias name first")
		return
	}

	home := GetComponentHome(p.cliApp.Name)

	if !IsFileOrDir(home, true) {
		return
	}

	lockfilePath := p.getLockeFileName()

	if _, err := os.Stat(lockfilePath); os.IsNotExist(err) {
		fmt.Printf("[spirit] component of spirit %s not created before\n", p.alias)
		return
	}

	if lockfile, e := OpenLockFile(lockfilePath, 0640); e != nil {
		fmt.Printf("[spirit] open spirit of %s context failed, error: %s\n", p.alias, e)
		return
	} else if content, e := lockfile.ReadContent(); e != nil {
		fmt.Printf("[spirit] read spirit of %s context failed, path: %s, error: %s\n", p.alias, lockfilePath, e)
		return
	} else {
		if !IsProcessAlive(content.PID) {
			fmt.Printf("[spirit] spirit of %s already exited\n", p.alias)
			return
		}

		if e := KillProcess(content.PID); e != nil {
			fmt.Printf("[spirit] stop spirit of %s failed, pid: %d, error: %s\n", p.alias, content.PID, e)
			return
		}
		fmt.Printf("[spirit] stop spirit of %s success, pid: %d\n", p.alias, content.PID)
		return
	}
}

func (p *ClassicSpirit) restartProcess(c *cli.Context) {
	//TODO improve process logic, it was dirty implament currently
	if args := c.Args(); args == nil || len(args) == 0 {
		fmt.Println("[spirit] please input alias name first")
		return
	} else {
		p.alias = strings.TrimSpace(args[0])
	}

	if p.alias == "" {
		fmt.Println("[spirit] please input alias name first")
		return
	}

	useSTD := c.Bool("std")
	extEnvs := c.StringSlice("env")

	for _, env := range extEnvs {
		if v := strings.Split(env, "="); len(v) != 2 {
			fmt.Println("[spirit] env params format error, e.g.: ENV_KEY='value'")
			return
		}
	}

	home := GetComponentHome(p.cliApp.Name)

	if !IsFileOrDir(home, true) {
		return
	}

	lockfilePath := p.getLockeFileName()

	if _, err := os.Stat(lockfilePath); os.IsNotExist(err) {
		fmt.Printf("[spirit] component of spirit %s not created before\n", p.alias)
		return
	}

	if lockfile, e := OpenLockFile(lockfilePath, 0640); e != nil {
		fmt.Printf("[spirit] open spirit of %s context failed, error: %s\n", p.alias, e)
		return
	} else if content, e := lockfile.ReadContent(); e != nil {
		fmt.Printf("[spirit] read spirit of %s context failed, path: %s, error: %s\n", p.alias, lockfilePath, e)
		return
	} else {
		if IsProcessAlive(content.PID) {
			if e := KillProcess(content.PID); e != nil {
				fmt.Printf("[spirit] stop spirit of %s failed, pid: %d, error: %s\n", p.alias, content.PID, e)
				return
			}
		}

		newProcessArgs := []string{}
		if argsI, exist := content.Context["args"]; !exist {
			fmt.Printf("[spirit] spirit of %s's context damage please use command of component run to recreate instance\n", p.alias)
			return
		} else if args, ok := argsI.([]interface{}); ok {
			for i := 0; i < len(args); i++ {
				if strArg, ok := args[i].(string); ok {
					newProcessArgs = append(newProcessArgs, strArg)
				} else {
					fmt.Printf("[spirit] spirit of %s's context damage please use command of component run to recreate instance\n", p.alias)
					return
				}
			}
		}

		originEnvs := []string{}
		if envI, exist := content.Context["envs"]; exist && envI != nil {
			if envs, ok := envI.([]interface{}); ok {
				for i := 0; i < len(envs); i++ {
					if env, ok := envs[i].(string); ok {
						originEnvs = append(originEnvs, env)
					} else {
						fmt.Printf("[spirit] spirit of %s's context damage please use command of component run to recreate instance\n", p.alias)
						return
					}
				}
			} else {
				fmt.Printf("[spirit] spirit of %s's context damage please use command of component run to recreate instance\n", p.alias)
				return
			}
		}

		absPath := ""
		if path, e := filepath.Abs(os.Args[0]); e != nil {
			fmt.Printf("[spirit] restart spirit of %s failed, error: %s\n", p.alias, e)
			return
		} else {
			absPath = path
		}

		if pid, e := StartProcess(absPath, newProcessArgs, originEnvs, useSTD, extEnvs...); e != nil {
			fmt.Printf("[spirit] restart spirit of %s failed, error: %s\n", p.alias, e)
			return
		} else {
			fmt.Printf("[spirit] restart spirit of %s success, pid: %d\n", p.alias, pid)
			return
		}

		fmt.Printf("[spirit] restart spirit of %s success, pid: %d\n", p.alias, content.PID)
		return
	}
}

func (p *ClassicSpirit) getPID() (pid int) {
	if p.lockfile == nil {
		return
	}

	if lockfile, err := OpenLockFile(p.getLockeFileName(), 0640); err != nil {
		return
	} else if content, e := lockfile.ReadContent(); e != nil {
		return
	} else {
		pid = content.PID
	}

	return
}

func (p *ClassicSpirit) getLockeFileName() string {
	home := GetComponentHome(p.cliApp.Name)
	return home + "/" + p.alias + EXT_SPIRIT
}

func (p *ClassicSpirit) lock() (err error) {
	if p.alias == "" ||
		p.runningComponent == nil ||
		p.runningComponent.Name() == "" {
		return
	}

	if _, err = MakeComponentHome(p.cliApp.Name); err != nil {
		fmt.Printf("[spirit] make componet home dir failed, error: %s\n", err)
		return
	}

	if p.lockfile, err = CreateLockFile(p.getLockeFileName(), 0640); err != nil {
		return
	}

	args := []string{}

	for i := 1; i < len(os.Args); i++ {
		arg := strings.Trim(os.Args[i], "-")
		if arg != "co" &&
			arg != "creat-only" &&
			arg != "e" &&
			arg != "env" {
			args = append(args, os.Args[i])
		}

		if arg == "e" || arg == "env" {
			i++
		}
	}

	context := map[string]interface{}{"args": args, "envs": p.envs}
	if err = p.lockfile.WriteContent(context); err != nil {
		fmt.Println("[spirit] lock componet failed, error:", err)
		return
	}

	return
}
