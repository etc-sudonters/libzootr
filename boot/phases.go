package boot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"sudonters/libzootr/magicbean"
	"sudonters/libzootr/magicbean/tracking"
	"sudonters/libzootr/mido"
	"sudonters/libzootr/mido/ast"
	"sudonters/libzootr/mido/objects"
	"sudonters/libzootr/mido/optimizer"
	"sudonters/libzootr/settings"
	"sudonters/libzootr/zecs"
)

func Phase1_InitializeStorage(ddl []zecs.DDL) (zecs.Ocm, error) {
	ocm, err := zecs.New()
	if err != nil {
		return ocm, fmt.Errorf("failed to initialize ocm: %w", err)
	}

	if staticDdlErr := zecs.Apply(&ocm, staticddl()); staticDdlErr != nil {
		return ocm, fmt.Errorf("failed to apply static DDL: %w", staticDdlErr)
	}

	if len(ddl) > 0 {
		if appDdlErr := zecs.Apply(&ocm, ddl); appDdlErr != nil {
			return ocm, fmt.Errorf("failed to apply app DDL: %w", appDdlErr)
		}
	}
	return ocm, nil
}

func Phase2_ImportFromFiles(ctx context.Context, settings *settings.Model, ocm *zecs.Ocm, set *tracking.Set, paths LoadPaths) error {
	if err := storeScripts(ocm, paths); err != nil {
		return fmt.Errorf("failed while storing scripts: %w", err)
	}
	if err := storeTokens(set.Tokens, paths); err != nil {
		return fmt.Errorf("failed while storing tokens: %w", err)
	}
	if err := storePlacements(set.Nodes, set.Tokens, paths); err != nil {
		return fmt.Errorf("failed while storing placements: %w", err)
	}
	if err := storeRelations(set.Nodes, set.Tokens, paths); err != nil {
		return fmt.Errorf("failed while storing relations: %w", err)
	}

	if paths.Spoiler != "" {
		fh, err := os.Open(paths.Spoiler)
		if err != nil {
			return fmt.Errorf("failed to open spoiler log file %q: %w", paths.Spoiler, err)
		}
		defer fh.Close()

		slog.DebugContext(ctx, "loading spoiler log", "path", paths.Spoiler)
		if err := (LoadSpoilerData(ctx, fh, settings, &set.Nodes, &set.Tokens)); err != nil {
			return fmt.Errorf("failed while loading spoiler log file %q: %w", paths.Spoiler, err)
		}
	}
	return nil
}

func paniciferr(err error) {
	if err != nil {
		panic(err)
	}
}

func Phase3_ConfigureCompiler(ocm *zecs.Ocm, theseSettings *settings.Model, options ...mido.ConfigureCompiler) (env mido.CompileEnv, err error) {
	defer func() {
		if r := recover(); r != nil {
			if cause, ok := r.(error); ok {
				err = cause
			} else if str, ok := r.(string); ok {
				err = errors.New(str)
			} else {
				err = errors.New("panic!")
			}
		}
	}()
	defaults := []mido.ConfigureCompiler{
		mido.CompilerDefaults(),
		mido.CompilerWithGenerationSettings(settings.Names()),
		func(env *mido.CompileEnv) {
			env.Optimize.AddOptimizer(func(env *mido.CompileEnv) ast.Rewriter {
				reader := settings.Reader{Model: theseSettings}
				return optimizer.InlineSettings(reader, env.Symbols)
			})
			paniciferr(loadsymbols(ocm, env.Symbols))
			paniciferr(loadscripts(ocm, env))
			paniciferr(aliassymbols(ocm, env.Symbols))
		},
		installCompilerFunctions(theseSettings),
		installConnectionGenerator(ocm),
		mido.WithBuiltInFunctionDefs(func(*mido.CompileEnv) []objects.BuiltInFunctionDef {
			return magicbean.CreateBuiltInDefs()
		}),
		func(env *mido.CompileEnv) {
			createptrs(ocm, env.Symbols, env.Objects)
		},
	}
	defaults = slices.Concat(defaults, options)
	env = mido.NewCompileEnv(defaults...)
	return
}

func Phase4_Compile(ocm *zecs.Ocm, compiler *mido.CodeGen) error {
	if err := parseall(ocm, compiler); err != nil {
		return fmt.Errorf("failed while parsing: %w", err)
	}
	if err := optimizeall(ocm, compiler); err != nil {
		return fmt.Errorf("failed while optimizing: %w", err)
	}
	if err := compileall(ocm, compiler); err != nil {
		return fmt.Errorf("failed while compiling: %w", err)
	}
	return nil
}

func Phase5_CreateWorld(
	ocm *zecs.Ocm,
	settings *settings.Model,
	objects objects.Table,
	trackset *tracking.Set,
) (magicbean.ExplorableWorld, error) {
	magicbean.PlaceAlwaysItems(trackset)
	if err := magicbean.PromoteRemainingDefaultTokens(ocm); err != nil {
		return magicbean.ExplorableWorld{}, err
	}
	return explorableworldfrom(ocm)
}
