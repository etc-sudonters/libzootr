package cmdlib

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sudonters/libzootr/internal/buffers"
	"sudonters/libzootr/internal/files"
	"time"

	"github.com/etc-sudonters/substrate/slipup"
)

type LoggingConfig struct {
	MaxBytes    uint64
	PathPattern string
	LogLevel    slog.Level
	LogFormat   string
	Quiet       bool
}

func (this *LoggingConfig) AddFlags(flags *flag.FlagSet) {
	flags.StringVar(&this.PathPattern, "log-file", "", "Log file name. If log-file-max-bytes is set, then this may be a fmt.Sprintf pattern passed a single string")
	flags.Uint64Var(&this.MaxBytes, "log-file-max-bytes", 0, "Max file size for log file")

	flags.Func("log-level", "Sets log level, see slog.LogLevel for options", func(v string) error {
		if err := this.LogLevel.UnmarshalText([]byte(v)); err != nil {
			return slipup.Createf("invalid log level %q", v)
		}
		return nil
	})

	flags.Func("log-format", "Either json or kv", func(v string) error {
		if strings.ToLower(v) == "json" {
			this.LogFormat = "json"
		} else {
			this.LogFormat = "kv"
		}
		return nil
	})
	flags.BoolVar(&this.Quiet, "very-very-quiet", false, "Disables logging entirely, highest priority")
}

func (this *LoggingConfig) CreateLogger() (*slog.Logger, error) {
	if this.Quiet || this.PathPattern == "" || this.PathPattern == "/dev/null" {
		return slog.New(&DropHandler{}), nil
	}

	var sink io.Writer
	var sinkErr error
	if this.MaxBytes > 0 {
		sink, sinkErr = newLogFileRotater(files.OsFS, this)
	} else {
		sink, sinkErr = files.OsFS.OpenFile(this.PathPattern, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	}

	if sinkErr != nil {
		return nil, sinkErr
	}

	handlerOptions := &slog.HandlerOptions{
		Level: this.LogLevel,
	}

	var handler slog.Handler
	if this.LogFormat == "json" {
		handler = slog.NewJSONHandler(sink, handlerOptions)
	} else {
		handler = slog.NewTextHandler(sink, handlerOptions)
	}

	return slog.New(handler), nil
}

func newLogFileRotater(fs files.OpenFS, opts *LoggingConfig) (*buffers.Rotater, error) {
	pattern := opts.PathPattern

	namer := func() string {
		now := time.Now()
		return fmt.Sprintf(pattern, now.Format("2006-01-02T15:04:05-07:00:00"))
	}

	if !strings.Contains(pattern, "%") {
		namer = func() string { return pattern }
	}

	return buffers.NewRotater(buffers.NewFileSystem(fs, namer), opts.MaxBytes)
}

var _ slog.Handler = (*DropHandler)(nil)

type DropHandler struct{}

func (d *DropHandler) Enabled(context.Context, slog.Level) bool {
	return false
}

func (d *DropHandler) Handle(context.Context, slog.Record) error {
	return nil
}

func (d *DropHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return d
}

func (d *DropHandler) WithGroup(name string) slog.Handler {
	return d
}
