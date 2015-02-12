package spirit

import (
	"github.com/gogap/errors"
)

const SPIRIT_ERR_NS = "SPIRIT"

var (
	ERR_MESSAGE_GRAPH_ADDRESS_NOT_FOUND = errors.TN(SPIRIT_ERR_NS, 1, "could not found next address in graph, current component name: {{.compName}}, port name: {{.portName}}")
	ERR_MESSAGE_GRAPH_IS_NIL            = errors.TN(SPIRIT_ERR_NS, 2, "component graph is nil")
	ERR_MESSAGE_ADDRESS_IS_EMPTY        = errors.TN(SPIRIT_ERR_NS, 3, "component message address is empty")
	ERR_MESSAGE_SERIALIZE_FAILED        = errors.TN(SPIRIT_ERR_NS, 4, "component message serialize failed, err: {{.err}}")

	ERR_COMPONENT_HANDLER_RETURN_ERROR = errors.TN(SPIRIT_ERR_NS, 5, "component handler return error, error: {{.err}}")
	ERR_COMPONENT_HANDLER_NOT_EXIST    = errors.TN(SPIRIT_ERR_NS, 6, "component handler not exist, component name: {{.name}},  handler name: {{.handlerName}}")

	ERR_SENDER_CREDENTIAL_IS_NIL = errors.TN(SPIRIT_ERR_NS, 8, "credential is nil, type: {{.type}}, url: {{.url}}")
	ERR_SENDER_MQS_CLIENT_IS_NIL = errors.TN(SPIRIT_ERR_NS, 9, "sender mqs client is nil, type: {{.type}}, url: {{.url}}")
	ERR_SENDER_SEND_FAILED       = errors.TN(SPIRIT_ERR_NS, 10, "component message send failed, type: {{.type}}, url: {{.url}}, err:{{.err}}")

	ERR_SENDER_CREATE_FAILED    = errors.TN(SPIRIT_ERR_NS, 11, "create sender failed, driver type: {{.type}}")
	ERR_SENDER_DRIVER_NOT_EXIST = errors.TN(SPIRIT_ERR_NS, 12, "message sender driver not exist, type: {{.type}}")
	ERR_SENDER_BAD_DRIVER       = errors.TN(SPIRIT_ERR_NS, 13, "bad message sender driver of {{.type}}")

	ERR_RECEIVER_CREDENTIAL_IS_NIL     = errors.TN(SPIRIT_ERR_NS, 14, "credential is nil, type: {{.type}}")
	ERR_RECEIVER_MQS_CLIENT_IS_NIL     = errors.TN(SPIRIT_ERR_NS, 15, "receiver mqs client is nil, type: {{.type}}, url: {{.url}}")
	ERR_RECEIVER_DELETE_MSG_ERROR      = errors.TN(SPIRIT_ERR_NS, 16, "delete message error, type: {{.type}}, url: {{.url}}, err: {{.err}}")
	ERR_RECEIVER_UNMARSHAL_MSG_FAILED  = errors.TN(SPIRIT_ERR_NS, 17, "receiver unmarshal message failed, type: {{.type}}, url: {{.url}}, err: {{.err}}")
	ERR_RECEIVER_RECV_ERROR            = errors.TN(SPIRIT_ERR_NS, 18, "receiver recv error, type: {{.type}}, url: {{.url}}, err: {{.err}}")
	ERR_RECEIVER_BAD_DRIVER            = errors.TN(SPIRIT_ERR_NS, 19, "bad message receiver driver of {{.type}}")
	ERR_RECEIVER_CREATE_FAILED         = errors.TN(SPIRIT_ERR_NS, 20, "create receiver failed, driver type: {{.type}}, url: {{.url}}")
	ERR_RECEIVER_DRIVER_NOT_EXIST      = errors.TN(SPIRIT_ERR_NS, 21, "message receiver driver not exist, type: {{.type}}")
	ERR_RECEIVER_DECODE_MESSAGE_FAILED = errors.TN(SPIRIT_ERR_NS, 22, "message receiver decode message failed, type: {{.type}}, url: {{.url}}, err: {{.err}}")

	ERR_PORT_NAME_IS_EMPTY    = errors.TN(SPIRIT_ERR_NS, 23, "port name is empty, component: {{.name}}")
	ERR_HANDLER_NAME_IS_EMPTY = errors.TN(SPIRIT_ERR_NS, 24, "handler name is empty, component: {{.name}}")
)
