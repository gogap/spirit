package spirit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/gogap/cli"
)

type ClassicSpirit struct {
	cliApp *cli.App

	receiverFactory MessageReceiverFactory
	senderFactory   MessageSenderFactory

	components map[string]Component
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
	componentName := c.String("name")
	receiverAddrs := c.StringSlice("address")

	if receiverAddrs == nil {
		fmt.Print("receiver address list is nil")
		return
	}

	componens := map[string]Component{}
	for _, receiverAddr := range receiverAddrs {
		receiverType := ""
		receiverConfig := ""
		receiverUrl := ""
		portName := ""
		handlerName := ""

		if componentName == "" {
			fmt.Println("component name is empty.")
			return
		}

		if receiverAddr == "" {
			fmt.Println("address is empty.")
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
			fmt.Println("address format error. example: port.in|delete|mqs|http://xxxx.com/queue?param=1|/etc/a.conf")
			return
		}

		if portName == "" {
			fmt.Println("receiver port name is empty.")
			return
		}

		if handlerName == "" {
			fmt.Println("handler name is empty.")
			return
		}

		if receiverType == "" {
			fmt.Println("receiver type is empty.")
			return
		}

		if receiverUrl == "" {
			fmt.Println("receiver url is empty.")
			return
		}

		var component Component
		if comp, exist := p.components[componentName]; !exist {
			fmt.Printf("component %s dose not hosting.", componentName)
			return
		} else {
			component = comp
		}

		if !p.receiverFactory.IsExist(receiverType) {
			fmt.Printf("the receiver type of %s dose not registered.", receiverType)
			return
		}

		if receiver, e := p.receiverFactory.NewReceiver(receiverType, receiverUrl, receiverConfig); e != nil {
			fmt.Println(e)
			return
		} else {
			component.BindHandler(portName, handlerName).
				BindReceiver(portName, receiver).
				SetMessageSenderFactory(p.senderFactory)
			componens[component.Name()] = component
		}
	}

	for _, component := range componens {
		component.Build().Run()
	}

	for {
		time.Sleep(time.Second)
	}
}

func (p *ClassicSpirit) cmdListComponent(c *cli.Context) {
	for _, component := range p.components {
		fmt.Println(component.Name())
	}
}

func (p *ClassicSpirit) cmdCallHandler(c *cli.Context) {
	componentName := c.String("name")
	handlerName := c.String("handler")
	payloadFile := c.String("payload")
	toJson := c.Bool("json")

	if componentName == "" {
		fmt.Println("component name is empty.")
		return
	}

	if handlerName == "" {
		fmt.Println("handler name is empty.")
		return
	}

	var component Component
	if comp, exist := p.components[componentName]; !exist {
		fmt.Printf("component %s dose not hosting.", componentName)
		return
	} else {
		component = comp
	}

	var bPayload []byte
	if payloadFile != "" {
		if data, e := ioutil.ReadFile(payloadFile); e != nil {
			fmt.Println("reade payload file error.", e)
			return
		} else {
			bPayload = data
		}
	}

	var payload *Payload

	if payloadFile != "" {
		payload = new(Payload)
		if e := payload.UnSerialize(bPayload); e != nil {
			fmt.Println("parse payload file failed, please make sure it is json format", e)
			return
		}
	}

	if result, e := component.CallHandler(handlerName, payload); e != nil {
		fmt.Println(e)
	} else {
		if toJson {
			if result != nil {
				if b, e := json.MarshalIndent(result, "", " "); e != nil {
					fmt.Println("format result to json failed.", e)
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

func (p *ClassicSpirit) Run() {
	p.cliApp.Run(os.Args)
}
