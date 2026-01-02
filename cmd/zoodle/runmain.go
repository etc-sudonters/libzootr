package main

import (
	"context"
	"io/fs"
	"math/rand/v2"
	"path/filepath"
	"sudonters/libzootr/internal"
	"sudonters/libzootr/internal/settings"
	"sudonters/libzootr/magicbean/tracking"

	"github.com/etc-sudonters/substrate/dontio"
	"github.com/etc-sudonters/substrate/rng"
	"github.com/etc-sudonters/substrate/skelly/bitset32"
	"github.com/etc-sudonters/substrate/stageleft"

	"sudonters/libzootr/cmd/zoodle/bootstrap"
	"sudonters/libzootr/magicbean"
	"sudonters/libzootr/mido"
	"sudonters/libzootr/mido/objects"
)

func runMain(ctx context.Context, std dontio.Std, opts cliOptions, fs fs.FS) stageleft.ExitCode {
	paths := bootstrap.LoadPaths{
		Tokens:     filepath.Join(opts.dataDir, "items.json"),
		Placements: filepath.Join(opts.dataDir, "locations.json"),
		Scripts:    filepath.Join(opts.logicDir, "..", "helpers.json"),
		Relations:  opts.logicDir,
	}

	theseSettings := settings.Default()
	theseSettings.Seed = 0x76E76E14E9691280
	theseSettings.Shuffling.OcarinaNotes = true
	theseSettings.Spawns.StartingAge = settings.StartAgeAdult
	theseSettings.Locations.OpenDoorOfTime = true
	generation := setup(ctx, fs, paths, &theseSettings)
	generation.Settings = theseSettings
	CollectStartingItems(&generation)
	visited := bitset32.Bitset{}
	workset := generation.World.Graph.Roots()
	xplr := magicbean.Exploration{
		Visited: &visited,
		Workset: &workset,
	}
	results := explore(ctx, &xplr, &generation, AgeAdult)
	std.WriteLineOut("Visited %d", visited.Len())
	std.WriteLineOut("Reached %d", results.Reached.Len())
	std.WriteLineOut("Pending %d", results.Pending.Len())
	return stageleft.ExitCode(0)
}

func setup(ctx context.Context, fs fs.FS, paths bootstrap.LoadPaths, settings *settings.Zootr) (generation magicbean.Generation) {
	tbl, entities := bootstrap.Phase1_InitializeStorage(nil)
	_ = tbl
	trackSet, trackingErr := tracking.NewTrackingSet(entities)
	internal.PanicOnError(trackingErr)
	bootstrap.PanicWhenErr(bootstrap.Phase2_ImportFromFiles(ctx, fs, entities, &trackSet, paths))

	compileEnv := bootstrap.Phase3_ConfigureCompiler(entities, settings)

	codegen := mido.Compiler(&compileEnv)

	bootstrap.PanicWhenErr(bootstrap.Phase4_Compile(
		entities, &codegen,
	))

	world := bootstrap.Phase5_CreateWorld(entities, settings, objects.TableFrom(compileEnv.Objects))

	generation.Entities = entities
	generation.World = world
	generation.Objects = objects.TableFrom(compileEnv.Objects)
	generation.Inventory = magicbean.NewInventory()
	generation.Rng = *rand.New(rng.NewXoshiro256PPFromU64(settings.Seed))

	return generation
}
