package spirit

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/gogap/env_json"
	"github.com/gogap/errors"
	"github.com/nu7hatch/gouuid"
)

const (
	FILE_CHUNK = 8192
)

var (
	viewDetails = false
)

func initalSpiritHome(spiritName string) (err error) {
	err = os.MkdirAll(getSpiritHome(spiritName), os.FileMode(0755))
	return
}

func getSpiritHome(spiritName string) string {
	tmpDir := strings.TrimRight(os.TempDir(), "/")
	return fmt.Sprintf("%s/%s/%s", tmpDir, SPIRIT, spiritName)
}

func getSpiritTmp() string {
	tmpDir := strings.TrimRight(os.TempDir(), "/")
	return fmt.Sprintf("%s/%s/tmp", tmpDir, SPIRIT)
}

func getInstanceHome(instanceName string) string {
	procDir := filepath.Dir(os.Args[0])

	if isChildInstance() {
		procDir = getBinHome(instanceName)
	}

	return fmt.Sprintf("%s/.%s/.%s", procDir, SPIRIT, instanceName)
}

func getAbsInstanceHome(instanceName string) string {
	instanceHome := getInstanceHome(instanceName)
	if filepath.IsAbs(instanceHome) {
		return instanceHome
	}

	if absPath, err := filepath.Abs(instanceHome); err != nil {
		panic(err)
	} else {
		return absPath
	}
}

func getProcDir() string {
	arg := os.Args[0]
	procDir := filepath.Dir(arg)
	if absPath, err := filepath.Abs(procDir); err != nil {
		panic(err)
	} else {
		return absPath
	}
}

func getBinHome(instanceName string) string {
	return os.Getenv(SPIRIT_BIN_HOME)
}

func isPersistentProcess(instanceName string) bool {
	return getAbsInstanceHome(instanceName) == getProcDir()
}

func isChildInstance() bool {
	if os.Getenv(SPIRIT_CHILD_INSTANCE) == "1" {
		return true
	}
	return false
}

func getFileHash(fileName string) string {
	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	info, _ := file.Stat()

	filesize := info.Size()

	blocks := uint64(math.Ceil(float64(filesize) / float64(FILE_CHUNK)))

	hash := md5.New()

	for i := uint64(0); i < blocks; i++ {
		blocksize := int(math.Min(FILE_CHUNK, float64(filesize-int64(i*FILE_CHUNK))))
		buf := make([]byte, blocksize)

		file.Read(buf)
		io.WriteString(hash, string(buf))
	}

	return fmt.Sprintf("%0x", hash.Sum(nil))
}

func getProcBinHash() string {
	return getFileHash(os.Args[0])
}

func loadConfig(fileName string, v interface{}) (strMD5 string, err error) {
	envJson := env_json.NewEnvJson(ENV_NAME, ENV_EXT)

	var data []byte
	if data, err = ioutil.ReadFile(fileName); err != nil {
		return
	}

	if err = envJson.Unmarshal(data, v); err != nil {
		return
	}

	//make sure the md5 genreate from same struct
	ops := Options{}
	if err = envJson.Unmarshal(data, &ops); err != nil {
		return
	}

	if data, err = envJson.Marshal(ops); err != nil {
		return
	}

	md5Data := md5.Sum(data)

	strMD5 = fmt.Sprintf("%0x", md5Data)

	return
}

func printObject(title string, v interface{}) {
	if v == nil {
		return
	}

	data, _ := env_json.MarshalIndent(v, "", "    ")

	fmt.Println(title+":\n", string(data))
}

func exitError(err error) {
	if errCode, ok := err.(errors.ErrCode); ok {
		if viewDetails {
			fmt.Printf("[ERR-%s-%d] %s \n%s\n", errCode.Namespace(), errCode.Code(), errCode.Error(), errCode.StackTrace())
		} else {
			fmt.Printf("[ERR-%s-%d] %s \n", errCode.Namespace(), errCode.Code(), errCode.Error())
		}
	} else {
		fmt.Printf("[ERR-%s] %s \n", SPIRIT_ERR_NS, err.Error())
	}

	os.Exit(1)
}

func isProcessAlive(pid int) bool {
	p, _ := os.FindProcess(pid)
	if e := p.Signal(syscall.Signal(0)); e == nil {
		return true
	}
	return false
}

func isFileOrDir(filename string, decideDir bool) bool {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return false
	}
	isDir := fileInfo.IsDir()
	if decideDir {
		return isDir
	}
	return !isDir
}

func startProcess(execFileName string, args []string, envs []string, stdIn bool, attach bool, cwd string, extEnvs ...string) (ret *exec.Cmd, err error) {

	cmd := exec.Command(execFileName, args...)

	if stdIn {
		cmd.Stdin = os.Stdin
	}

	if attach {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	cmd.Dir = cwd

	newEnvsMap := map[string]string{}

	if envs != nil {
		for _, env := range envs {
			kv := strings.Split(env, "=")
			newEnvsMap[kv[0]] = kv[1]
		}
	}

	if extEnvs != nil {
		for _, env := range extEnvs {
			kv := strings.Split(env, "=")
			newEnvsMap[kv[0]] = kv[1]
		}
	}

	for k, v := range newEnvsMap {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	if err = cmd.Start(); err != nil {
		return
	}

	ret = cmd
	return
}

func killProcess(pid int) (err error) {
	err = syscall.Kill(pid, syscall.SIGKILL)
	return
}

func stopProcess(pid int) (err error) {
	err = syscall.Kill(pid, syscall.SIGTERM)
	return
}

func pauseProcess(pid int) (err error) {
	err = syscall.Kill(pid, syscall.SIGUSR1)
	return
}

func getUUID() string {
	id, _ := uuid.NewV4()
	return id.String()
}

func copyFile(dstName, srcName string, perm os.FileMode) (written int64, err error) {
	src, err := os.Open(srcName)
	if err != nil {
		return
	}

	defer src.Close()
	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, perm)
	if err != nil {
		return
	}

	defer dst.Close()
	return io.Copy(dst, src)
}
