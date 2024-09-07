package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"sudonters/zootler/carpenters"
	"sudonters/zootler/carpenters/ichiro"
	"sudonters/zootler/internal/app"

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
		app.Setup(&carpenters.Mutoh{
			Ichiro: ichiro.DataLoader{
				Table: ichiro.TableLoader{
					Scheme: ichiro.BaseDDL(),
				},
				DataPath: opts.dataDir,
			}}),
		app.AddResource(AstAllRuleEdges{
			AllEdgeRulesFrom: AllEdgeRulesFrom{Path: opts.logicDir},
		}),
		app.Setup(&IceArrowRuntime{}),
	)

	if appCreateErr != nil {
		exitWithErr(3, appCreateErr)
		return
	}

	/*
		if appCmdErr := z.Run(example); appCmdErr != nil {
			exitWithErr(4, appCmdErr)
		}
	*/

	runtime.KeepAlive(z)
}
