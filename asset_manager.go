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

type assetCache struct {
	MD5  string `json:"md5"`
	Data []byte `json:"data"`
}

type assetsManager struct {
	loadLocker sync.Mutex

	assetMetadataFile string
	assets            map[string]assetCache
}

func newAssetsManager(instanceName string) *assetsManager {
	assetMetadataFile := ""

	if instanceName != "" {
		assetMetadataFile = fmt.Sprintf("%s/assets.config", getInstanceHome(instanceName))

		if e := os.MkdirAll(filepath.Dir(assetMetadataFile), os.FileMode(0755)); e != nil {
			panic(e)
		}
	}

	return &assetsManager{
		assetMetadataFile: assetMetadataFile,
		assets:            make(map[string]assetCache),
	}
}

func (p *assetsManager) save() (err error) {
	if p.assetMetadataFile == "" {
		return
	}

	if _, e := os.Stat(p.assetMetadataFile); e != nil {
		if !os.IsNotExist(e) {
			err = ERR_CACHE_FILE_LOAD_FAILED.New(errors.Params{"fileName": p.assetMetadataFile, "err": e})
			return
		}
	} else {
		backupFile := fmt.Sprintf("%s.%d", p.assetMetadataFile, time.Now().Unix())
		if e := os.Rename(p.assetMetadataFile, backupFile); e != nil {
			err = ERR_CACHE_ASSET_BACKUP_FAILED.New(errors.Params{"fileName": p.assetMetadataFile, "err": err})
			return
		}
	}

	strJosn := ""
	if strJosn, err = p.serializeAssets(); err != nil {
		return
	}

	if e := ioutil.WriteFile(p.assetMetadataFile, []byte(strJosn), os.FileMode(0644)); e != nil {
		err = ERR_CACHE_FILE_COULD_NOT_SAVE.New(errors.Params{"fileName": p.assetMetadataFile, "err": e})
		return
	}
	return
}

func (p *assetsManager) serializeAssets() (str string, err error) {
	if p.assetMetadataFile == "" {
		return
	}

	if _, e := os.Stat(p.assetMetadataFile); e != nil {
		if !os.IsNotExist(e) {
			err = ERR_CACHE_FILE_LOAD_FAILED.New(errors.Params{"fileName": p.assetMetadataFile, "err": e})
			return
		}
	}

	var tmp struct {
		Assets map[string]assetCache `json:"assets"`
	}

	tmp.Assets = p.assets

	var data []byte
	if data, err = json.MarshalIndent(&tmp, "", "    "); err != nil {
		err = ERR_CACHE_FILE_SERIALIZE_FAILED.New(errors.Params{"fileName": p.assetMetadataFile, "err": err})
		return
	}

	str = string(data)

	return
}

func (p *assetsManager) loadAssets() (err error) {
	if p.assetMetadataFile == "" {
		return
	}

	if _, e := os.Stat(p.assetMetadataFile); e != nil {
		err = ERR_CACHE_FILE_LOAD_FAILED.New(errors.Params{"fileName": p.assetMetadataFile, "err": e})
		return
	}

	var tmp struct {
		Assets map[string]assetCache `json:"assets"`
	}

	if _, e := loadConfig(p.assetMetadataFile, &tmp); e != nil {
		err = ERR_CACHE_FILE_LOAD_FAILED.New(errors.Params{"fileName": p.assetMetadataFile, "err": e})
		return
	}

	p.assets = tmp.Assets

	return
}

func (p *assetsManager) Get(fileName string) (data []byte, err error) {
	if fileName, err = p.getCacheFileName(fileName); err != nil {
		return
	}

	if asset, exist := p.assets[fileName]; !exist {
		err = ERR_ASSET_FILE_NOT_CACHED.New(errors.Params{"fileName": fileName})
		return
	} else {
		data = asset.Data
	}

	return
}

func (p *assetsManager) Unmarshal(fileName string, v interface{}) (err error) {
	var data []byte
	if data, err = p.Get(fileName); err != nil {
		return
	}

	if e := json.Unmarshal(data, v); e != nil {
		err = ERR_JSON_UNMARSHAL.New(errors.Params{"err": e})
		return
	}

	return
}

func (p *assetsManager) getCacheFileName(fileName string) (cacheFileName string, err error) {
	relFileName := ""
	absFileName := ""

	if !filepath.IsAbs(fileName) {
		if absFileName, err = filepath.Abs(fileName); err != nil {
			err = ERR_GET_FILE_ABS_PATH_FAILED.New(errors.Params{"fileName": fileName, "err": err})
			return
		}
	} else {
		absFileName = fileName
	}

	if relFileName, err = filepath.Rel(getAbsBinHome(), absFileName); err != nil {
		err = ERR_GET_FILE_REL_PATH_FAILED.New(errors.Params{"fileName": fileName, "err": err})
		return
	}

	cacheFileName = relFileName
	return
}

func (p *assetsManager) load(fileName string, build bool) (err error) {
	p.loadLocker.Lock()
	defer p.loadLocker.Unlock()

	if fileName, err = p.getCacheFileName(fileName); err != nil {
		return
	}

	md5 := ""
	var data []byte
	if md5, data, err = readFile(fileName, build); err != nil {
		return
	}

	if oldAsset, exist := p.assets[fileName]; exist {
		if oldAsset.MD5 != md5 {
			err = ERR_ASSET_ALREADY_LOADED.New(errors.Params{"fileName": fileName})
			return
		}
		return
	}

	p.assets[fileName] = assetCache{Data: data, MD5: md5}

	return
}
