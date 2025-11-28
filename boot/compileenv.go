package boot

import (
	"fmt"
	"regexp"
	"strings"
	"sudonters/libzootr/components"
	"sudonters/libzootr/magicbean/tracking"
	"sudonters/libzootr/mido"
	"sudonters/libzootr/mido/ast"
	"sudonters/libzootr/mido/objects"
	"sudonters/libzootr/mido/optimizer"
	"sudonters/libzootr/mido/symbols"
	"sudonters/libzootr/settings"
	"sudonters/libzootr/zecs"
)

var kind2tag = map[symbols.Kind]objects.PtrTag{
	symbols.REGION:  objects.PtrRegion,
	symbols.TRANSIT: objects.PtrTrans,
	symbols.TOKEN:   objects.PtrToken,
}

func createptrs(ocm *zecs.Ocm, syms *symbols.Table, objs *objects.Builder) {
	q := ocm.Query()
	q.Build(zecs.Load[symbols.Kind], zecs.WithOut[components.Ptr], zecs.Load[components.Name])

	for ent, tup := range q.Rows {
		kind := tup.Values[0].(symbols.Kind)
		tag, exists := kind2tag[kind]
		if !exists {
			continue
		}
		name := tup.Values[1].(components.Name)
		symbol := syms.LookUpByName(string(name))

		if symbol == nil {
			panic(fmt.Errorf("found %s in ocm but not in symbols", name))
		}

		entity := ocm.Proxy(ent)
		ptr := objects.PackPtr32(objects.Ptr32{Tag: tag, Addr: objects.Addr32(ent)})
		objs.AssociateSymbol(symbol, ptr)
		entity.Attach(components.Ptr(ptr))
	}
}

func loadsymbols(ocm *zecs.Ocm, syms *symbols.Table) error {
	batches := []tagging{
		{kind: symbols.REGION, q: []zecs.BuildQuery{zecs.With[components.RegionMarker]}},
		{kind: symbols.TRANSIT, q: []zecs.BuildQuery{zecs.With[components.Connection]}},
		{kind: symbols.TOKEN, q: []zecs.BuildQuery{zecs.With[components.TokenMarker]}},
		{kind: symbols.SCRIPTED_FUNC, q: []zecs.BuildQuery{zecs.With[components.ScriptDecl]}},
	}

	for _, batch := range batches {
		if err := batch.tagall(ocm, syms); err != nil {
			return fmt.Errorf("while tagging batch %v: %w", batch.kind, err)
		}
	}

	return nil
}

type tagging struct {
	kind symbols.Kind
	q    []zecs.BuildQuery
}

func (this tagging) tagall(ocm *zecs.Ocm, syms *symbols.Table) error {
	q := ocm.Query()
	q.Build(zecs.Load[name], this.q...)
	rows, err := q.Execute()
	if err != nil {
		return err
	}
	for ent, tup := range rows.All {
		entity := ocm.Proxy(ent)
		name := string(tup.Values[0].(name))
		syms.Declare(name, this.kind)
		entity.Attach(this.kind)
	}
	return nil
}

func loadscripts(ocm *zecs.Ocm, env *mido.CompileEnv) error {
	q := ocm.Query()
	q.Build(
		zecs.Load[name],
		zecs.Load[components.ScriptDecl],
		zecs.Load[components.ScriptSource],
		zecs.WithOut[components.RuleParsed],
	)

	rows, rowErr := q.Execute()
	if rowErr != nil {
		return rowErr
	}
	decls := make(map[string]string, rows.Len())

	for _, tup := range rows.All {
		decl := tup.Values[1].(components.ScriptDecl)
		body := tup.Values[2].(components.ScriptSource)
		decls[string(decl)] = string(body)
	}

	if err := (env.BuildScriptedFuncs(decls)); err != nil {
		return fmt.Errorf("while building scripted func declarations: %w", err)
	}

	eng := ocm.Engine()
	for entity, tup := range rows.All {
		name := tup.Values[0].(name)
		script, exists := env.ScriptedFuncs.Get(string(name))
		if !exists {
			panic(fmt.Errorf("somehow scripted func %s is missing, a mystery", name))
		}
		eng.SetValues(entity, zecs.Values{components.ScriptParsed{script.Body}})
	}

	return nil
}

