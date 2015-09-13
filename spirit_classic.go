package spirit

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/codegangsta/cli"
	"github.com/gogap/errors"
	"github.com/gogap/event_center"
	"github.com/gogap/logs"
)

const (
	EXT_SPIRIT = ".spirit"

	ENV_NAME = "SPIRIT_ENV"
	ENV_EXT  = ".env"
)

type hostingComponent struct {
	Component   Component
	InitalFuncs []InitalFunc
}

type Author struct {
	Name  string
	Email string
}

type ClassicSpirit struct {
	name         string
	instanceName string

	cliApp *cli.App

	hostingComponents map[string]hostingComponent
	runningComponents map[string]Component

	runComponentsConf []runComponentConf

	metadata instanceMetadata

	isBuilt             bool
	isRunCommand        bool
	isRunCheckedCorrect bool
	viewDetails         bool
	inspectMode         bool
}

func NewClassicSpirit(name, description, version string, authors []Author) Spirit {
	newSpirit := new(ClassicSpirit)

	app := cli.NewApp()
	app.Usage = description
	app.Authors = nil
	app.Version = version
	app.HideVersion = true
	app.Name = name

	if authors != nil {
		for _, author := range authors {
			app.Authors = append(app.Authors, cli.Author{Name: author.Name, Email: author.Email})
		}
	}

	app.Commands = []cli.Command{
		commandRun(newSpirit.cmdRun),
		commandCreate(newSpirit.cmdCreate),
		commandComponents(newSpirit.cmdListComponents),
		commandPS(newSpirit.cmdPS),
		commandStart(newSpirit.cmdStartInstance),
		commandStop(newSpirit.cmdStopInstance),
		commandKill(newSpirit.cmdKillInstance),
		commandPause(newSpirit.cmdPauseInstance),
		commandRemoveInstance(newSpirit.cmdRemoveInstance),
		commandInspect(newSpirit.cmdInspectInstance),
		commandInstanceBins(newSpirit.cmdInstanceBins),
		commandVersion(newSpirit.cmdVersion),
		commandCommit(newSpirit.cmdCommit),
		commandRemoveBinary(newSpirit.cmdRemoveBinary),
		commandCheckout(newSpirit.cmdCheckout),
	}

	newSpirit.cliApp = app
	newSpirit.name = name
	newSpirit.hostingComponents = make(map[string]hostingComponent)
	newSpirit.runningComponents = make(map[string]Component)

	instManager = newInstanceManager(name)

	return newSpirit
}

func (p *ClassicSpirit) Hosting(component Component, initalFuncs ...InitalFunc) Spirit {
	if component == nil {
		panic("component is nil")
	}

	if component.Name() == "" {
		panic("component name is empty")
	}

	if _, exist := p.hostingComponents[component.Name()]; exist {
		panic("component of " + component.Name() + " already hosting")
	}

	p.hostingComponents[component.Name()] = hostingComponent{Component: component, InitalFuncs: initalFuncs}

	return p
}

func (p *ClassicSpirit) Build() Spirit {
	p.cliApp.Run(os.Args)
	p.isBuilt = true
	return p
}

func (p *ClassicSpirit) Run() {
	if !p.isRunCommand {
		return
	}

	var err error
	defer func() {
		if err != nil {
			exitError(err)
		}
	}()

	if !p.isBuilt {
		err = ERR_SPIRIT_SPIRIT_DID_NOT_BUILD.New()
		return
	}

	if err = p.runComponents(); err != nil {
		return
	}

	if err = p.startHeartbeat(); err != nil {
		return
	}

	p.waitSignal()
}

