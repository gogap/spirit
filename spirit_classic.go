package spirit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/gogap/cli"
)

type ClassicSpirit struct {
	cliApp *cli.App

	receiverFactory MessageReceiverFactory
	senderFactory   MessageSenderFactory

	components       map[string]Component
	runningComponent Component

	heartbeaters       map[string]Heartbeater
	heartbeatSleepTime time.Duration

	isBuilt      bool
	isRunCommand bool
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

	receiverFactory := NewDefaultMessageReceiverFactory()
	receiverFactory.RegisterMessageReceivers(new(MessageReceiverMQS))

	senderFactory := NewDefaultMessageSenderFactory()
	senderFactory.RegisterMessageSenders(new(MessageSenderMQS))

	newSpirit.receiverFactory = receiverFactory
	newSpirit.senderFactory = senderFactory

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
						}, cli.IntFlag{
							Name:  "heartbeat, b",
							Value: 0,
							Usage: "the heartbeat sleep time, default = 0 (Millisecond), disabled",
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
	}
}

func (p *ClassicSpirit) cmdRunComponent(c *cli.Context) {
	p.isRunCommand = true
	componentName := c.String("name")
	receiverAddrs := c.StringSlice("address")
	heartbeatSleepTime := c.Int("heartbeat")

	p.heartbeatSleepTime = time.Millisecond * time.Duration(heartbeatSleepTime)

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

func (p *ClassicSpirit) Run() {
	if p.isRunCommand {
		if !p.isBuilt {
			fmt.Println("[spirit] spirit should build first")
			return
		}

		p.runningComponent.Run()

		fmt.Printf("[spirit] component %s running\n", p.runningComponent.Name())

		if p.heartbeatSleepTime > 0 {
			heartbeatMessage := p.getHeartbeatMessage()

			for _, heartbeater := range p.heartbeaters {
				if e := heartbeater.Start(); e != nil {
					panic(e)
				}
				fmt.Printf("[spirit] heartbeater %s running\n", heartbeater.Name())
				go func(beater Heartbeater, msg HeartbeatMessage, sleepTime time.Duration) {
					for {
						msg.CurrentTime = time.Now()
						msg.HeartbeatCount += 1
						beater.Heartbeat(msg)
						time.Sleep(sleepTime)
					}
				}(heartbeater, heartbeatMessage, p.heartbeatSleepTime)
			}
		}

		for {
			time.Sleep(time.Second)
		}
	}
}
