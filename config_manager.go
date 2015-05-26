package spirit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gogap/errors"
)

type configCache struct {
	MD5    string      `json:"md5"`
	Config interface{} `json:"config"`
}

type configManager struct {
	loadLocker sync.Mutex

	cacheFile string
	Configs   map[string]configCache `json:"configs"`
}

func newConfigManager(instanceName string) *configManager {
	cacheFile := ""

	if instanceName != "" {
		cacheFile = fmt.Sprintf("%s/cache.config", getInstanceHome(instanceName))

		if e := os.MkdirAll(filepath.Dir(cacheFile), os.FileMode(0755)); e != nil {
			panic(e)
		}
	}

	return &configManager{
		cacheFile: cacheFile,
		Configs:   make(map[string]configCache),
	}
}

func (p *configManager) saveCaches() (err error) {
	if p.cacheFile == "" {
		return
	}

	if _, e := os.Stat(p.cacheFile); e != nil {
		if !os.IsNotExist(e) {
			err = ERR_CACHE_FILE_LOAD_FAILED.New(errors.Params{"fileName": p.cacheFile, "err": e})
			return
		}
	} else {
		backupFile := fmt.Sprintf("%s.%d", p.cacheFile, time.Now().Unix())
		if e := os.Rename(p.cacheFile, backupFile); e != nil {
			err = ERR_CACHE_CONFIG_BACKUP_FAILED.New(errors.Params{"fileName": p.cacheFile, "err": err})
			return
		}
	}

	var data []byte
	if data, err = json.MarshalIndent(p, "", "    "); err != nil {
		err = ERR_CACHE_FILE_SERIALIZE_FAILED.New(errors.Params{"fileName": p.cacheFile, "err": err})
		return
	}

	if e := ioutil.WriteFile(p.cacheFile, data, os.FileMode(0644)); e != nil {
		err = ERR_CACHE_FILE_COULD_NOT_SAVE.New(errors.Params{"fileName": p.cacheFile, "err": e})
		return
	}
	return
}

func (p *configManager) serializeCaches() (str string, err error) {
	if p.cacheFile == "" {
		return
	}

	if _, e := os.Stat(p.cacheFile); e != nil {
		if !os.IsNotExist(e) {
			err = ERR_CACHE_FILE_LOAD_FAILED.New(errors.Params{"fileName": p.cacheFile, "err": e})
			return
		}
	}

	var data []byte
	if data, err = json.MarshalIndent(p, "", "    "); err != nil {
		err = ERR_CACHE_FILE_SERIALIZE_FAILED.New(errors.Params{"fileName": p.cacheFile, "err": err})
		return
	}

	str = string(data)

	return
}

func (p *configManager) loadCaches() (err error) {
	if p.cacheFile == "" {
		return
	}

	if _, e := os.Stat(p.cacheFile); e != nil {
		err = ERR_CACHE_FILE_LOAD_FAILED.New(errors.Params{"fileName": p.cacheFile, "err": e})
		return
	}

	if _, e := loadConfig(p.cacheFile, p); e != nil {
		err = ERR_CACHE_FILE_LOAD_FAILED.New(errors.Params{"fileName": p.cacheFile, "err": e})
		return
	}

	return
}

func (p *configManager) Get(fileName string) (v interface{}, err error) {
	if conf, exist := p.Configs[fileName]; !exist {
		err = ERR_CONFIG_FILE_NOT_CACHED.New(errors.Params{"fileName": fileName})
		return
	} else {
		v = conf.Config
	}

	return
}

func (p *configManager) TmpFile(fileName string) (tmpFile string, err error) {
	if conf, exist := p.Configs[fileName]; !exist {
		err = ERR_CONFIG_FILE_NOT_CACHED.New(errors.Params{"fileName": fileName})
		return
	} else {

		tmpDir := getSpiritTmp()

		if e := os.MkdirAll(tmpDir, 0744); e != nil {
			err = ERR_COULD_NOT_MAKE_SPIRIT_TMP_DIR.New(errors.Params{"err": e})
			return
		}

		tmpFilePath := tmpDir + "/" + conf.MD5 + ".conf"

		if data, e := json.MarshalIndent(conf.Config, "", "    "); e != nil {
			if e := ioutil.WriteFile(tmpFilePath, data, 0644); e != nil {
				err = ERR_MARSHAL_CONFIG_FAILED.New(errors.Params{"fileName": fileName, "err": e})
				return
			}
		} else {
			if e := ioutil.WriteFile(tmpFilePath, data, 0644); e != nil {
				err = ERR_WRITE_TMP_CONFIG_FILE_FAILED.New(errors.Params{"fileName": fileName, "err": e})
				return
			}
		}

		tmpFile = tmpFilePath
	}
	return
}

func (p *configManager) load(fileName string, v interface{}) (err error) {
	p.loadLocker.Lock()
	defer p.loadLocker.Unlock()

	md5 := ""
	if md5, err = loadConfig(fileName, &v); err != nil {
		return
	}

	if oldConf, exist := p.Configs[fileName]; exist {
		if oldConf.MD5 != md5 {
			err = ERR_CONFIG_ALREADY_LOADED.New(errors.Params{"fileName": fileName})
		}
		return
	}

	p.Configs[fileName] = configCache{Config: v, MD5: md5}

	return
}
