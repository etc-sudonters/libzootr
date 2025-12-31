package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"runtime/debug"

	"github.com/etc-sudonters/substrate/dontio"
	"github.com/etc-sudonters/substrate/stageleft"
)

func main() {
	var opts cliOptions
	var appExitCode stageleft.ExitCode = stageleft.ExitSuccess
	std := dontio.Std{
		In:  os.Stdin,
		Out: os.Stdout,
		Err: os.Stderr,
	}
	defer func() {
		os.Exit(int(appExitCode))
	}()
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				std.WriteLineErr("%s", err)
			} else if str, ok := r.(string); ok {
				std.WriteLineErr("%s", str)
			}
			_, _ = std.Err.Write(debug.Stack())
			if appExitCode != stageleft.ExitSuccess {
				appExitCode = stageleft.AsExitCode(r, stageleft.ExitCode(126))
			}
		}
	}()

	ctx := context.Background()
	ctx = dontio.AddStdToContext(ctx, &std)

	if argsErr := (&opts).init(); argsErr != nil {
		appExitCode = 2
		std.WriteLineErr(argsErr.Error())
		return
	}

	appExitCode = runMain(ctx, opts, &bfs)
	return
}

type missingRequired string // option name

func (arg missingRequired) Error() string {
	return fmt.Sprintf("%s is required", string(arg))
}

type cliOptions struct {
	logicDir  string
	dataDir   string
	includeMq bool
	profile   string
}

func (opts *cliOptions) init() error {
	flag.StringVar(&opts.logicDir, "l", "", "Directory where logic files are located")
	flag.StringVar(&opts.dataDir, "d", "", "Directory where data files are stored")
	flag.StringVar(&opts.profile, "p", "", "profile file name")
	flag.BoolVar(&opts.includeMq, "M", false, "Whether or not to include MQ data")
	flag.Parse()

	if opts.logicDir == "" {
		return missingRequired("-l")
	}

	if opts.dataDir == "" {
		return missingRequired("-d")
	}

	return nil
}

var bfs = BasicFS{}

type BasicFS struct{}

func (_ *BasicFS) Open(name string) (fs.File, error) {
	return os.Open(name)
}
