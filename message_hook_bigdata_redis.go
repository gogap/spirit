package spirit

type MessageHookBigDataRedis struct {
}

func (p *MessageHookBigDataRedis) Init(configFile string) (err error) {
	return
}

func (p *MessageHookBigDataRedis) Name() string {
	return "message_hook_big_data_redis"
}

func (p *MessageHookBigDataRedis) Hook(event HookEvent,
	previousMetadatas []MessageHookMetadata,
	currentMetadatas []MessageHookMetadata,
	payload *Payload) (ignored bool, newMetaData MessageHookMetadata, err error) {
	return
}