func (p *ClassicSpirit) cmdStartInstance(c *cli.Context) {
	var err error

	defer func() {
		if err != nil {
			exitError(err)
		}
	}()

	viewDetails = c.Bool("v")

	args := c.Args()

	if len(args) == 0 {
		err = ERR_SPIRIT_INSTANCE_NAME_NOT_INPUT.New()
		return
	} else if len(args) != 1 {
		err = ERR_SPIRIT_ONLY_ONE_INST_NAME.New()
		return
	}

	instanceName := args[0]

	if instanceName == "" {
		err = ERR_SPIRIT_INSTANCE_NAME_NOT_INPUT.New()
		return
	}

	p.instanceName = instanceName

	var metadata *instanceMetadata
	if metadata, err = instManager.GetMetadata(instanceName); err != nil {
		return
	}

	p.metadata = *metadata

	if isChildInstance() || isPersistentProcess(instanceName) {
		if err = p.startInstance(); err != nil {
			return
		}
	} else {

		attach := c.Bool("attach")
		stdin := c.Bool("interactive")
		hash := c.String("hash")

		var exeCMD *exec.Cmd
		if exeCMD, err = p.startChildInstance(hash, attach, stdin); err != nil {
			return
		}

		if attach {
			exeCMD.Wait()
		}
	}
	return
}

func (p *ClassicSpirit) startInstance() (err error) {

	if !instManager.IsInstanceExist(p.instanceName) {
		err = ERR_SPIRIT_INSTANCE_NOT_EXIST.New(errors.Params{"name": p.instanceName})
		return
	}

	Assets = newAssetsManager(p.instanceName)

	p.initalEnvs()

	if err = Assets.loadAssets(); err != nil {
		return
	}

	if p.instanceName != "" {
		if err = instManager.TryLockPID(p.instanceName); err != nil {
			return
		}
	}

	if err = p.buildRunComponents(); err != nil {
		return
	}

	p.isRunCommand = true
	p.isRunCheckedCorrect = true

	p.metadata.LastStartTime = time.Now()
	p.metadata.UpdateTime = time.Now()
	p.metadata.RunTimes += 1

	if err = p.saveMetadata(); err != nil {
		return
	}

	return
}

func (p *ClassicSpirit) startChildInstance(hash string, attach, stdin bool) (exeCMD *exec.Cmd, err error) {

	cwd := "./"
	if isPersistentProcess(p.instanceName) {
		cwd = "../../../"
	}

	if cwd, err = filepath.Abs(cwd); err != nil {
		return
	}

	binHome := filepath.Dir(os.Args[0])

	instManager = newInstanceManager(p.name)

	execHash := p.metadata.CurrentHash
	if hash != "" {
		execHash = hash
	}

	binPath := instManager.GetInstanceBinPath(p.instanceName, execHash)

	cmd := []string{"start", p.instanceName}
	if viewDetails {
		cmd = append(cmd, "-v")
	}

	extEnvs := os.Environ()

	extEnvs = append(extEnvs, SPIRIT_CHILD_INSTANCE+"=1", SPIRIT_BIN_HOME+"="+binHome)

	if exeCMD, err = startProcess(binPath, cmd, p.metadata.Envs, stdin, attach, cwd, extEnvs...); err != nil {
		err = ERR_SPIRIT_START_CHILD_INSTANCE.New(errors.Params{"name": p.instanceName, "err": err})
		return
	}

	if viewDetails {
		fmt.Printf("new spirit instance of %s started, pid: %d\n", p.instanceName, exeCMD.Process.Pid)
	}

	return
}

func (p *ClassicSpirit) cmdStopInstance(c *cli.Context) {
	var err error

	defer func() {
		if err != nil {
			exitError(err)
		}
	}()

	args := c.Args()

	if len(args) == 0 {
		err = ERR_SPIRIT_INSTANCE_NAME_NOT_INPUT.New()
		return
	} else if len(args) != 1 {
		err = ERR_SPIRIT_ONLY_ONE_INST_NAME.New()
		return
	}

	instanceName := args[0]

	if instanceName == "" {
		err = ERR_INSTANCE_NAME_IS_EMPTY.New()
		return
	}

	if !instManager.IsInstanceExist(instanceName) {
		err = ERR_SPIRIT_INSTANCE_NOT_EXIST.New(errors.Params{"name": instanceName})
		return
	}

	pid := 0
	if pid, err = instManager.ReadPID(instanceName); err != nil {
		return
	}

	if pid <= 0 {
		err = ERR_SPIRIT_INSTANCE_NOT_RUNNING.New(errors.Params{"name": instanceName})
		return
	}

	if e := stopProcess(pid); e != nil {
		err = ERR_SPIRIT_STOP_INSTANCE_FIALED.New(errors.Params{"name": instanceName, "pid": pid, "err": e})
		return
	}

	for isProcessAlive(pid) {
		time.Sleep(time.Millisecond * 100)
	}

	return
}

