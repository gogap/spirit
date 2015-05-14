package spirit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gogap/errors"
	"github.com/nightlyone/lockfile"
)

const (
	INSTANCE_RUNNING = "Running"
	INSTANCE_EXITED  = "Exited"
)

type instanceStatus struct {
	pid    string
	name   string
	status string
	hash   string
}

type instanceBin struct {
	hash       string
	message    string
	createTime string
}

type instanceMetadata struct {
	Envs          []string      `json:"envs"`
	RunConfig     runtimeConfig `json:"runtime_config"`
	RunComponents []string      `json:"run_components"`
	LastStartTime time.Time     `json:"last_start_time"`
	CreateTime    time.Time     `json:"create_time"`
	UpdateTime    time.Time     `json:"update_time"`
	RunTimes      int64         `json:"run_times"`
	CurrentHash   string        `json:"current_hash"`
}

func (p *instanceMetadata) Serialize() (str string, err error) {
	if data, e := json.MarshalIndent(p, " ", "  "); e != nil {
		err = ERR_JSON_MARSHAL.New(errors.Params{"err": e})
		return
	} else {
		str = string(data)
		return
	}
}

type instanceManager struct {
	spiritName string
}

func newInstanceManager(spiritName string) *instanceManager {
	return &instanceManager{
		spiritName: spiritName,
	}
}

func (p *instanceManager) SaveMetadata(instanceName string, metadata instanceMetadata) (err error) {
	if instanceName == "" {
		err = ERR_INSTANCE_NAME_IS_EMPTY.New()
		return
	}

	instanceHome := getInstanceHome(instanceName)

	if e := os.MkdirAll(instanceHome, os.FileMode(0755)); e != nil {
		err = ERR_INSTANCE_HOME_CREATE_FAILED.New(errors.Params{"dir": instanceHome})
		return
	}

	metadata.UpdateTime = time.Now()

	strMeta := ""
	if strMeta, err = metadata.Serialize(); err != nil {
		return
	}

	if e := ioutil.WriteFile(instanceHome+"/metadata", []byte(strMeta), os.FileMode(0644)); e != nil {
		err = ERR_SAVE_INSTANCE_METADATA_FAILED.New(errors.Params{"name": instanceName, "err": e})
		return
	}

	return
}

func (p *instanceManager) GetInstanceCurBinPath(instanceName string) (filePath string, err error) {
	var metadata *instanceMetadata
	if metadata, err = p.GetMetadata(instanceName); err != nil {
		return
	}

	if metadata.CurrentHash == "" {
		err = ERR_INSTANCE_HASH_IS_EMPTY.New(errors.Params{"name": instanceName})
		return
	}

	filePath = p.GetInstanceBinPath(instanceName, metadata.CurrentHash)

	return
}

func (p *instanceManager) GetInstanceBinPath(instanceName string, hash string) (filePath string) {

	instanceHome := getAbsInstanceHome(instanceName)

	filePath = fmt.Sprintf("%s/%s.%s", instanceHome, instanceName, hash)

	return
}

func (p *instanceManager) ListInstanceBins(instanceName string) (bins []instanceBin, err error) {
	if instanceName == "" {
		err = ERR_INSTANCE_NAME_IS_EMPTY.New()
		return
	}

	instanceHome := getInstanceHome(instanceName)

	var files []os.FileInfo
	if files, err = ioutil.ReadDir(instanceHome); err != nil {
		return
	}

	instBins := []instanceBin{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if strings.HasPrefix(file.Name(), instanceName+".") &&
			filepath.Ext(file.Name()) != ".msg" {

			hash := strings.TrimPrefix(file.Name(), instanceName+".")

			binMsgPath := p.GetInstanceBinPath(instanceName, hash) + ".msg"

			binMsg := ""
			if _, e := os.Stat(binMsgPath); e == nil {
				if data, e := ioutil.ReadFile(binMsgPath); e == nil {
					binMsg = string(data)
				}
			}

			bin := instanceBin{
				hash:       hash,
				message:    binMsg,
				createTime: file.ModTime().Format("2006-01-02 15:04:05"),
			}

			instBins = append(instBins, bin)
		}
	}

	bins = instBins

	return
}

