package buffers

import (
	"io"
	"log/slog"
	"strings"
	"testing"
)

type counter struct {
	inner BufferCreator
	n     int
}

func (this *counter) CreateBuffer(hint uint64) (io.WriteCloser, error) {
	this.n += 1
	return this.inner.CreateBuffer(hint)
}

func TestBufferRotation(t *testing.T) {
	buffers := &counter{&Memory{}, 0}
	rotation, err := NewRotater(buffers, 64)
	if err != nil {
		t.Fatal(err)
	}
	handler := slog.NewTextHandler(rotation, nil)
	logger := slog.New(handler)
	logger.Warn(strings.Repeat("a", 34))
	logger.Warn(strings.Repeat("a", 34))

	if buffers.n <= 1 {
		t.Fatalf("expected to create at least 2 buffers but only created %v", buffers.n)
	}
}
