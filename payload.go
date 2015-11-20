package spirit

type Metadata map[string][]interface{}
type Contexts map[string]interface{}

type Payload interface {
	Id() (id string)

	GetData() (data interface{}, err error)
	SetData(data interface{}) (err error)
	DataToObject(v interface{}) (err error)

	GetError() (err error)
	SetError(err error)

	AppendMetadata(name string, values ...interface{}) (err error)
	GetMetadata(name string) (values []interface{}, exist bool)
	Metadata() (metadata Metadata)

	GetContext(name string) (v interface{}, exist bool)
	SetContext(name string, v interface{}) (err error)
	Contexts() (contexts Contexts)
	DeleteContext(name string) (err error)
}
