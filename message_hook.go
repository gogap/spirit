package spirit

type HookEvent int32

const (
	HookEventBeforeCallHandler HookEvent = 1
	HookEventAfterCallHandler  HookEvent = 2
)

type MessageHookMetadata struct {
	HookName string                 `json:"hook_name"`
	Context  map[string]interface{} `json:"context"`
}

type MessageHook interface {
	Init(configFile string) error
	Name() string
	HookBefore(
		currentMetadata MessageHookMetadata,
		previousMetadatas []MessageHookMetadata,
		contextMetadatas []MessageHookMetadata,
		payload *Payload) (ignored bool, newMetaData MessageHookMetadata, err error)

	HookAfter(
		previousMetadatas []MessageHookMetadata,
		contextMetadatas []MessageHookMetadata,
		payload *Payload) (ignored bool, newMetaData MessageHookMetadata, err error)
}
