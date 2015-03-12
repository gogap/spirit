package spirit

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"syscall"
)

type LockFileContent struct {
	PID     int                    `json:"pid"`
	Context map[string]interface{} `json:"context"`
}

type LockFile struct {
	*os.File
}

func NewLockFile(file *os.File) *LockFile {
	return &LockFile{file}
}

func CreateLockFile(name string, perm os.FileMode) (lockfile *LockFile, err error) {
	if lockfile, err = OpenLockFile(name, perm); err != nil {
		return
	}
	if err = lockfile.Lock(); err != nil {
		lockfile.Remove()
		return
	}
	if err = lockfile.WriteContent(nil); err != nil {
		lockfile.Remove()
	}
	return
}

func OpenLockFile(name string, perm os.FileMode) (lockfile *LockFile, err error) {
	var file *os.File
	if file, err = os.OpenFile(name, os.O_RDWR|os.O_CREATE, perm); err == nil {
		lockfile = &LockFile{file}
	}
	return
}

func (p *LockFile) Lock() error {
	return syscall.Flock(int(p.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
}

func (p *LockFile) Unlock() error {
	return syscall.Flock(int(p.Fd()), syscall.LOCK_UN)
}

func Read(name string) (content LockFileContent, err error) {
	var file *os.File
	if file, err = os.OpenFile(name, os.O_RDONLY, 0640); err != nil {
		return
	}
	defer file.Close()

	lockfile := &LockFile{file}
	content, err = lockfile.ReadContent()
	return
}

func (p *LockFile) ReadContent() (content LockFileContent, err error) {
	if _, err = p.Seek(0, os.SEEK_SET); err != nil {
		return
	}

	if data, e := ioutil.ReadAll(p); e != nil {
		err = e
		return
	} else if e := json.Unmarshal(data, &content); e != nil {
		err = e
		return
	}

	return
}

func (p *LockFile) WriteContent(context map[string]interface{}) (err error) {
	if _, err = p.Seek(0, os.SEEK_SET); err != nil {
		return
	}

	fileContent := LockFileContent{
		PID:     os.Getpid(),
		Context: context,
	}

	var contentData []byte
	if contentData, err = json.Marshal(fileContent); err != nil {
		return
	}

	if err = p.Truncate(int64(len(contentData) * 2)); err != nil {
		return
	}

	if _, err = p.WriteString(string(contentData)); err != nil {
		return
	}

	err = p.Sync()
	return
}

func (p *LockFile) Remove() (err error) {
	defer p.Close()

	if err = p.Unlock(); err != nil {
		return err
	}

	name := ""
	name, err = GetFdName(p.Fd())
	if err != nil {
		return err
	}

	err = syscall.Unlink(name)
	return err
}

func GetFdName(fd uintptr) (name string, err error) {
	path := fmt.Sprintf("/proc/self/fd/%d", int(fd))

	var (
		fi os.FileInfo
		n  int
	)
	if fi, err = os.Lstat(path); err != nil {
		return
	}

	buf := make([]byte, fi.Size()+1)

	if n, err = syscall.Readlink(path, buf); err == nil {
		name = string(buf[:n])
	}
	return
}
