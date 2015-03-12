package spirit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/codegangsta/cli"
)

const (
	EXT_SPIRIT = ".spirit"
)

type ClassicSpirit struct {
	cliApp *cli.App

	receiverFactory MessageReceiverFactory
	senderFactory   MessageSenderFactory

	components       map[string]Component
	runningComponent Component

	heartbeaters       map[string]Heartbeater
	heartbeatSleepTime time.Duration
	heartbeatersToRun  map[string]bool
	heartbeaterConfig  string
	alias              string

	isBuilt                       bool
	isRunCommand                  bool
	isBuildCheckOnly              bool
	isCreatAliasdSpirtContextOnly bool

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
			ShortName: "c",
			Usage:     "options for component",
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
		},
		{
			Name:      "ps",
			ShortName: "",
			Usage:     "show running process which spirit have alias",
			Action:    p.showProcess,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "all, a",
					Usage: "show process which have alias, running and exited",
				},
			},
		}, {
			Name:      "start",
			ShortName: "",
			Usage:     "start the process create by command of component run and have alias",
			Action:    p.startProcess,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "alias, a",
					Usage: "component of spirit with alias have been ran",
				},
			},
		}, {
			Name:      "stop",
			ShortName: "",
			Usage:     "stop the process create by command of component run and have alias",
			Action:    p.stopProcess,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "alias, a",
					Usage: "running component of spirit with alias",
				},
			},
		}, {
			Name:      "restart",
			ShortName: "",
			Usage:     "restart the process create by command of component run and have alias",
			Action:    p.restartProcess,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "alias, a",
					Usage: "restart the running component of spirit with alias",
				},
			},
		},
	}
}

func (p *ClassicSpirit) cmdRunComponent(c *cli.Context) {
	p.isRunCommand = true
	componentName := c.String("name")
	receiverAddrs := c.StringSlice("address")

	heartbeatersToRun := c.StringSlice("heartbeat")
	heartbeatSleepTime := c.Int("heartbeat-sleeptime")
	heartbeaterConfig := c.String("heartbeat-config")

	p.alias = c.String("alias")
	p.isBuildCheckOnly = c.Bool("build-check")
	p.isCreatAliasdSpirtContextOnly = c.Bool("create-only")

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

	if receiverAddrs == nil {
		fmt.Println("[spirit] receiver address list is nil")
		return
	}

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
	if p.isRunCommand {
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
				if e := initFunc(); e != nil {
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
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	for {
		select {
		case killSignal := <-interrupt:
			if killSignal == os.Interrupt {
				fmt.Printf("[spirit] component %s was interruped by system signal\n", p.runningComponent.Name())
				return
			}
			fmt.Printf("[spirit] component %s was killed\n", p.runningComponent.Name())
			return
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
			if filepath.Ext(name) == EXT_SPIRIT && name != EXT_SPIRIT {
				if lockfile, e := OpenLockFile(p.getLockeFileName(), 0640); e != nil {
					fmt.Println("[spirit] open spirit context file failed, error:", e)
				} else if content, e := lockfile.ReadContent(); e != nil {
					fmt.Println("[spirit] error context: ", p.getLockeFileName())
				} else {
					if IsProcessAlive(content.PID) || showAll {
						contents = append(contents, content)
					}
				}
			}
		}
	}

	if data, e := json.MarshalIndent(contents, " ", "  "); e != nil {
		fmt.Printf("[spirit] format contents to json failed, error:", e)
		return
	} else {
		fmt.Println(string(data))
	}
}

func (p *ClassicSpirit) startProcess(c *cli.Context) {
	//TODO improve process logic, it was dirty implament currently
	p.alias = c.String("alias")
	p.alias = strings.TrimSpace(p.alias)

	if p.alias == "" {
		fmt.Println("[spirit] please input alias name first")
		return
	}

	home := GetComponentHome(p.cliApp.Name)

	if !IsFileOrDir(home, true) {
		return
	}

	lockfilePath := p.getLockeFileName()
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

		absPath := ""
		if path, e := filepath.Abs(os.Args[0]); e != nil {
			fmt.Printf("[spirit] start spirit of %s failed, error: %s\n", p.alias, e)
			return
		} else {
			absPath = path
		}

		if pid, e := StartProcess(absPath, newProcessArgs); e != nil {
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
	p.alias = c.String("alias")
	p.alias = strings.TrimSpace(p.alias)

	if p.alias == "" {
		fmt.Println("[spirit] please input alias name first")
		return
	}

	home := GetComponentHome(p.cliApp.Name)

	if !IsFileOrDir(home, true) {
		return
	}

	lockfilePath := p.getLockeFileName()
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
	p.alias = c.String("alias")
	p.alias = strings.TrimSpace(p.alias)

	if p.alias == "" {
		fmt.Println("[spirit] please input alias name first")
		return
	}

	home := GetComponentHome(p.cliApp.Name)

	if !IsFileOrDir(home, true) {
		return
	}

	lockfilePath := p.getLockeFileName()
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

		absPath := ""
		if path, e := filepath.Abs(os.Args[0]); e != nil {
			fmt.Printf("[spirit] restart spirit of %s failed, error: %s\n", p.alias, e)
			return
		} else {
			absPath = path
		}

		if pid, e := StartProcess(absPath, newProcessArgs); e != nil {
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
		fmt.Printf("[spirit] make componet home dir failed, error:", err)
		return
	}

	if p.lockfile, err = CreateLockFile(p.getLockeFileName(), 0640); err != nil {
		return
	}

	args := []string{}

	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] != "--co" &&
			os.Args[i] != "--creat-only" &&
			os.Args[i] != "-creat-only" &&
			os.Args[i] != "-co" {
			args = append(args, os.Args[i])
		}
	}

	context := map[string]interface{}{"args": args}
	if err = p.lockfile.WriteContent(context); err != nil {
		fmt.Println("[spirit] lock componet failed, error:", err)
		return
	}

	return
}
