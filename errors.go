package spirit

import (
	"github.com/gogap/errors"
)

const SpiritErrNamespace = "SPIRIT"

var (
	ErrReceiverURNNotExist          = errors.New("receiver urn not exist")
	ErrReceiverNameDuplicate        = errors.New("receiver name duplicate")
	ErrReceiverAlreadyRunning       = errors.New("receiver already running")
	ErrReceiverDidNotHaveTranslator = errors.New("receiver did not have translator")
	ErrReceiverDidNotRunning        = errors.New("receiver did not running")
	ErrReceiverReceiveTimeout       = errors.New("receiver receive timeout")
	ErrReceiverDeliveryPutterIsNil  = errors.New("receiver delivery putter is nil")
	ErrReceiverDidNotHaveReaderPool = errors.New("receiver did not have reader pool")
	ErrReceiverTypeNotSupport       = errors.New("receiver type not support")

	ErrSenderURNNotExist          = errors.New("sender urn not exist")
	ErrSenderNameDuplicate        = errors.New("sender name duplicate")
	ErrSenderAlreadyRunning       = errors.New("sender already running")
	ErrSenderCanNotCreaterWriter  = errors.New("sender can not create writer")
	ErrSenderDidNotHaveTranslator = errors.New("sender did not have translator")
	ErrSenderDidNotRunning        = errors.New("sender did not running")
	ErrSenderDeliveryGetterIsNil  = errors.New("sender delivery getter is nil")
	ErrSenderDidNotHaveWriterPool = errors.New("sender did not have writer pool")
	ErrSenderTypeNotSupport       = errors.New("sender type not support")

	ErrRouterURNNotExist                     = errors.New("router urn not exist")
	ErrRouterNameDuplicate                   = errors.New("router name duplicate")
	ErrRouterAlreadyRunning                  = errors.New("router already running")
	ErrRouterDidNotRunning                   = errors.New("router did not running")
	ErrRouterAlreadyHaveThisInbox            = errors.New("router already added this inbox")
	ErrRouterAlreadyHaveThisOutbox           = errors.New("router already added this outbox")
	ErrRouterAlreadyHaveThisComponent        = errors.New("router already added this component")
	ErrRouterToComponentHandlerFailed        = errors.New("could not router to component handler")
	ErrRouterToComponentFailed               = errors.New("could not router to component")
	ErrRouterComponentNotExist               = errors.New("router component not exist")
	ErrRouterDidNotHaveComponentLabelMatcher = errors.New("could not router to component")
	ErrRouterHandlerCountNotEqualURNsCount   = errors.New("handler's count is not equal delivery urn's count, some component may not add")
	ErrRouterDeliveryURNFormatError          = errors.New("delivery urn format error")
	ErrRouterOnlyOneGlobalComponentAllowed   = errors.New("only one same urn global component allowed")

	ErrComponentURNNotExist       = errors.New("component urn not exist")
	ErrComponentNameDuplicate     = errors.New("component name duplicate")
	ErrComponentNameIsEmpty       = errors.New("component name is empty")
	ErrComponentURNIsEmpty        = errors.New("component URN is empty")
	ErrComponentNotExist          = errors.New("component not exist")
	ErrComponentHandlerNotExit    = errors.New("component handler not exist")
	ErrComponentDidNotHaveHandler = errors.New("component did not have handler")

	ErrDeliveryURNIsEmpty = errors.New("delivery urn is empty")

	ErrReaderURNNotExist   = errors.New("reader urn not exist")
	ErrReaderNameDuplicate = errors.New("reader name duplicate")

	ErrWriterURNNotExist   = errors.New("writer urn not exist")
	ErrWriterNameDuplicate = errors.New("writer name duplicate")

	ErrWriterIsNil              = errors.New("writer is nil")
	ErrWriterInUse              = errors.New("writer already in use")
	ErrWriterPoolAlreadyClosed  = errors.New("writer pool already closed")
	ErrWriterPoolNameDuplicate  = errors.New("writer pool name duplicate")
	ErrWriterPoolURNNotExist    = errors.New("writer pool urn not exist")
	ErrWriterPoolTooManyWriters = errors.New("writer pool have too many writers")

	ErrReaderIsNil              = errors.New("reader is nil")
	ErrReaderInUse              = errors.New("reader already in use")
	ErrReaderPoolAlreadyClosed  = errors.New("reader pool already closed")
	ErrReaderPoolTooManyReaders = errors.New("reader pool have too many readers")
	ErrReaderPoolNameDuplicate  = errors.New("reader pool name duplicate")
	ErrReaderPoolURNNotExist    = errors.New("reader pool urn not exist")

	ErrInboxURNNotExist   = errors.New("inbox urn not exist")
	ErrInboxNameDuplicate = errors.New("inbox name duplicate")

	ErrOutboxURNNotExist   = errors.New("outbox urn not exist")
	ErrOutboxNameDuplicate = errors.New("outbox name duplicate")

	ErrLabelMatcherURNNotExist   = errors.New("label matcher urn not exist")
	ErrLabelMatcherNameDuplicate = errors.New("label matcher name duplicate")

	ErrURNRewriterURNNotExist   = errors.New("urn rewriter urn not exist")
	ErrURNRewriterNameDuplicate = errors.New("urn rewriter name duplicate")

	ErrInputTranslatorURNNotExist   = errors.New("input translator urn not exist")
	ErrInputTranslatorNameDuplicate = errors.New("input translator name duplicate")

	ErrOutputTranslatorURNNotExist   = errors.New("output translator urn not exist")
	ErrOutputTranslatorNameDuplicate = errors.New("output translator name duplicate")

	ErrConsoleURNNotExist   = errors.New("urn console urn not exist")
	ErrConsoleNameDuplicate = errors.New("urn console name duplicate")

	ErrSpiritAlreadyRunning   = errors.New("spirit already running")
	ErrSpiritAlreadyStopped   = errors.New("spirit already stopped")
	ErrSpiritAlreadyBuilt     = errors.New("spirit already built")
	ErrSpiritNotBuild         = errors.New("spirit not build")
	ErrSpiritActorURNNotExist = errors.New("spirit actor urn not exist")

	ErrActorRouterNotExist           = errors.New("actor:router not exist")
	ErrActorComponentNotExist        = errors.New("actor:component not exist")
	ErrActorConsoleNotExist          = errors.New("actor:console not exist")
	ErrActorInBoxNotExist            = errors.New("actor:inbox not exist")
	ErrActorReceiverNotExist         = errors.New("actor:receiver not exist")
	ErrActorInputTranslatorNotExist  = errors.New("actor:input_translator not exist")
	ErrActorOutputTranslatorNotExist = errors.New("actor:output_translator not exist")
	ErrActorReaderPoolNotExist       = errors.New("actor:reader_pool not exist")
	ErrActorWriterNotExist           = errors.New("actor:writer not exist")
	ErrActorWriterPoolNotExist       = errors.New("actor:writer_pool not exist")
	ErrActorOutboxNotExist           = errors.New("actor:outbox not exist")
	ErrActorLabelMatcerNotExist      = errors.New("actor:label_matcher not exist")
	ErrActorSenderNotExist           = errors.New("actor:sender not exist")
	ErrActorURNRewriterNotExist      = errors.New("actor:urn_rewriter not exist")

	ErrComposeNameIsEmpty = errors.New("spirit compose name is empty")

	ErrDefaultLogerNotExist = errors.New("defualt logger not exist")
)
