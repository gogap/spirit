package spirit

type HookEvent int32

const (
	HookEventBeforeCallHandler HookEvent = 1
	HookEventAfterCallHandler  HookEvent = 2
)

type MessageHookMetadata map[string]interface{}

type MessageHook interface {
	Init(configFile string) error
	Name() string
	Hook(event HookEvent,
		previousMetadatas []MessageHookMetadata,
		currentMetadatas []MessageHookMetadata,
		payload *Payload) (ignored bool, newMetaData MessageHookMetadata, err error)
}
