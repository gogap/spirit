package spirit

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/gogap/cache_storages"
	"github.com/gogap/env_json"
	"github.com/gogap/errors"
	"github.com/nu7hatch/gouuid"
)

const (
	BIG_DATA_MESSAGE_ID           = "BIG_DATA_MESSAGE_ID"
	MAX_DATA_LENGTH               = 1024 * 30 //30k
	BIG_DATA_REDIS_EXPIRE_SECONDS = 600
)

var redisStorage cache_storages.CacheStorage

type MessageHookBigDataRedis struct{}

func (p *MessageHookBigDataRedis) Init(configFile string) (err error) {
	var config redisStorageConf
	r, err := os.Open(configFile)
	if err != nil {
		err = ERR_HOOK_BIG_DATA_OPEN_FILE.New(errors.Params{"err": err.Error()})
		return
	}
	defer r.Close()

	data, err := ioutil.ReadAll(r)
	if err != nil {
		err = ERR_HOOK_BIG_DATA_OPEN_FILE.New(errors.Params{"err": err.Error()})
		return
	}

	envJson := env_json.NewEnvJson(ENV_NAME, ENV_EXT)
	if err = envJson.Unmarshal(data, &config); err != nil {
		err = ERR_HOOK_BIG_DATA_UNMARSHAL_CONF.New(errors.Params{"err": err.Error()})
		return
	}
	if config.Auth == "" {
		redisStorage, err = cache_storages.NewRedisStorage(config.Conn, config.Index)
		if err != nil {
			err = ERR_HOOK_BIG_DATA_CREATE_REDIS.New(errors.Params{"err": err.Error()})
			return
		}
	} else {
		redisStorage, err = cache_storages.NewAuthRedisStorage(config.Conn, config.Index, config.Auth)
		if err != nil {
			err = ERR_HOOK_BIG_DATA_CREATE_REDIS.New(errors.Params{"err": err.Error()})
			return
		}
	}
	return
}

func (p *MessageHookBigDataRedis) Name() string {
	return "message_hook_big_data_redis"
}

func (p *MessageHookBigDataRedis) HookBefore(
	currentMetadata MessageHookMetadata,
	previousMetadatas []MessageHookMetadata,
	contextMetadatas []MessageHookMetadata,
	payload *Payload) (ignored bool, newMetaData MessageHookMetadata, err error) {

	for key, value := range currentMetadata.Context {
		if key == BIG_DATA_MESSAGE_ID {
			var data string
			data, err = redisStorage.Get(value.(string))
			if err != nil {
				err = ERR_HOOK_BIG_DATA_REDIS_GET.New(errors.Params{"err": err.Error()})
				return
			}
			var container interface{}
			err = json.Unmarshal([]byte(data), &container)
			if err != nil {
				err = ERR_JSON_UNMARSHAL.New(errors.Params{"err": err.Error()})
				return
			}
			if data != "" {
				payload.SetContent(container)
			}
		}
	}
	return
}

func (p *MessageHookBigDataRedis) HookAfter(
	previousMetadatas []MessageHookMetadata,
	contextMetadatas []MessageHookMetadata,
	payload *Payload) (ignored bool, newMetaData MessageHookMetadata, err error) {

	bit, err := json.Marshal(payload.GetContent())
	if err != nil {
		err = ERR_JSON_MARSHAL.New(errors.Params{"err": err.Error()})
		return
	}

	if len(bit) > MAX_DATA_LENGTH {
		messageId := getUUID()
		err = redisStorage.Set(messageId, string(bit), BIG_DATA_REDIS_EXPIRE_SECONDS)
		if err != nil {
			err = ERR_HOOK_BIG_DATA_REDIS_SET.New(errors.Params{"err": err.Error()})
			return
		}
		newMetaData.Context = make(map[string]interface{})
		newMetaData.Context[BIG_DATA_MESSAGE_ID] = messageId
		payload.SetContent(nil)
		return
	}
	ignored = true
	return
}

type redisStorageConf struct {
	Auth  string `json:"auth"`
	Conn  string `json:"conn"`
	Index int    `json:"index"`
}

func getUUID() string {
	id, _ := uuid.NewV4()
	return id.String()
}