func (p *ClassicSpirit) cmdKillInstance(c *cli.Context) {
	var err error

	defer func() {
		if err != nil {
			exitError(err)
		}
	}()

	args := c.Args()

	if len(args) == 0 {
		err = ERR_SPIRIT_INSTANCE_NAME_NOT_INPUT.New()
		return
	} else if len(args) != 1 {
		err = ERR_SPIRIT_ONLY_ONE_INST_NAME.New()
		return
	}

	instanceName := args[0]

	if instanceName == "" {
		err = ERR_INSTANCE_NAME_IS_EMPTY.New()
		return
	}

	if !instManager.IsInstanceExist(instanceName) {
		err = ERR_SPIRIT_INSTANCE_NOT_EXIST.New(errors.Params{"name": instanceName})
		return
	}

	pid := 0
	if pid, err = instManager.ReadPID(instanceName); err != nil {
		return
	}

	if pid <= 0 {
		err = ERR_SPIRIT_INSTANCE_NOT_RUNNING.New(errors.Params{"name": instanceName})
		return
	}

	if e := killProcess(pid); e != nil {
		err = ERR_SPIRIT_KILL_INSTANCE_FIALED.New(errors.Params{"name": instanceName, "pid": pid, "err": e})
		return
	}
}

func (p *ClassicSpirit) cmdPauseInstance(c *cli.Context) {
	var err error

	defer func() {
		if err != nil {
			exitError(err)
		}
	}()

	args := c.Args()

	if len(args) == 0 {
		err = ERR_SPIRIT_INSTANCE_NAME_NOT_INPUT.New()
		return
	} else if len(args) != 1 {
		err = ERR_SPIRIT_ONLY_ONE_INST_NAME.New()
		return
	}

	instanceName := args[0]

	if instanceName == "" {
		err = ERR_INSTANCE_NAME_IS_EMPTY.New()
		return
	}

	if !instManager.IsInstanceExist(instanceName) {
		err = ERR_SPIRIT_INSTANCE_NOT_EXIST.New(errors.Params{"name": instanceName})
		return
	}

	pid := 0
	if pid, err = instManager.ReadPID(instanceName); err != nil {
		return
	}

	if pid <= 0 {
		err = ERR_SPIRIT_INSTANCE_NOT_RUNNING.New(errors.Params{"name": instanceName})
		return
	}

	if e := pauseProcess(pid); e != nil {
		err = ERR_SPIRIT_PAUSE_INSTANCE_FIALED.New(errors.Params{"name": instanceName, "pid": pid, "err": e})
		return
	}
}

func (p *ClassicSpirit) cmdRemoveBinary(c *cli.Context) {
	var err error

	defer func() {
		if err != nil {
			exitError(err)
		}
	}()

	args := c.Args()

	if len(args) == 0 {
		err = ERR_SPIRIT_INSTANCE_NAME_NOT_INPUT.New()
		return
	} else if len(args) != 1 {
		err = ERR_SPIRIT_ONLY_ONE_INST_NAME.New()
		return
	}

	instanceName := args[0]

	if instanceName == "" {
		err = ERR_INSTANCE_NAME_IS_EMPTY.New()
		return
	}

	hash := c.String("hash")

	if hash == "" {
		err = ERR_SPIRIT_INSTANCE_HASH_NOT_INPUT.New()
		return
	}

	if !instManager.IsInstanceExist(instanceName) {
		err = ERR_SPIRIT_INSTANCE_NOT_EXIST.New(errors.Params{"name": instanceName})
		return
	}

	if err = instManager.TryLockPID(instanceName); err != nil {
		return
	}

	if err = instManager.TryLockPID(instanceName); err != nil {
		return
	}

	err = instManager.RemoveBinary(instanceName, hash)
}

func (p *ClassicSpirit) cmdRemoveInstance(c *cli.Context) {
	var err error

	defer func() {
		if err != nil {
			exitError(err)
		}
	}()

	args := c.Args()

	if len(args) == 0 {
		err = ERR_SPIRIT_INSTANCE_NAME_NOT_INPUT.New()
		return
	} else if len(args) != 1 {
		err = ERR_SPIRIT_ONLY_ONE_INST_NAME.New()
		return
	}

	instanceName := args[0]

	if err = instManager.TryLockPID(instanceName); err != nil {
		return
	}

	err = instManager.RemoveInstance(instanceName)
}

