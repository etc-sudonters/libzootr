package files

import (
	"errors"
	"io/fs"
	"os"
)

type WriteableFile interface {
	fs.File
	Write([]byte) (int, error)
	Close() error
}

type OpenFS interface {
	fs.FS
	OpenFile(name string, flag int, perm fs.FileMode) (WriteableFile, error)
}

var memFileClosedErr = errors.New("...")
var OsFS OpenFS = &osfs{}

type osfs struct{}

func (this *osfs) Open(name string) (fs.File, error) {
	return this.OpenFile(name, os.O_RDONLY, 0)
}

func (this *osfs) OpenFile(name string, flag int, perm fs.FileMode) (WriteableFile, error) {
	return os.OpenFile(name, flag, perm)
}