func (p *instanceManager) CommitBinary(instanceName string, message string) (err error) {
	if instanceName == "" {
		err = ERR_INSTANCE_NAME_IS_EMPTY.New()
		return
	}

	newBinHash := getProcBinHash()

	filePath := p.GetInstanceBinPath(instanceName, newBinHash)

	if _, e := os.Stat(filePath); e != nil {
		if !os.IsNotExist(e) {
			err = ERR_GET_INSTNACE_COMMIT_INFO_ERROR.New(errors.Params{"name": instanceName})
			return
		}
	} else {
		err = ERR_INSTANCE_BIN_ALREADY_COMMIT.New(errors.Params{"name": instanceName, "hash": newBinHash})
		return
	}

	if _, e := copyFile(filePath, os.Args[0], 0744); e != nil {
		err = ERR_SPIRIT_COPY_INSTANCE_ERROR.New(errors.Params{"err": e})
		return
	}

	if message != "" {
		if e := ioutil.WriteFile(filePath+".msg", []byte(message), 0644); e != nil {
			err = ERR_WRITE_COMMIT_MESSAGE_ERROR.New(errors.Params{"err": e})
			return
		}
	}

	var metadata *instanceMetadata
	if metadata, err = p.GetMetadata(instanceName); err != nil {
		return
	}

	metadata.CurrentHash = newBinHash

	err = p.SaveMetadata(instanceName, *metadata)

	return
}

func (p *instanceManager) CheckoutBinary(instanceName string, hash string) (msg string, err error) {
	if instanceName == "" {
		err = ERR_INSTANCE_NAME_IS_EMPTY.New()
		return
	}

	filePath := p.GetInstanceBinPath(instanceName, hash)
	msgFilePath := filePath + ".msg"

	binMsg := ""
	if _, e := os.Stat(msgFilePath); e == nil {
		if data, e := ioutil.ReadFile(msgFilePath); e == nil {
			binMsg = string(data)
		}
	}

	var metadata *instanceMetadata
	if metadata, err = p.GetMetadata(instanceName); err != nil {
		return
	}

	if metadata.CurrentHash == hash {
		msg = binMsg
		return
	}

	metadata.CurrentHash = hash

	if err = p.SaveMetadata(instanceName, *metadata); err != nil {
		return
	}

	msg = binMsg

	return
}

func (p *instanceManager) RemoveBinary(instanceName string, hash string) (err error) {
	if instanceName == "" {
		err = ERR_INSTANCE_NAME_IS_EMPTY.New()
		return
	}

	filePath := p.GetInstanceBinPath(instanceName, hash)

	var metadata *instanceMetadata
	if metadata, err = p.GetMetadata(instanceName); err != nil {
		return
	}

	if metadata.CurrentHash == hash {
		err = ERR_MUST_CHECKOUT_BEFORE_REMOVE.New(errors.Params{"hash": hash})
		return
	}

	os.Remove(filePath + ".msg")

	if e := os.Remove(filePath); e != nil {
		err = ERR_REMOVE_BIN_ERROR.New(errors.Params{"hash": hash, "err": err})
		return
	}

	return
}

func (p *instanceManager) GetMetadata(instanceName string) (metadata *instanceMetadata, err error) {
	if instanceName == "" {
		err = ERR_INSTANCE_NAME_IS_EMPTY.New()
		return
	}

	instanceHome := getInstanceHome(instanceName)

	instanceFile := instanceHome + "/metadata"
	if _, e := os.Stat(instanceFile); e != nil {
		err = ERR_LOAD_INSTANCE_METADATA_FAILED.New(errors.Params{"name": instanceName, "err": e})
		return
	}

	data := new(instanceMetadata)
	if _, e := loadConfig(instanceFile, data); e != nil {
		err = ERR_LOAD_INSTANCE_METADATA_FAILED.New(errors.Params{"name": instanceName, "err": e})
		return
	}

	metadata = data
	return
}

func (p *instanceManager) TryLockPID(instanceName string) (err error) {
	pidPath := getSpiritHome(p.spiritName) + "/" + instanceName

	if lckFile, e := lockfile.New(pidPath); e != nil {
		err = ERR_SPIRIT_WRITE_PID_FILE_ERROR.New(errors.Params{"err": e})
		return
	} else if e := lckFile.TryLock(); e != nil {
		err = ERR_SPIRIT_LOCK_PID_FILE_FAILED.New(errors.Params{"name": instanceName, "err": e})
		return
	}

	return
}

