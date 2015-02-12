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

	newSpirit.receiverFactory = receiverFactory

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
						}, cli.StringFlag{
							Name:  "address, a",
							Value: "",
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
							Name:  "port, p",
							Value: "",
							Usage: "the name of component port's handler will be call",
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
	receiverAddr := c.String("address")
	receiverType := ""
	receiverConfig := ""
	receiverUrl := ""
	portName := ""

	if componentName == "" {
		fmt.Println("component name is empty.")
		return
	}

	if receiverAddr == "" {
		fmt.Println("address is empty.")
		return
	}

	addr := strings.Split(receiverAddr, "|")

	if len(addr) == 3 {
		portName = addr[0]
		receiverType = addr[1]
		receiverUrl = addr[2]
	} else if len(addr) == 4 {
		portName = addr[0]
		receiverType = addr[1]
		receiverUrl = addr[2]
		receiverConfig = addr[3]
	} else {
		fmt.Println("address format error. example: port.in|mqs|http://xxxx.com/queue?param=1|/etc/a.conf")
		return
	}

	if portName == "" {
		fmt.Println("receiver port name is empty")
		return
	}

	if receiverType == "" {
		fmt.Println("receiver type is empty")
		return
	}

	if receiverUrl == "" {
		fmt.Println("receiver url is empty")
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
		component.BindReceiver(portName, receiver)
		component.Build()
		component.Run()

		for {
			time.Sleep(time.Second)
		}
	}
}

func (p *ClassicSpirit) cmdListComponent(c *cli.Context) {
	for _, component := range p.components {
		fmt.Println(component.Name())
	}
}

func (p *ClassicSpirit) cmdCallHandler(c *cli.Context) {
	componentName := c.String("name")
	portName := c.String("port")
	payloadFile := c.String("payload")
	toJson := c.Bool("json")

	if componentName == "" {
		fmt.Println("component name is empty.")
		return
	}

	if portName == "" {
		fmt.Println("port name is empty.")
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
		if e := json.Unmarshal(bPayload, payload); e != nil {
			fmt.Println("parse payload file failed, please make sure it is json format", e)
			return
		}
	}

	if result, e := component.CallHandler(portName, payload); e != nil {
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
