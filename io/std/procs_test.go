package std

import (
	"fmt"
	"io"
	"testing"

	"github.com/gogap/spirit"
)

func TestStdReadAndWrite(t *testing.T) {

	optsR := spirit.Config{
		"name":  "ping",
		"proc":  "ping",
		"args":  []string{"-c", "10", "www.baidu.com"},
		"envs":  []string{},
		"delim": "\n",
	}

	optsW := spirit.Config{
		"name":  "write",
		"proc":  "my-program-w",
		"args":  []string{},
		"envs":  []string{},
		"delim": "\n",
	}

	var reader io.ReadCloser
	var writer io.WriteCloser
	var err error

	reader, err = NewStdout(optsR)

	if err != nil {
		t.Error(err)
		return
	}

	writer, err = NewStdin(optsW)

	var data = make([]byte, 1024, 1024)

	for {
		_, err = reader.Read(data)

		fmt.Println("Read:", err, string(data))

		writeData := []byte("I Am From Reader - " + string(data) + "\n")
		_, err = writer.Write(writeData)

		fmt.Println("Write:", err, string(writeData))
	}

}
