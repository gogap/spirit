package spirit

type Metadata map[string]interface{}
type Context map[string]interface{}

type Error struct {
	Code       uint64
	Id         string
	Namespace  string
	Message    string
	StackTrace string
	Context    Context
}

func (p Error) Error() string {
	return p.Message
}

type Payload interface {
	Id() (id string)

	GetData() (data interface{}, err error)
	SetData(data interface{}) (err error)
	DataToObject(v interface{}) (err error)

	Errors() (err []*Error)
	AppendError(err ...*Error)
	LastError() (err *Error)
	ClearErrors()

	GetContext(name string) (v interface{}, exist bool)
	SetContext(name string, v interface{}) (err error)
	Context() (context Context)
	DeleteContext(name string) (err error)
}
