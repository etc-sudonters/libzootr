package main

import (
	"context"
	"path/filepath"
	"sudonters/libzootr/boot"
	"sudonters/libzootr/cmd/knowitall/bubbles/generation"
	"sudonters/libzootr/cmd/knowitall/leaves"
	"sudonters/libzootr/components"
	"sudonters/libzootr/magicbean"
	"sudonters/libzootr/magicbean/tracking"
	"sudonters/libzootr/mido"
	"sudonters/libzootr/playthrough"
	"sudonters/libzootr/settings"
	"sudonters/libzootr/zecs"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/etc-sudonters/substrate/dontio"
	"github.com/etc-sudonters/substrate/slipup"
	"github.com/etc-sudonters/substrate/stageleft"
)

func createGeneration(ctx context.Context, std *dontio.Std, opts *cliOptions) (magicbean.Generation, error) {
	paths := boot.LoadPaths{
		Tokens:     filepath.Join(opts.dataDir, "items.json"),
		Placements: filepath.Join(opts.dataDir, "locations.json"),
		Scripts:    filepath.Join(opts.worldDir, "..", "helpers.json"),
		Relations:  opts.worldDir,
		Spoiler:    opts.spoiler,
	}
	these := settings.Default()
	these.Seed = 0x76E76E14E9691280
	these.Logic.Spawns.StartAge = settings.StartAgeChild
	these.Logic.Connections.Flags |= settings.ConnectionOpenDoorOfTime

	generation, err := boot.Default(ctx, paths, &these)
	if err != nil {
		return generation, slipup.Describe(err, "failed to boot zootr engine")
	}
	if err := magicbean.CollectStartingItems(&generation); err != nil {
		return generation, slipup.Describe(err, "while collecting starting items")
	}

	return generation, nil
}

func runMain(ctx context.Context, std *dontio.Std, opts *cliOptions) error {
	gen, genErr := createGeneration(ctx, std, opts)
	if genErr != nil {
		return stageleft.AttachExitCode(slipup.Describe(genErr, "failed to boot zootr engine"), 101)
	}

	vms := [2]mido.VM{
		magicbean.CreateVMForAge(&gen, magicbean.AgeAdult),
		magicbean.CreateVMForAge(&gen, magicbean.AgeChild),
	}
	searchers := [2]playthrough.Search{
		playthrough.SearchFromRoots(&vms[0], &gen.World),
		playthrough.SearchFromRoots(&vms[1], &gen.World),
	}

	searches := playthrough.Searches{
		Adult: &searchers[0],
		Child: &searchers[1],
	}

	nametable, nametableErr := tracking.NameTableFrom(&gen.Ocm, zecs.With[components.Name])
	if nametableErr != nil {
		return stageleft.AttachExitCode(slipup.Describe(nametableErr, "failed to create nametable"), 102)
	}

	mount := generation.New(&gen, nametable, searches)
	app := leaves.NewApp(ctx, std, mount)
	p := tea.NewProgram(app)
	final, err := p.Run()
	if err != nil {
		return stageleft.AttachExitCode(slipup.Describe(err, "application exited"), 5)
	}

	app = final.(leaves.App)
	if app.Err != nil {
		return stageleft.AttachExitCode(slipup.Describe(app.Err, "application exited due to err"), 99)
	}

	return nil
}
