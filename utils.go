package spirit

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

const (
	SPIRIT = "spirit"
)

func GetComponentHome(spiritName string) string {
	tmpDir := strings.TrimRight(os.TempDir(), "/")
	return fmt.Sprintf("%s/%s/%s", tmpDir, SPIRIT, spiritName)
}

func MakeComponentHome(spiritName string) (dir string, err error) {
	err = os.MkdirAll(GetComponentHome(spiritName), 0700)
	return
}

func IsProcessAlive(pid int) bool {
	p, _ := os.FindProcess(pid)
	if e := p.Signal(syscall.Signal(0)); e == nil {
		return true
	}
	return false
}

func IsFileOrDir(filename string, decideDir bool) bool {
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

func StartProcess(execFileName string, args []string, envs []string, std bool, extEnvs ...string) (pid int, err error) {

	cmd := exec.Command(execFileName, args...)

	if std {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

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

	}

	if err = cmd.Start(); err == nil {
		pid = cmd.Process.Pid
	}

	return
}

func KillProcess(pid int) (err error) {
	err = syscall.Kill(pid, syscall.SIGKILL)
	return
}
