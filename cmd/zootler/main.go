package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"
	"runtime/debug"
	"sudonters/zootler/internal/app"
	"sudonters/zootler/internal/settings"

	"github.com/etc-sudonters/substrate/dontio"
	"github.com/etc-sudonters/substrate/stageleft"
)

type missingRequired string // option name

func (arg missingRequired) Error() string {
	return fmt.Sprintf("%s is required", string(arg))
}

type cliOptions struct {
	logicDir  string
	dataDir   string
	includeMq bool
}

func (opts *cliOptions) init() error {
	flag.StringVar(&opts.logicDir, "l", "", "Directory where logic files are located")
	flag.StringVar(&opts.dataDir, "d", "", "Directory where data files are stored")
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
			}
			_, _ = std.Err.Write(debug.Stack())
			if appExitCode != stageleft.ExitSuccess {
				appExitCode = stageleft.AsExitCode(r, stageleft.ExitCode(126))
			}
		}
	}()

	exitWithErr := func(code stageleft.ExitCode, err error) {
		appExitCode = code
		std.WriteLineErr(err.Error())
	}

	ctx := context.Background()
	ctx = dontio.AddStdToContext(ctx, &std)

	if argsErr := (&opts).init(); argsErr != nil {
		exitWithErr(2, argsErr)
		return
	}

	z, appCreateErr := app.New(ctx,
		app.Setup(CreateScheme{DDL: MakeDDL()}),
		app.Setup(DataFileLoader[FileItem]{
			IncludeMQ: opts.includeMq,
			Path:      path.Join(opts.dataDir, "items.json"),
		}),
		app.Setup(DataFileLoader[FileLocation]{
			IncludeMQ: opts.includeMq,
			Path:      path.Join(opts.dataDir, "locations.json"),
			Add:       new(AttachDefaultItem),
		}),
		app.AddResource[settings.ZootrSettings](settings.Default()), // pretend we loaded it from somewhere
		app.Setup(WorldLoader{
			Path:    opts.logicDir,
			Helpers: path.Join(path.Dir(opts.logicDir), "..", "helpers.json"),
		}),
	)

	if appCreateErr != nil {
		exitWithErr(3, appCreateErr)
		return
	}

	if appCmdErr := z.Run(example); appCmdErr != nil {
		exitWithErr(4, appCmdErr)
	}
}