func (p *ClassicSpirit) cmdInspectInstance(c *cli.Context) {
	var err error

	defer func() {
		if err != nil {
			exitError(err)
		}
	}()

	args := c.Args()

	if len(args) == 0 {
		err = ERR_SPIRIT_INSTANCE_NAME_NOT_INPUT.New()
		return
	} else if len(args) != 1 {
		err = ERR_SPIRIT_ONLY_ONE_INST_NAME.New()
		return
	}

	instanceName := args[0]

	if instanceName == "" {
		err = ERR_INSTANCE_NAME_IS_EMPTY.New()
		return
	}

	if !instManager.IsInstanceExist(instanceName) {
		err = ERR_SPIRIT_INSTANCE_NOT_EXIST.New(errors.Params{"name": instanceName})
		return
	}

	printStr := ""

	if c.Bool("config") {
		confManager := newAssetsManager(instanceName)
		if err = confManager.loadAssets(); err != nil {
			return
		}

		confStr := ""
		if confStr, err = confManager.serializeAssets(); err != nil {
			return
		}
		printStr += "* config:\n"
		printStr += confStr

	}

	if c.Bool("metadata") {
		var metadata *instanceMetadata
		if metadata, err = instManager.GetMetadata(instanceName); err != nil {
			return
		}
		metaStr := ""
		if metaStr, err = metadata.Serialize(); err != nil {
			return
		}
		printStr += "\n\n* metadata:\n"
		printStr += metaStr
	}

	fmt.Println(printStr)
}

func (p *ClassicSpirit) cmdInstanceBins(c *cli.Context) {
	var err error

	defer func() {
		if err != nil {
			exitError(err)
		}
	}()

	args := c.Args()

	if len(args) == 0 {
		err = ERR_SPIRIT_INSTANCE_NAME_NOT_INPUT.New()
		return
	} else if len(args) != 1 {
		err = ERR_SPIRIT_ONLY_ONE_INST_NAME.New()
		return
	}

	instanceName := args[0]

	if instanceName == "" {
		err = ERR_INSTANCE_NAME_IS_EMPTY.New()
		return
	}

	if !instManager.IsInstanceExist(instanceName) {
		err = ERR_SPIRIT_INSTANCE_NOT_EXIST.New(errors.Params{"name": instanceName})
		return
	}

	var instanceBins []instanceBin
	if instanceBins, err = instManager.ListInstanceBins(instanceName); err != nil {
		return
	}

	var metadata *instanceMetadata
	if metadata, err = instManager.GetMetadata(instanceName); err != nil {
		return
	}

	strPrint := fmt.Sprintf("ID\tHASH\t%43s\t%22s\n", "CREATE TIME", "COMMIT MESSAGE")
	strPrintTpl := "%s%d\t%32s\t%s\t%s\n"
	for id, bin := range instanceBins {
		star := " "
		if metadata.CurrentHash == bin.hash {
			star = "*"
		}
		strPrint += fmt.Sprintf(strPrintTpl, star, id+1, bin.hash, bin.createTime, bin.message)
	}
	fmt.Print(strPrint)
}

func (p *ClassicSpirit) cmdVersion(c *cli.Context) {
	fmt.Println(getProcBinHash())
}

func (p *ClassicSpirit) cmdCommit(c *cli.Context) {
	var err error

	defer func() {
		if err != nil {
			exitError(err)
		}
	}()

	commitMessage := c.String("message")

	if commitMessage == "" {
		err = ERR_SPIRIT_COMMIT_MSG_NOT_INPUT.New()
		return
	}

	args := c.Args()

	if len(args) == 0 {
		err = ERR_SPIRIT_INSTANCE_NAME_NOT_INPUT.New()
		return
	} else if len(args) != 1 {
		err = ERR_SPIRIT_ONLY_ONE_INST_NAME.New()
		return
	}

	instanceName := args[0]

	if instanceName == "" {
		err = ERR_INSTANCE_NAME_IS_EMPTY.New()
		return
	}

	if !instManager.IsInstanceExist(instanceName) {
		err = ERR_SPIRIT_INSTANCE_NOT_EXIST.New(errors.Params{"name": instanceName})
		return
	}

	err = instManager.CommitBinary(instanceName, commitMessage)

	return
}

