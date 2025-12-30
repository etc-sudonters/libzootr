package main

import (
	"context"
	"errors"
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
	args := os.Args[1:]
	flags := flag.CommandLine
	if argsErr := (&opts).init(args, flags); argsErr != nil {
		appExitCode = 2
		realStd.WriteLineErr(argsErr.Error())
		return
	}
	appStd := dontio.Std{}
	cleanup, err := redirectAppStd(&appStd, &opts)
	defer cleanup()
	if err != nil {
		realStd.WriteLineErr("Failed to redirect application std{in,out,err}\n%v", err)
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

func noop() {}

func redirectAppStd(std *dontio.Std, opts *cliOptions) (func(), error) {
	if !filepath.IsAbs(opts.logDir) {
		path, pathErr := filepath.Abs(opts.logDir)
		if pathErr != nil {
			return noop, slipup.Describef(pathErr, "failed to initialize log dir %q", path)
		}
		opts.logDir = path
	}
	logDirErr := os.Mkdir(opts.logDir, 0777)
	if logDirErr != nil && !os.IsExist(logDirErr) {
		return noop, slipup.Describef(logDirErr, "failed to initialize log dir %q", opts.logDir)
	}

	std.In = dontio.AlwaysErrReader{io.ErrUnexpectedEOF}
	return dontio.FileStd(std, opts.logDir)
}

type missingRequired string // option name

func (arg missingRequired) Error() string {
	clivar, ok := clivars[string(arg)]
	if !ok {
		panic(slipup.Createf("unknown cliflag %s", string(arg)))
	}
	description := clivar.description
	return fmt.Sprintf("-%s is required: %s", string(arg), description)
}

type cliOptions struct {
	logDir   string
	worldDir string
	dataDir  string
	spoiler  string
	seed     uint64
}

func (opts *cliOptions) init(cliargs []string, flags *flag.FlagSet) error {
	for _, clivar := range clivars {
		clivar.add(opts, flags)
	}

	err := flags.Parse(cliargs)

	if opts.worldDir == "" {
		err = errors.Join(err, missingRequired("w"))
	}

	if opts.dataDir == "" {
		err = errors.Join(err, missingRequired("d"))
	}

	if opts.spoiler == "" {
		err = errors.Join(err, missingRequired("s"))
	}

	return err
}

type clivar struct {
	name, description string
	defaultValue      any
	ptr               func(*cliOptions) any
}

func (this *clivar) add(opts *cliOptions, flags *flag.FlagSet) {
	switch defaultValue := this.defaultValue.(type) {
	case string:
		ptr := this.ptr(opts).(*string)
		flags.StringVar(ptr, this.name, defaultValue, this.description)
		break
	case bool:
		ptr := this.ptr(opts).(*bool)
		flags.BoolVar(ptr, this.name, defaultValue, this.description)
		break
	case uint64:
		ptr := this.ptr(opts).(*uint64)
		flags.Uint64Var(ptr, this.name, defaultValue, this.description)
	default:
		panic(slipup.Createf("unknown cli type: %t", defaultValue))
	}
}

const defaultSeed uint64 = 0x76E76E14E9691280

var clivars = map[string]clivar{
	"data":    {"data", "Directory where data files are stored", "", func(opts *cliOptions) any { return &opts.dataDir }},
	"logdir":  {"logdir", "Directory open log files in", ".logs", func(opts *cliOptions) any { return &opts.logDir }},
	"spoiler": {"spoiler", "Path to spoiler log to import", "", func(opts *cliOptions) any { return &opts.spoiler }},
	"world":   {"world", "Directory where logic files are located", "", func(opts *cliOptions) any { return &opts.worldDir }},
	"seed":    {"seed", "Seed for generation", defaultSeed, func(opts *cliOptions) any { return &opts.seed }},
}
