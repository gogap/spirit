package std

import (
	"bufio"
	"fmt"
	"sync"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/gogap/spirit"
)

type ioType int

const (
	_Input  ioType = 0
	_Output        = 1
)

var (
	procLocker    sync.Mutex
	newProcLocker sync.Mutex

	allProcs map[string][]*STDProcess = make(map[string][]*STDProcess)
)

type _Out struct {
	output []byte
	err    error
}

type STDProcess struct {
	takeLock sync.Mutex
	rwLock   sync.Mutex

	procMetadata *ProcessMetadata

	out        chan _Out
	terminated chan bool

	isInputTaken  bool
	isOutputTaken bool

	conf StdIOConfig

	delim string

	isRunning bool
}

func takeSTDIO(typ ioType, conf StdIOConfig) (proc *STDProcess, err error) {
	procLocker.Lock()
	defer procLocker.Unlock()

	indexName := conf.Name + "-" + conf.Proc

	if procs, exist := allProcs[indexName]; exist {
		for _, p := range procs {
			if typ == _Input {
				err = p.TakeInput()
			} else {
				err = p.TakeOutput()
			}

			if err == nil {
				proc = p
				return
			}
		}
	}

	if p, e := newProc(conf.Proc, conf.Args, conf.Envs, conf.Dir); e != nil {
		err = e
		return
	} else {
		if typ == _Input {
			p.TakeInput()
		} else {
			p.TakeOutput()
		}

		p.conf = conf

		if procs, exist := allProcs[indexName]; exist {
			allProcs[indexName] = append(procs, p)
		} else {
			allProcs[indexName] = []*STDProcess{p}
		}

		p.delim = conf.delim
		if p.delim == "" {
			p.delim = "\n"
		}

		proc = p
	}

	go func(indexName string, proc *STDProcess) {
		proc.procMetadata.cmd.Wait()

		spirit.Logger().WithFields(logrus.Fields{"actor": "io", "type": "STDProcess", "path": proc.procMetadata.cmd.Path, "args": proc.procMetadata.cmd.Args}).Debugln("process terminaled")

		select {
		case proc.terminated <- true:
			{
			}
		case <-time.After(time.Second):
			{
			}
		}

		proc.procMetadata.cmd.Process.Signal(syscall.SIGINT)

		select {
		case proc.terminated <- true:
			{
			}
		case <-time.After(time.Second):
			{
			}
		}

		procLocker.Lock()
		defer procLocker.Unlock()

		if procs, exist := allProcs[indexName]; exist {
			newProclist := []*STDProcess{}
			for _, p := range procs {
				if p.procMetadata.cmd.Process.Pid !=
					proc.procMetadata.cmd.Process.Pid {
					newProclist = append(newProclist, p)
				}
			}
			allProcs[indexName] = newProclist
		}
	}(indexName, proc)

	return
}

func newProc(name string, args []string, envs []string, dir string) (proc *STDProcess, err error) {
	newProcLocker.Lock()
	defer newProcLocker.Unlock()

	var procMetadata *ProcessMetadata
	if procMetadata, err = ExecuteCommand(name, args, envs, dir); err != nil {
		return
	}

	proc = &STDProcess{
		procMetadata: procMetadata,
		out:          make(chan _Out),
		terminated:   make(chan bool),
	}

	return
}

func (p *STDProcess) TakeInput() (err error) {
	p.takeLock.Lock()
	defer p.takeLock.Unlock()

	if p.isInputTaken == true {
		err = fmt.Errorf("io/std proc input already taken")
		return
	}

	p.isInputTaken = true

	return
}

func (p *STDProcess) TakeOutput() (err error) {
	p.takeLock.Lock()
	defer p.takeLock.Unlock()

	if p.isOutputTaken == true {
		err = fmt.Errorf("io/std proc output already taken")
		return
	}

	p.isOutputTaken = true

	return
}

func (p *STDProcess) Start() {
	p.rwLock.Lock()
	defer p.rwLock.Unlock()

	if p.isRunning {
		return
	}

	p.isRunning = true
	go func() {
		reader := bufio.NewReader(p.procMetadata.stdout)
		delim := byte(p.delim[0])
		for {
			data, err := reader.ReadBytes(delim)
			p.out <- _Out{output: data, err: err}
			if !p.isRunning {
				return
			}
		}
	}()
}

func (p *STDProcess) Stop() {
	p.rwLock.Lock()
	defer p.rwLock.Unlock()

	p.isRunning = false

	procLocker.Lock()
	defer procLocker.Unlock()

	p.procMetadata.cmd.Process.Signal(syscall.SIGINT)

	indexName := p.conf.Name + "-" + p.conf.Proc

	if procs, exist := allProcs[indexName]; exist {
		newProclist := []*STDProcess{}
		for _, proc := range procs {
			if p.procMetadata.cmd.Process.Pid !=
				proc.procMetadata.cmd.Process.Pid {
				newProclist = append(newProclist, p)
			}
		}
		allProcs[indexName] = newProclist
	}
}

func (p *STDProcess) Read(data []byte) (n int, err error) {
	p.rwLock.Lock()
	defer p.rwLock.Unlock()

	select {
	case ret := <-p.out:
		{
			if err = ret.err; err == nil {
				copy(data, ret.output)
				n = len(ret.output)
			}
			return
		}
	case <-time.After(time.Second):
		{
			return
		}
	}

	return
}

func (p *STDProcess) Write(data []byte) (n int, err error) {
	p.rwLock.Lock()
	defer p.rwLock.Unlock()

	n, err = p.procMetadata.stdin.Write(data)

	return
}