func (p *ClassicSpirit) cmdCheckout(c *cli.Context) {
	var err error

	defer func() {
		if err != nil {
			exitError(err)
		}
	}()

	args := c.Args()

	if len(args) == 0 {
		err = ERR_SPIRIT_INSTANCE_NAME_NOT_INPUT.New()
		return
	} else if len(args) != 1 {
		err = ERR_SPIRIT_ONLY_ONE_INST_NAME.New()
		return
	}

	hash := c.String("hash")

	if hash == "" {
		err = ERR_SPIRIT_INSTANCE_HASH_NOT_INPUT.New()
		return
	}

	instanceName := args[0]

	if instanceName == "" {
		err = ERR_INSTANCE_NAME_IS_EMPTY.New()
		return
	}

	if !instManager.IsInstanceExist(instanceName) {
		err = ERR_SPIRIT_INSTANCE_NOT_EXIST.New(errors.Params{"name": instanceName})
		return
	}

	if err = instManager.TryLockPID(instanceName); err != nil {
		return
	}

	message := ""
	if message, err = instManager.CheckoutBinary(instanceName, hash); err != nil {
		return
	}

	fmt.Println("Commit message:", message)
}

func (p *ClassicSpirit) cmdRun(c *cli.Context) {
	var err error

	defer func() {
		if err != nil {
			exitError(err)
		}
	}()

	p.isRunCommand = true
	p.metadata.Envs = c.StringSlice("env")
	p.instanceName = c.String("name")

	message := c.String("message")

	viewDetails = c.Bool("v")
	p.inspectMode = c.Bool("i")

	if p.instanceName != "" {
		if instManager.IsInstanceExist(p.instanceName) {
			err = ERR_SPIRIT_INSTANCE_ALREADY_CREATED.New(errors.Params{"name": p.instanceName})
			return
		}
	}

	if err = initalSpiritHome(p.name); err != nil {
		return
	}

	Assets = newAssetsManager(p.instanceName)

	var runComponents []string
	if runComponents, err = p.getRunComponents(c); err != nil {
		return
	}

	p.metadata.RunComponents = runComponents

	if err = p.initalEnvs(); err != nil {
		return
	}

	if err = p.initalConf(c); err != nil {
		return
	}

	if p.instanceName != "" {
		if err = instManager.TryLockPID(p.instanceName); err != nil {
			return
		}
	}

	if err = p.buildRunComponents(); err != nil {
		return
	}

	if err = p.buildEventInspectSubscribers(); err != nil {
		return
	}

	p.isRunCheckedCorrect = true

	if err = Assets.save(); err != nil {
		return
	}

	p.metadata.CreateTime = time.Now()
	p.metadata.LastStartTime = time.Now()
	p.metadata.UpdateTime = time.Now()
	p.metadata.RunTimes = 1

	if err = p.saveMetadata(); err != nil {
		return
	}

	if p.instanceName != "" {
		if err = instManager.CommitBinary(p.instanceName, message); err != nil {
			return
		}
	}
}

