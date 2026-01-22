package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"runtime/pprof"
	"sudonters/libzootr/cmd/cmdlib"

	"github.com/etc-sudonters/substrate/dontio"
	"github.com/etc-sudonters/substrate/files"
	"github.com/etc-sudonters/substrate/stageleft"
)

func main() {
	ctx, cancelApp := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	_ = cancelApp
	appstd := dontio.Std{
		In:  os.Stdin,
		Out: os.Stdout,
		Err: os.Stderr,
	}
	ctx = dontio.AddStdToContext(ctx, &appstd)
	appExitCode := stageleft.ExitSuccess
	opts := cliOptions{
		logging: &cmdlib.LoggingConfig{},
	}

	defer func() {
		os.Exit(int(appExitCode))
	}()
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				appstd.WriteLineErr("%s", err)
			} else if str, ok := r.(string); ok {
				appstd.WriteLineErr("%s", str)
			}
			_, _ = appstd.Err.Write(debug.Stack())
			if appExitCode != stageleft.ExitSuccess {
				appExitCode = stageleft.AsExitCode(r, stageleft.ExitCode(126))
			}
		}
	}()

	flags := flag.CommandLine
	args := os.Args[1:]
	fs := files.OsFS

	if argsErr := (&opts).init(flags, args); argsErr != nil {
		appExitCode = 2
		appstd.WriteLineErr(argsErr.Error())
		return
	}

	logger, loggerErr := opts.logging.CreateLogger(&appstd)
	if loggerErr != nil {
		appExitCode = 2
		appstd.WriteLineErr("failed to configure logging: %s", loggerErr)
		return
	}
	slog.SetDefault(logger)

	stopProfiling := profileto(opts.profile)
	defer stopProfiling()
	appExitCode = runMain(ctx, appstd, opts, fs)
}

type missingRequired string // option name

func (arg missingRequired) Error() string {
	return fmt.Sprintf("-%s is required", string(arg))
}

type cliOptions struct {
	logicDir  string
	dataDir   string
	includeMq bool
	profile   string
	logging   *cmdlib.LoggingConfig
}

func (opts *cliOptions) init(flags *flag.FlagSet, args []string) error {
	flags.StringVar(&opts.logicDir, "l", "", "Directory where logic files are located")
	flags.StringVar(&opts.dataDir, "d", "", "Directory where data files are stored")
	flags.StringVar(&opts.profile, "p", "", "profile file name")
	flags.BoolVar(&opts.includeMq, "M", false, "Whether or not to include MQ data")
	opts.logging.AddFlags(flags)

	flagErr := flags.Parse(args)
	if flagErr != nil {
		return flagErr
	}

	if opts.logicDir == "" {
		flagErr = errors.Join(flagErr, missingRequired("l"))
	}

	if opts.dataDir == "" {
		flagErr = errors.Join(flagErr, missingRequired("d"))
	}

	return flagErr
}

func noop() {}

func profileto(path string) func() {
	if path == "" {
		return noop
	}
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	pprof.StartCPUProfile(f)
	return func() { pprof.StopCPUProfile() }
}
