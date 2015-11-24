package spirit

type Metadata map[string]interface{}
type Contexts map[string]interface{}

type Error struct {
	Code       uint64
	Id         string
	Namespace  string
	Message    string
	StackTrace string
	Contexts   Contexts
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
	Contexts() (contexts Contexts)
	DeleteContext(name string) (err error)
}