func (p *instanceManager) ReadPID(instanceName string) (pid int, err error) {
	pidPath := getSpiritHome(p.spiritName) + "/" + instanceName

	if lckFile, e := lockfile.New(pidPath); e != nil {
		err = ERR_SPIRIT_WRITE_PID_FILE_ERROR.New(errors.Params{"err": e})
		return
	} else if ps, e := lckFile.GetOwner(); e != nil {
		err = ERR_SPIRIT_GET_INSTANCE_PID_FAILED.New(errors.Params{"name": instanceName, "err": e})
		return
	} else {
		pid = ps.Pid
	}

	return
}

func (p *instanceManager) ListInstanceStatus(all bool) (instances []instanceStatus, err error) {
	spiritHome := getSpiritHome(p.spiritName)

	var files []os.FileInfo

	if files, err = ioutil.ReadDir(spiritHome); err != nil {
		err = ERR_SPIRIT_LIST_INSTANCE_FAILED.New(errors.Params{"err": err})
		return
	}

	inslist := []instanceStatus{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		instanceName := file.Name()

		pidFile := spiritHome + "/" + instanceName

		if _, e := os.Stat(pidFile); e != nil {
			continue
		}

		ins := instanceStatus{
			name: instanceName,
		}

		pid := 0
		strPID := ""
		if d, e := ioutil.ReadFile(pidFile); e != nil {
			err = ERR_SPIRIT_READ_PID_FILE_ERROR.New(errors.Params{"name": instanceName, "err": e})
			return
		} else {
			strPID = string(bytes.TrimSpace(d))
		}

		if p, e := strconv.Atoi(strPID); e != nil {
			err = ERR_SPIRIT_READ_PID_FILE_ERROR.New(errors.Params{"name": instanceName, "err": e})
			return
		} else {
			pid = p
		}

		var metadata *instanceMetadata
		if metadata, err = p.GetMetadata(instanceName); err != nil {
			return
		}

		if isProcessAlive(pid) {
			ins.pid = strPID
			ins.status = INSTANCE_RUNNING
			ins.hash = metadata.CurrentHash
			inslist = append(inslist, ins)
		} else if all {
			ins.pid = "-"
			ins.status = INSTANCE_EXITED
			ins.hash = metadata.CurrentHash
			inslist = append(inslist, ins)
		}
	}

	instances = inslist

	return
}

func (p *instanceManager) StartInstance() {
}

func (p *instanceManager) StopInstance() {
}

func (p *instanceManager) IsInstanceExist(instanceName string) bool {
	if instanceName == "" {
		return false
	}

	pidPath := getSpiritHome(p.spiritName) + "/" + instanceName

	if _, err := os.Stat(pidPath); err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
	} else {
		return true
	}

	metadataFile := getInstanceHome(instanceName) + "/metadata"

	if _, err := os.Stat(metadataFile); err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
		return false
	} else {
		return true
	}
}

func (p *instanceManager) RemoveInstance(instanceName string) (err error) {
	if instanceName == "" {
		err = ERR_SPIRIT_INSTANCE_NAME_NOT_INPUT.New()
		return
	}

	pidPath := getSpiritHome(p.spiritName) + "/" + instanceName

	if _, err = os.Stat(pidPath); err != nil {
		if !os.IsNotExist(err) {
			err = ERR_SPIRIT_REMOVE_INSTANCE_FAILED.New(errors.Params{"name": instanceName, "err": err})
			return
		}
	}

	pid, _ := p.ReadPID(instanceName)
	currentPID := os.Getpid()

	if pid > 0 && pid != currentPID && isProcessAlive(pid) {
		err = ERR_INSTANCE_CAN_NOT_DEL_WHILE_RUN.New(errors.Params{"name": instanceName, "pid": pid})
		return
	}

	if err = os.Remove(pidPath); err != nil {
		err = ERR_SPIRIT_REMOVE_INSTANCE_FAILED.New(errors.Params{"name": instanceName, "err": err})
		return
	}

	instanceHome := getInstanceHome(instanceName)

	if _, err = os.Stat(instanceHome); err != nil {
		if !os.IsNotExist(err) {
			err = ERR_SPIRIT_REMOVE_INSTANCE_FAILED.New(errors.Params{"name": instanceName, "err": err})
			return
		}
	}

	if err = os.RemoveAll(instanceHome); err != nil {
		err = ERR_SPIRIT_REMOVE_INSTANCE_FAILED.New(errors.Params{"name": instanceName, "err": err})
		return
	}

	return nil
}
