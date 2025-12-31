package bootstrap

import (
	"context"
	"io/fs"
	"slices"
	"sudonters/libzootr/internal/settings"
	"sudonters/libzootr/magicbean"
	"sudonters/libzootr/magicbean/tracking"
	"sudonters/libzootr/mido"
	"sudonters/libzootr/mido/ast"
	"sudonters/libzootr/mido/objects"
	"sudonters/libzootr/mido/optimizer"
	"sudonters/libzootr/zecs"
)

func PanicWhenErr(err error) {
	if err != nil {
		panic(err)
	}
}

func Phase1_InitializeStorage(ddl []zecs.DDL) zecs.Ocm {
	ocm, err := zecs.New()
	PanicWhenErr(err)
	PanicWhenErr(zecs.Apply(&ocm, staticddl()))
	return ocm
}

func Phase2_ImportFromFiles(ctx context.Context, fs fs.FS, ocm *zecs.Ocm, set *tracking.Set, paths LoadPaths) error {
	PanicWhenErr(storeScripts(ctx, fs, ocm, paths))
	PanicWhenErr(storeTokens(ctx, fs, set.Tokens, paths))
	PanicWhenErr(storePlacements(ctx, fs, set.Nodes, set.Tokens, paths))
	PanicWhenErr(storeRelations(ctx, fs, set.Nodes, set.Tokens, paths))
	return nil
}

func Phase3_ConfigureCompiler(ocm *zecs.Ocm, theseSettings *settings.Zootr, options ...mido.ConfigureCompiler) mido.CompileEnv {
	defaults := []mido.ConfigureCompiler{
		mido.CompilerDefaults(),
		func(env *mido.CompileEnv) {
			env.Optimize.AddOptimizer(func(env *mido.CompileEnv) ast.Rewriter {
				return optimizer.InlineSettings(theseSettings, env.Symbols)
			})
			PanicWhenErr(loadsymbols(ocm, env.Symbols))
			PanicWhenErr(loadscripts(ocm, env))
			PanicWhenErr(aliassymbols(ocm, env.Symbols))
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
	return mido.NewCompileEnv(defaults...)
}

func Phase4_Compile(ocm *zecs.Ocm, compiler *mido.CodeGen) error {
	PanicWhenErr(parseall(ocm, compiler))
	PanicWhenErr(optimizeall(ocm, compiler))
	PanicWhenErr(compileall(ocm, compiler))
	return nil
}

func Phase5_CreateWorld(ocm *zecs.Ocm, settings *settings.Zootr, objects objects.Table) magicbean.ExplorableWorld {
	xplore := explorableworldfrom(ocm)
	return xplore
}
