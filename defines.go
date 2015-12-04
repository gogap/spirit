package spirit

type Status int32

const (
	StatusStopped Status = 0
	StatusRunning Status = 1
	StatusPaused  Status = 2
)

type ActorType string

var (
	ActorReader           ActorType = "reader"
	ActorWriter           ActorType = "writer"
	ActorInputTranslator  ActorType = "input_translator"
	ActorOutputTranslator ActorType = "output_translator"
	ActorReceiver         ActorType = "receiver"
	ActorSender           ActorType = "sender"
	ActorLabelMatcher     ActorType = "label_matcher"
	ActorRouter           ActorType = "router"
	ActorComponent        ActorType = "component"
	ActorInbox            ActorType = "inbox"
	ActorOutbox           ActorType = "outbox"
	ActorReaderPool       ActorType = "reader_pool"
	ActorWriterPool       ActorType = "writer_pool"
	ActorURNRewriter      ActorType = "urn_rewriter"
	ActorConsole          ActorType = "console"
	ActorMessenger        ActorType = "messenger"
)

type Labels map[string]string

const (
	SpiritInitialLogLevelEnvKey = "SPIRIT_INIT_LOG_LEVEL"
)
