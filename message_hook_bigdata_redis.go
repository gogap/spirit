package spirit

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/gogap/cache_storages"
	"github.com/gogap/errors"
)

var (
	_ MessageHook = (*MessageHookBigDataRedis)(nil)
)

const (
	BIG_DATA_MESSAGE_ID           = "BIG_DATA_MESSAGE_ID"
	MAX_DATA_LENGTH               = 1024 * 30 //30k
	BIG_DATA_REDIS_EXPIRE_SECONDS = 600
	REDIS_DATA_SIZE               = 1000 * 1000 * 512 //512 MegaBytes
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

	value, exist := currentMetadata.Context[BIG_DATA_MESSAGE_ID]
	if !exist {
		return
	}

	strMessageIds, _ := value.(string)

	if len(strMessageIds) == 0 {
		return
	}

	messageIds := strings.Split(strMessageIds, ",")

	var values map[string]string
	values, err = redisStorage.GetMulti(messageIds)

	if err != nil {
		err = ERR_HOOK_BIG_DATA_REDIS_GET.New(errors.Params{"err": err.Error()})
		return
	}

	buf := bytes.NewBuffer(nil)
	for i := 0; i < len(messageIds); i++ {
		value, exist := values[messageIds[i]]
		if !exist {
			err = ERR_HOOK_BIG_DATA_REDIS_GET.New(errors.Params{"err": errors.New("redis message dismissed")})
			return
		}

		var data []byte
		data, err = base64.StdEncoding.DecodeString(value)
		if err != nil {
			err = ERR_HOOK_BIG_DATA_REDIS_GET.New(errors.Params{"err": err.Error()})
			return
		}

		buf.Write(data)
	}

	if buf.Len() == 0 {
		return
	}

	var content interface{}

	decoder := json.NewDecoder(buf)
	decoder.UseNumber()

	if e := decoder.Decode(&content); e != nil {
		err = ERR_JSON_UNMARSHAL.New(errors.Params{"err": err.Error()})
		return
	}

	payload.SetContent(content)

	return
}

func (p *MessageHookBigDataRedis) HookAfter(
	previousMetadatas []MessageHookMetadata,
	contextMetadatas []MessageHookMetadata,
	payload *Payload) (ignored bool, newMetaData MessageHookMetadata, err error) {

	data, err := json.Marshal(payload.GetContent())
	if err != nil {
		err = ERR_JSON_MARSHAL.New(errors.Params{"err": err.Error()})
		return
	}

	if len(data) > MAX_DATA_LENGTH {
		buf := bytes.NewBuffer(data)

		var messageIds []string

		for {
			next := buf.Next(REDIS_DATA_SIZE)
			if len(next) == 0 {
				break
			}

			messageId := getUUID()

			err = redisStorage.Set(messageId, base64.StdEncoding.EncodeToString(next), BIG_DATA_REDIS_EXPIRE_SECONDS)
			if err != nil {
				err = ERR_HOOK_BIG_DATA_REDIS_SET.New(errors.Params{"err": err.Error()})
				return
			}

			messageIds = append(messageIds, messageId)
		}

		newMetaData.Context = make(map[string]interface{})
		newMetaData.Context[BIG_DATA_MESSAGE_ID] = strings.Join(messageIds, ",")
		payload.SetContent(nil)
		return
	}

	ignored = true
	return
}