func (p *ClassicSpirit) cmdCreate(c *cli.Context) {
	var err error

	defer func() {
		if err != nil {
			exitError(err)
		}
	}()

	p.metadata.Envs = c.StringSlice("env")
	p.instanceName = c.String("name")

	message := c.String("message")

	if p.instanceName == "" {
		err = ERR_SPIRIT_INSTANCE_NAME_NOT_INPUT.New()
		return
	}

	if instManager.IsInstanceExist(p.instanceName) {
		err = ERR_SPIRIT_INSTANCE_ALREADY_CREATED.New(errors.Params{"name": p.instanceName})
		return
	}

	if err = initalSpiritHome(p.name); err != nil {
		return
	}

	Assets = newAssetsManager(p.instanceName)

	var runComponents []string
	if runComponents, err = p.getRunComponents(c); err != nil {
		return
	}

	p.metadata.RunComponents = runComponents

	if err = p.initalEnvs(); err != nil {
		return
	}

	if err = p.initalConf(c); err != nil {
		return
	}

	if err = p.buildRunComponents(); err != nil {
		return
	}

	if p.instanceName != "" {
		if err = instManager.TryLockPID(p.instanceName); err != nil {
			return
		}

		binHash := getProcBinHash()

		// binPath := instManager.GetInstanceBinPath(p.instanceName, binHash)

		// if _, e := copyFile(binPath, os.Args[0], 744); e != nil {
		// 	err = ERR_SPIRIT_COPY_INSTANCE_ERROR.New(errors.Params{"err": e})
		// 	return
		// }

		p.metadata.CreateTime = time.Now()
		p.metadata.LastStartTime = time.Now()
		p.metadata.UpdateTime = time.Now()
		p.metadata.RunTimes = 0
		p.metadata.CurrentHash = binHash

		if err = p.saveMetadata(); err != nil {
			return
		}

		if err = Assets.save(); err != nil {
			return
		}

		if err = instManager.CommitBinary(p.instanceName, message); err != nil {
			return
		}
	}
}

func (p *ClassicSpirit) buildEventInspectSubscribers() (err error) {
	if !p.inspectMode {
		return
	}

	events := EventCenter.ListEvents()

	loggerSubscriber := event_center.NewSubscriber(func(eventName string, values ...interface{}) {

		str := ""
		for i, v := range values {
			str = str + fmt.Sprintf("v_%d: %v\n", i, v)
		}

		logs.Info(fmt.Sprintf("SPIRIT-EVENT-INSPECT EVENT_NAME: %s\nValues:\n%s\n", eventName, str))

	})

	for _, event := range events {
		if err = EventCenter.Subscribe(event, loggerSubscriber); err != nil {
			return
		}
	}

	return
}

func (p *ClassicSpirit) saveMetadata() (err error) {
	if p.instanceName == "" {
		return
	}

	err = instManager.SaveMetadata(p.instanceName, p.metadata)

	return
}

func (p *ClassicSpirit) cmdListComponents(c *cli.Context) {
	viewDetails := c.Bool("v")

	for name, hostingComponent := range p.hostingComponents {
		fmt.Println(" *", name)
		if viewDetails {
			if handlers, e := hostingComponent.Component.ListHandlers(); e != nil {
				fmt.Println("[spirit] "+name+":", e.Error())
			} else {
				for name, _ := range handlers {
					fmt.Println("  -", name)
				}
			}
		}
	}
}

func (p *ClassicSpirit) cmdPS(c *cli.Context) {
	var err error

	defer func() {
		if err != nil {
			exitError(err)
		}
	}()

	showAll := c.Bool("all")
	nameOnly := c.Bool("quiet")

	var instances []instanceStatus
	if instances, err = instManager.ListInstanceStatus(showAll); err != nil {
		return
	}

	if nameOnly {
		for _, instance := range instances {
			fmt.Println(instance.name)
		}
	} else {
		strPrint := fmt.Sprintf("ID\tPID\tNAME%10s\tHASH%28s\tSTATUS\tCOMMIT MESSAGE\n", "", "")
		strPrintTpl := "%d\t%s\t%s\t%s\t%s\t%s\n"
		for id, instance := range instances {
			strPrint += fmt.Sprintf(strPrintTpl, id+1, instance.pid, instance.name, instance.hash, instance.status, instance.message)
		}
		fmt.Print(strPrint)
	}

	return
}

func (p *ClassicSpirit) initalEnvs() (err error) {
	if p.metadata.Envs != nil {
		for _, env := range p.metadata.Envs {
			kv := strings.Split(env, "=")
			if len(kv) == 1 {
				os.Setenv(kv[0], "")
			} else if len(kv) == 2 {
				os.Setenv(kv[0], kv[1])
			} else {
				err = ERR_SPIRIT_ENV_FORAMT_ERROR.New(errors.Params{"env": kv})
				return
			}
		}
	}

	return
}

