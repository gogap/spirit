package spirit

import (
	"encoding/json"

	"github.com/gogap/cache_storages"
	"github.com/gogap/errors"
)

const (
	BIG_DATA_MESSAGE_ID           = "BIG_DATA_MESSAGE_ID"
	MAX_DATA_LENGTH               = 1024 * 60 //60k
	BIG_DATA_REDIS_EXPIRE_SECONDS = 600
)

var redisStorage cache_storages.CacheStorage

type MessageHookBigDataRedis struct{}

func (p *MessageHookBigDataRedis) Init(options Options) (err error) {
	auth := ""
	address := ""
	var db int64 = 0

	if auth, err = options.GetStringValue("auth"); err != nil {
		return
	}

	if address, err = options.GetStringValue("address"); err != nil {
		return
	}

	if db, err = options.GetInt64Value("db"); err != nil {
		return
	}

	if auth == "" {
		if redisStorage, err = cache_storages.NewRedisStorage(address, int(db)); err != nil {
			err = ERR_HOOK_BIG_DATA_CREATE_REDIS.New(errors.Params{"err": err.Error()})
			return
		}
	} else {
		if redisStorage, err = cache_storages.NewAuthRedisStorage(address, int(db), auth); err != nil {
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
