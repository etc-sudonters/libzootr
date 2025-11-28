package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/etc-sudonters/substrate/dontio"
	"github.com/etc-sudonters/substrate/slipup"
	"github.com/etc-sudonters/substrate/stageleft"
)

func main() {
	var opts cliOptions
	var appExitCode stageleft.ExitCode = stageleft.ExitSuccess
	realStd := dontio.StdIo()
	defer func() {
		os.Exit(int(appExitCode))
	}()
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				realStd.WriteLineErr("%s", err)
			} else if str, ok := r.(string); ok {
				realStd.WriteLineErr("%s", str)
			}
			_, _ = realStd.Err.Write(debug.Stack())
			if appExitCode != stageleft.ExitSuccess {
				appExitCode = stageleft.AsExitCode(r, stageleft.ExitCode(126))
			}
		}
	}()

	ctx := context.Background()
	if argsErr := (&opts).init(); argsErr != nil {
		appExitCode = 2
		realStd.WriteLineErr(argsErr.Error())
		return
	}
	appStd := dontio.Std{}
	cleanup, err := redirectAppStd(&appStd, &opts)
	if cleanup != nil {
		defer cleanup()
	}
	if err != nil {
		fmt.Fprintf(realStd.Err, "Failed to redirect application std{in,out,err}\n%v\n", err)
		appExitCode = 3
		return
	}
	ctx = dontio.AddStdToContext(ctx, &appStd)

	if err := runMain(ctx, &appStd, &opts); err != nil {
		fmt.Fprintln(realStd.Err, err)
		appExitCode = stageleft.AsExitCode(err, 126)
	}
	return
}

func redirectAppStd(std *dontio.Std, opts *cliOptions) (func(), error) {
	if !filepath.IsAbs(opts.logDir) {
		path, pathErr := filepath.Abs(opts.logDir)
		if pathErr != nil {
			return nil, slipup.Describef(pathErr, "failed to initialize log dir %q", path)
		}
		opts.logDir = path
	}
	logDirErr := os.Mkdir(opts.logDir, 0777)
	if logDirErr != nil && !os.IsExist(logDirErr) {
		return nil, slipup.Describef(logDirErr, "failed to initialize log dir %q", opts.logDir)
	}

	std.In = dontio.AlwaysErrReader{io.ErrUnexpectedEOF}
	return dontio.FileStd(std, opts.logDir)
}

type missingRequired string // option name

func (arg missingRequired) Error() string {
	return fmt.Sprintf("%s is required", string(arg))
}

type cliOptions struct {
	logDir   string
	worldDir string
	dataDir  string
	spoiler  string
}

func (opts *cliOptions) init() error {
	flag.StringVar(&opts.logDir, "l", ".logs", "Directory open log files in")
	flag.StringVar(&opts.worldDir, "w", "", "Directory where logic files are located")
	flag.StringVar(&opts.dataDir, "d", "", "Directory where data files are stored")
	flag.StringVar(&opts.spoiler, "s", "", "Path to spoiler log to import")
	flag.Parse()
	if opts.worldDir == "" {
		return missingRequired("-w")
	}

	if opts.dataDir == "" {
		return missingRequired("-d")
	}

	if opts.spoiler == "" {
		return missingRequired("-s")
	}
	return nil
}