func (p *ClassicSpirit) initalConf(c *cli.Context) (err error) {
	configFile := c.String("config")

	configFile = strings.TrimSpace(configFile)

	if configFile == "" {
		err = ERR_SPIRIT_NO_CONFIG_SPECIFIC.New()
		return
	}

	if err = Assets.load(configFile, true); err != nil {
		return
	}

	if err = Assets.Unmarshal(configFile, &p.metadata.RunConfig); err != nil {
		return
	}

	if p.metadata.RunConfig.Assets != nil {
		for _, assetProp := range p.metadata.RunConfig.Assets {
			if err = Assets.load(assetProp.FileName, assetProp.NeedBuild); err != nil {
				return
			}
		}
	}

	if p.viewDetails {
		confStr := ""
		if confStr, err = p.metadata.RunConfig.Serialize(); err != nil {
			return
		}
		fmt.Println("[SPIRIT-INFO] built config:\n", confStr)
	}

	return
}

func (p *ClassicSpirit) buildRunComponents() (err error) {
	componentsBuilt := map[string]bool{}

	for _, runComponent := range p.metadata.RunComponents {
		componentsBuilt[runComponent] = false
	}

	if p.metadata.RunConfig.Components == nil || len(p.metadata.RunConfig.Components) == 0 {
		err = ERR_SPIRIT_NO_COMPONENTS_TO_RUN.New()
		return
	}

	for _, componentConf := range p.metadata.RunConfig.Components {
		if built, exist := componentsBuilt[componentConf.Name]; !exist {
			continue
		} else if built {
			err = ERR_SPIRIT_COMPONENT_CONF_DUP.New(errors.Params{"name": componentConf.Name})
			return
		}

		var component Component

		if hostingComp, exist := p.hostingComponents[componentConf.Name]; !exist {
			err = ERR_SPIRIT_COMPONENT_NOT_HOSTING.New(errors.Params{"name": componentConf.Name})
			return
		} else {
			component = hostingComp.Component
		}

		ports := map[string]bool{}

		//build address
		for _, address := range componentConf.Address {
			if !receiverFactory.IsExist(address.Type) {
				err = ERR_RECEIVER_DRIVER_NOT_EXIST.New(errors.Params{"type": address.Type})
				return
			}

			var receiver MessageReceiver
			if receiver, err = receiverFactory.NewReceiver(address.Type, address.Url, address.Options); err != nil {
				return
			}

			ports[address.PortName] = true

			component.
				BindHandler(address.PortName, address.HandlerName).
				BindReceiver(address.PortName, receiver)
		}

		inPortHooks := map[string][]string{}

		//inject global port hooks
		globalHookshooks := []string{}
		for _, hookConf := range p.metadata.RunConfig.GlobalHooks {
			if _, err = hookFactory.CreateHook(hookConf.Type, hookConf.Name, hookConf.Options); err != nil {
				return
			}
			globalHookshooks = append(globalHookshooks, hookConf.Name)
		}

		for portName, _ := range ports {
			inPortHooks[portName] = globalHookshooks
		}

		//inject port hooks
		for _, hookConf := range componentConf.PortHooks {
			if _, err = hookFactory.CreateHook(hookConf.Type, hookConf.Name, hookConf.Options); err != nil {
				return
			}

			if hooks, exist := inPortHooks[hookConf.Port]; exist {
				inPortHooks[hookConf.Port] = append(hooks, hookConf.Name)
			} else {
				inPortHooks[hookConf.Port] = []string{hookConf.Name}
			}
		}

		for portName, hooks := range inPortHooks {
			component.AddInPortHooks(portName, hooks...)
		}

		p.runComponentsConf = append(p.runComponentsConf, componentConf)

		component.Build()

		componentsBuilt[component.Name()] = true
	}

	for componentName, isBuilt := range componentsBuilt {
		if !isBuilt {
			err = ERR_SPIRIT_COMPONENT_CONF_NOT_EXIST.New(errors.Params{"name": componentName})
			return
		}
	}
	return
}

func (p *ClassicSpirit) runComponents() (err error) {
	for _, componentConf := range p.runComponentsConf {
		hostingComponent := p.hostingComponents[componentConf.Name]
		if hostingComponent.InitalFuncs != nil {
			for _, fn := range hostingComponent.InitalFuncs {
				//run component inital func, just like a `main` for component
				if err = fn(); err != nil {
					return
				}
			}
		}

		p.runningComponents[componentConf.Name] = hostingComponent.Component

		go p.hostingComponents[componentConf.Name].Component.Run()
	}

	return
}

