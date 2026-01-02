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
	"sudonters/libzootr/table"
	"sudonters/libzootr/table/ocm"
)

func PanicWhenErr(err error) {
	if err != nil {
		panic(err)
	}
}

func Phase1_InitializeStorage(ddl []table.DDL) (*table.Table, *ocm.Entities) {
	ddl = slices.Concat(ddl, staticddl(), ocm.DDL())
	tbl, tblErr := table.FromDDL(ddl...)
	PanicWhenErr(tblErr)
	entities := ocm.NewEntities(tbl)
	return tbl, entities
}

func Phase2_ImportFromFiles(ctx context.Context, fs fs.FS, entities *ocm.Entities, set *tracking.Set, paths LoadPaths) error {
	PanicWhenErr(storeScripts(ctx, fs, entities, paths))
	PanicWhenErr(storeTokens(ctx, fs, set.Tokens, paths))
	PanicWhenErr(storePlacements(ctx, fs, set.Nodes, set.Tokens, paths))
	PanicWhenErr(storeRelations(ctx, fs, set.Nodes, set.Tokens, paths))
	return nil
}

func Phase3_ConfigureCompiler(entities *ocm.Entities, theseSettings *settings.Zootr, options ...mido.ConfigureCompiler) mido.CompileEnv {
	defaults := []mido.ConfigureCompiler{
		mido.CompilerDefaults(),
		func(env *mido.CompileEnv) {
			env.Optimize.AddOptimizer(func(env *mido.CompileEnv) ast.Rewriter {
				return optimizer.InlineSettings(theseSettings, env.Symbols)
			})
			PanicWhenErr(loadsymbols(entities, env.Symbols))
			PanicWhenErr(loadscripts(entities, env))
			PanicWhenErr(aliassymbols(entities, env.Symbols))
		},
		installCompilerFunctions(theseSettings),
		installConnectionGenerator(entities),
		mido.WithBuiltInFunctionDefs(func(*mido.CompileEnv) []objects.BuiltInFunctionDef {
			return magicbean.CreateBuiltInDefs()
		}),
		func(env *mido.CompileEnv) {
			createptrs(entities, env.Symbols, env.Objects)
		},
	}
	defaults = slices.Concat(defaults, options)
	return mido.NewCompileEnv(defaults...)
}

func Phase4_Compile(entities *ocm.Entities, compiler *mido.CodeGen) error {
	PanicWhenErr(parseall(entities, compiler))
	PanicWhenErr(optimizeall(entities, compiler))
	PanicWhenErr(compileall(entities, compiler))
	return nil
}

func Phase5_CreateWorld(entities *ocm.Entities, settings *settings.Zootr, objects objects.Table) magicbean.ExplorableWorld {
	xplore := explorableworldfrom(entities)
	return xplore
}
