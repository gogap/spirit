Example
=======

## Quick start

> You should have aliyun mns account first, http://www.aliyun.com/product/mns

currenty we only support aliyun mns for message queue engine, more message queue engine will coming soon.....

```bash
./example run --name todo -a port.new|new_task|mns|access_key_id:acces_key_secert@http://owner_id.mns-cn-hangzhou.aliyuncs.com/todo-new-task --alias todo_new_task
```

## More

#### Configure the `spirit.env` file

```json
{
	"owner_id":"your_owner_id",
	"access_key_id": "your_access_key_id",
	"acces_key_secert": "your_acces_key_secert",
	"mns_url":"mns-cn-hangzhou.aliyuncs.com"
}
```

Export `SPIRIT_ENV` for compile addres values

```
export SPIRIT_ENV=$GOPATH/src/github.com/gogap/spirit/example/spirit.env
```

or

```
export SPIRIT_ENV=$GOPATH/src/github.com/gogap/spirit/example
```

#### Start the example

```bash
./example run --name todo -a 'port.new|new_task|mns|{{.access_key_id}}:{{.acces_key_secert}}@http://{{.owner_id}}.{{.mns_url}}/todo_new_task'
```

> if you do not want use gloabl bash envs, you could specific the params

```bash
./example run --name todo -a 'port.new|new_task|mns|{{.access_key_id}}:{{.acces_key_secert}}@http://{{.owner_id}}.{{.mns_url}}/todo_new_task' -e SPIRIT_ENV=$GOPATH/src/github.com/gogap/spirit/example
```