func aliassymbols(ocm *zecs.Ocm, syms *symbols.Table) error {
	q := ocm.Query()
	q.Build(zecs.Load[name], zecs.With[components.TokenMarker])
	eng := ocm.Engine()
	rows, err := q.Execute()
	if err != nil {
		return err
	}

	for id, tup := range rows.All {
		name := string(tup.Values[0].(name))
		original := syms.LookUpByName(name)

		switch original.Kind {
		case symbols.FUNCTION, symbols.BUILT_IN_FUNCTION, symbols.COMPILER_FUNCTION, symbols.SCRIPTED_FUNC:
			continue
		case symbols.TOKEN:
			alias := escape(name)
			syms.Alias(original, alias)

			if err := (eng.SetValues(id, zecs.Values{components.AliasingName(alias)})); err != nil {
				return fmt.Errorf("while aliasing %q: %w", alias, err)
			}
		default:
			panic(fmt.Errorf("expected to only alias function or token: %s", original))
		}

	}

	return nil
}

func installCompilerFunctions(these *settings.Model) mido.ConfigureCompiler {

	return func(env *mido.CompileEnv) {
		hasNotesForSong := env.Symbols.Declare("has_notes_for_song", symbols.BUILT_IN_FUNCTION)
		isTrickEnabled := func(args []ast.Node, _ ast.Rewriting) (ast.Node, error) {
			switch arg := args[0].(type) {
			case ast.String:
				return ast.Boolean(these.Logic.Tricks[string(arg)]), nil
			default:
				return nil, fmt.Errorf("is_trick_enabled expects string as first argument got %#v", arg)
			}
		}

		hasAllNotesForSong := func(args []ast.Node, _ ast.Rewriting) (ast.Node, error) {
			if !settings.HasFlag(these.Logic.Shuffling.Flags, settings.ShuffleOcarinaNotes) {
				return ast.Boolean(true), nil
			}

			return ast.Invoke{
				Target: ast.IdentifierFrom(hasNotesForSong),
				Args:   args,
			}, nil
		}

		isTrialSkipped := ConstCompileFunc(true)
		regionHasShortcuts := ConstCompileFunc(false)
		hasTodAccess := ConstCompileFunc(true)
		hadNightStart := ConstCompileFunc(these.Logic.Spawns.TimeOfDay.IsNight())

		mido.WithCompilerFunctions(func(*mido.CompileEnv) optimizer.CompilerFunctionTable {
			return optimizer.CompilerFunctionTable{
				"region_has_shortcuts":   regionHasShortcuts,
				"is_trick_enabled":       isTrickEnabled,
				"had_night_start":        hadNightStart,
				"has_all_notes_for_song": hasAllNotesForSong,
				"at_dampe_time":          hasTodAccess,
				"at_day":                 hasTodAccess,
				"at_night":               hasTodAccess,
				"is_trial_skipped":       isTrialSkipped,
			}
		})(env)
	}
}

func installConnectionGenerator(ocm *zecs.Ocm) mido.ConfigureCompiler {
	return func(env *mido.CompileEnv) {
		env.Optimize.AddOptimizer(func(ce *mido.CompileEnv) ast.Rewriter {
			var conngen ConnectionGenerator
			conngen.Nodes = tracking.NewNodes(ocm)
			conngen.Tokens = tracking.NewTokens(ocm)
			conngen.Symbols = ce.Symbols
			conngen.Objects = ce.Objects

			return optimizer.NewConnectionGeneration(ce.Optimize.Context, ce.Symbols, conngen)
		})

	}
}

var escaping = regexp.MustCompile("['()[\\]-]")

func escape(name string) string {
	name = escaping.ReplaceAllLiteralString(name, "")
	return strings.ReplaceAll(name, " ", "_")
}

type ConnectionGenerator struct {
	Nodes   tracking.Nodes
	Tokens  tracking.Tokens
	Symbols *symbols.Table
	Objects *objects.Builder
}

func (this ConnectionGenerator) AddConnectionTo(region string, rule ast.Node) (*symbols.Sym, error) {
	hash := ast.Hash(rule)
	suffix := fmt.Sprintf("#%s#%16x", region, hash)
	tokenName := components.NameF("Token%s", suffix)

	if symbol := this.Symbols.LookUpByName(string(tokenName)); symbol != nil {
		return symbol, nil
	}

	token := this.Tokens.Named(tokenName)
	placement := this.Nodes.Placement(components.NameF("Place%s", suffix))

	placement.Fixed(token)
	ptr := objects.PackPtr32(objects.Ptr32{Tag: objects.PtrToken, Addr: objects.Addr32(token.Entity())})
	token.Attach(components.Event{}, ptr)

	node := this.Nodes.Region(components.Name(region))
	edge := node.Has(placement)
	edge.Proxy.Attach(components.RuleParsed{rule})

	symbol := this.Symbols.Declare(string(tokenName), symbols.TOKEN)
	this.Objects.AssociateSymbol(symbol, ptr)
	return symbol, nil
}

func ConstCompileFunc(b bool) optimizer.CompilerFunction {
	node := ast.Boolean(b)
	return func(n []ast.Node, r ast.Rewriting) (ast.Node, error) {
		return node, nil
	}
}
