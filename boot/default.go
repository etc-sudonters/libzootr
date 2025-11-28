package boot

import (
	"context"
	"fmt"
	"math/rand/v2"
	"sudonters/libzootr/magicbean/tracking"
	"sudonters/libzootr/settings"

	"github.com/etc-sudonters/substrate/rng"

	"sudonters/libzootr/magicbean"
	"sudonters/libzootr/mido"
	"sudonters/libzootr/mido/objects"
)

func Default(ctx context.Context, paths LoadPaths, settings *settings.Model) (generation magicbean.Generation, setupErr error) {
	ocm, phase1Err := Phase1_InitializeStorage(nil)
	if phase1Err != nil {
		setupErr = fmt.Errorf("phase 1: %w", phase1Err)
		return
	}
	trackSet := tracking.NewTrackingSet(&ocm)

	if phase2Err := Phase2_ImportFromFiles(ctx, settings, &ocm, &trackSet, paths); phase2Err != nil {
		setupErr = fmt.Errorf("phase 2: %w", phase2Err)
		return
	}

	compileEnv, phase3Err := Phase3_ConfigureCompiler(&ocm, settings)
	if phase3Err != nil {
		setupErr = fmt.Errorf("phase 3: %w", phase3Err)
		return
	}

	codegen := mido.Compiler(&compileEnv)

	if phase4Err := (Phase4_Compile(&ocm, &codegen)); phase4Err != nil {
		setupErr = fmt.Errorf("phase 4: %w", phase4Err)
		return
	}

	world, phase5Err := Phase5_CreateWorld(&ocm, settings, objects.TableFrom(compileEnv.Objects), &trackSet)
	if phase5Err != nil {
		setupErr = fmt.Errorf("phase 5: %w", phase5Err)
		return
	}

	generation.Ocm = ocm
	generation.World = world
	generation.Objects = objects.TableFrom(compileEnv.Objects)
	generation.Inventory = magicbean.EmptyInventory()
	generation.Rng = *rand.New(rng.NewXoshiro256PPFromU64(settings.Seed))
	generation.Tokens = trackSet.Tokens
	generation.Nodes = trackSet.Nodes
	generation.Settings = settings
	generation.Symbols = compileEnv.Symbols
	generation.CodeGen = codegen

	return generation, nil
}