func (p *ClassicSpirit) startHeartbeat() (err error) {
	if p.metadata.RunConfig.Heartbeat == nil || len(p.metadata.RunConfig.Heartbeat) == 0 {
		return
	}

	heartbeatMessage := p.getHeartbeatMessage()

	for _, heartbeatConf := range p.metadata.RunConfig.Heartbeat {
		if heartbeater, exist := heartbeaters[heartbeatConf.Heart]; exist {
			if err = heartbeater.Start(heartbeatConf.Options); err != nil {
				return
			}

			go func(beater Heartbeater, msg HeartbeatMessage, sleepTime time.Duration) {
				isStopped := false
				stopSubScriber := event_center.NewSubscriber(
					func(eventName string, values ...interface{}) {
						isStopped = true
					})

				EventCenter.Subscribe(EVENT_CMD_STOP, stopSubScriber)
				defer EventCenter.Unsubscribe(EVENT_CMD_STOP, stopSubScriber.Id())

				for !isStopped {
					time.Sleep(sleepTime)
					msg.CurrentTime = time.Now()
					msg.HeartbeatCount += 1
					beater.Heartbeat(msg)
				}
			}(heartbeater, heartbeatMessage, time.Duration(heartbeatConf.Interval)*time.Millisecond)
		} else {
			err = ERR_HEARTBEAT_NOT_EXIST.New(errors.Params{"name": heartbeatConf.Heart})
			return
		}
	}

	return
}

func (p *ClassicSpirit) getHeartbeatMessage() (message HeartbeatMessage) {
	hostName := ""
	if name, e := os.Hostname(); e != nil {
		panic(e)
	} else {
		hostName = name
	}

	message.InstanceName = p.instanceName
	message.HostName = hostName
	message.StartTime = time.Now()
	message.PID = syscall.Getpid()

	return
}

func (p *ClassicSpirit) getRunComponents(c *cli.Context) (components []string, err error) {
	runComponents := []string{}
	if c.Args() != nil || len(c.Args()) > 0 {
		for _, runComponentName := range c.Args() {
			if _, exist := p.hostingComponents[runComponentName]; !exist {
				err = ERR_SPIRIT_COMPONENT_NOT_HOSTING.New(errors.Params{"name": runComponentName})
				return
			}
			runComponents = append(runComponents, runComponentName)
		}
	}

	if len(runComponents) == 0 {
		err = ERR_SPIRIT_NO_COMPONENTS_TO_RUN.New()
		return
	}

	components = runComponents

	return
}

func (p *ClassicSpirit) waitSignal() {
	isStopping := false
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGUSR1)

	for {
		select {
		case signal := <-interrupt:
			logs.Info("singal received", signal)
			switch signal {
			case os.Interrupt, syscall.SIGTERM:
				{
					if isStopping {
						os.Exit(0)
					}

					isStopping = true
					EventCenter.PushEvent(EVENT_CMD_STOP)

					go func() {
						for _, runningComponent := range p.runningComponents {
							go runningComponent.Stop()
						}

						wg := sync.WaitGroup{}

						for _, runningComponent := range p.runningComponents {
							wg.Add(1)

							go func(component Component) {
								defer wg.Done()

								for component.Status() != STATUS_STOPPED {
									time.Sleep(time.Second)
								}
								logs.Info(fmt.Sprintf("[spirit] component %s was gracefully stoped\n", component.Name()))
							}(runningComponent)
						}
						wg.Wait()
						os.Exit(0)
					}()

				}
			case syscall.SIGUSR1:
				{
					for _, runningComponent := range p.runningComponents {
						runningComponent.PauseOrResume()
						if runningComponent.Status() == STATUS_PAUSED {
							logs.Info(fmt.Sprintf("spirit - component %s was paused\n", runningComponent.Name()))
						} else {
							logs.Info(fmt.Sprintf("spirit - component %s was resumed\n", runningComponent.Name()))
						}
					}
				}
			}
		case <-time.After(time.Second):
			{
				continue
			}
		}
	}
}
