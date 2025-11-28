package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sudonters/libzootr/components"
	"sudonters/libzootr/internal"
	"sudonters/libzootr/magicbean/tracking"
	"sudonters/libzootr/settings"
	"sudonters/libzootr/zecs"
	"text/tabwriter"

	"github.com/etc-sudonters/substrate/slipup"

	"github.com/etc-sudonters/substrate/dontio"
	"github.com/etc-sudonters/substrate/stageleft"

	"runtime/pprof"
	"sudonters/libzootr/boot"
	"sudonters/libzootr/magicbean"
)

func runMain(ctx context.Context, std *dontio.Std, opts *cliOptions) stageleft.ExitCode {
	stopProfiling := profileto(opts.profile)
	defer stopProfiling()

	paths := boot.LoadPaths{
		Tokens:     filepath.Join(opts.dataDir, "items.json"),
		Placements: filepath.Join(opts.dataDir, "locations.json"),
		Scripts:    filepath.Join(opts.logicDir, "..", "helpers.json"),
		Relations:  opts.logicDir,
		Spoiler:    opts.spoiler,
	}

	theseSettings := settings.Default()
	FinalizeSettings(&theseSettings)

	generation, err := boot.Default(ctx, paths, &theseSettings)
	if err != nil {
		fmt.Fprintf(std.Err, "failed to boot zootr engine: %s", err)
		return 5
	}
	magicbean.CollectStartingItems(&generation)

	tokenNames, err := tracking.NameTableFrom(&generation.Ocm, zecs.With[components.TokenMarker])
	internal.PanicOnError(err)

	internal.PanicOnError(PrintInventory(std.Out, generation.Inventory, tokenNames))

	adult := magicbean.Search{
		Pending:    generation.World.Graph.Roots(),
		Generation: &generation,
		Age:        magicbean.AgeAdult,
	}
	child := magicbean.Search{
		Pending:    generation.World.Graph.Roots(),
		Generation: &generation,
		Age:        magicbean.AgeChild,
	}
	var i int
	for {
		i++
		std.WriteLineOut("")
		std.WriteLineOut("Sphere %d", i)
		adultResult := adult.Visit()
		childResult := child.Visit()
		reached := adultResult.Reached.Union(childResult.Reached)
		if reached.Len() == 0 {
			std.WriteLineOut("did not reach more nodes in either age")
			break
		}
		precollect := magicbean.CopyInventory(generation.Inventory)
		internal.PanicOnError(magicbean.CollectTokensFrom(
			&generation.Ocm,
			reached,
			generation.Inventory,
		))

		std.WriteLineOut("Inventory after collection")
		internal.PanicOnError(PrintInventory(std.Out, magicbean.DiffInventories(precollect, generation.Inventory), tokenNames))

		tabber := tabwriter.NewWriter(std.Out, 8, 2, 1, ' ', 0)
		fmt.Fprintf(tabber, "\tAdult\tChild\n")
		fmt.Fprintf(tabber, "Visited\t%3d\t%3d\n", adult.Visited.Len(), child.Visited.Len())
		fmt.Fprintf(tabber, "Reached\t%3d\t%3d\n", adultResult.Reached.Len(), childResult.Reached.Len())
		fmt.Fprintf(tabber, "Pending\t%3d\t%3d\n", adult.Pending.Len(), child.Pending.Len())
		tabber.Flush()
	}
	return stageleft.ExitCode(0)
}

func FinalizeSettings(these *settings.Model) {
	these.Seed = 0x76E76E14E9691280
	these.Logic.Shuffling.Flags = settings.ShuffleOcarinaNotes
	these.Logic.Spawns.StartAge = settings.StartAgeChild
	these.Logic.Connections.Flags |= settings.ConnectionOpenDoorOfTime
}

func profileto(path string) func() {
	if path == "" {
		return func() {}
	}

	f, err := os.Create(path)
	slipup.NeedsErrorHandling(err)
	pprof.StartCPUProfile(f)
	return func() { pprof.StopCPUProfile() }
}

func PrintInventory(w io.Writer, inventory magicbean.Inventory, names tracking.NameTable) error {
	tabber := tabwriter.NewWriter(w, 8, 8, 2, ' ', 0)
	if _, err := fmt.Fprintln(tabber, "Inventory:"); err != nil {
		return err
	}
	for ent, qty := range inventory {
		name, hasName := names[ent]
		if !hasName {
			continue
		}

		if _, err := fmt.Fprintf(tabber, "%s\t%d\n", name, qty); err != nil {
			return err
		}
	}
	return tabber.Flush()
}
