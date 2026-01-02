package buffers

import (
	"bytes"
	"errors"
	"io"
	"os"
	"sudonters/libzootr/internal/files"
	"sync"

	"github.com/etc-sudonters/substrate/slipup"
)

type memorybuffer struct {
	*bytes.Buffer
}

func (_ memorybuffer) Close() error {
	return nil
}

type Memory struct{}

func (_ *Memory) CreateBuffer(sizehint uint64) (io.WriteCloser, error) {
	return memorybuffer{bytes.NewBuffer(make([]byte, 0, sizehint))}, nil
}

type FileSystem struct {
	naming func() string
	fs     files.OpenFS
}

func (this *FileSystem) CreateBuffer(uint64) (io.WriteCloser, error) {
	return this.fs.OpenFile(this.naming(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
}

func NewFileSystem(fs files.OpenFS, naming func() string) *FileSystem {
	fsbufs := new(FileSystem)
	fsbufs.naming = naming
	fsbufs.fs = fs
	return fsbufs
}

type BufferCreator interface {
	CreateBuffer(sizehint uint64) (io.WriteCloser, error)
}

func NewRotater(creator BufferCreator, maxBytes uint64) (*Rotater, error) {
	rotater := new(Rotater)
	rotater.m = new(sync.Mutex)
	rotater.maxbytes = maxBytes
	rotater.newbuffer = creator

	if err := rotater.rotate(); err != nil {
		return rotater, slipup.Describe(err, "failed to initialize buffer rotation")
	}

	return rotater, nil
}

type Rotater struct {
	newbuffer BufferCreator
	maxbytes  uint64
	curr      *rotatingbuffer
	m         *sync.Mutex
}

func (this *Rotater) Rotate() error {
	this.m.Lock()
	defer this.m.Unlock()
	return this.rotate()
}

func (this *Rotater) Write(b []byte) (int, error) {
	this.m.Lock()
	defer this.m.Unlock()
	return this.write(b)
}

func (this *Rotater) write(b []byte) (int, error) {
	if this.curr == nil || this.shouldRotate(b) {
		if rotateErr := this.rotate(); rotateErr != nil {
			return 0, slipup.Describe(rotateErr, "failed to rotate buffer")
		}
	}

	return this.curr.Write(b)
}

func (this *Rotater) shouldRotate(b []byte) bool {
	return this.curr.n+uint64(len(b)) >= this.maxbytes
}

func (this *Rotater) rotate() error {
	if this.curr != nil {
		this.curr.Close()
		this.curr = nil
	}

	buffer, bufferErr := this.newbuffer.CreateBuffer(this.maxbytes)
	if bufferErr != nil {
		return slipup.Describe(bufferErr, "failed to initialize buffer")
	}

	this.curr = &rotatingbuffer{buffer, 0, nil, false}
	return nil
}

type rotatingbuffer struct {
	w      io.WriteCloser
	n      uint64
	err    error
	closed bool
}

var errBufferClosed = errors.New("buffer already closed")

func (this *rotatingbuffer) Write(b []byte) (int, error) {
	if this.closed {
		return 0, errBufferClosed
	}
	if this.err != nil {
		return 0, this.err
	}
	n, err := this.w.Write(b)
	this.n += uint64(n)
	this.err = err
	return n, err
}

func (this *rotatingbuffer) Close() {
	if !this.closed {
		this.closed = true
		this.w.Close()
	}
}
