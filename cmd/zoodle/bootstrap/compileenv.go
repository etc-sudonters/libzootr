package bootstrap

import (
	"fmt"
	"github.com/etc-sudonters/substrate/slipup"
	"regexp"
	"slices"
	"strings"
	"sudonters/libzootr/internal/settings"
	"sudonters/libzootr/magicbean"
	"sudonters/libzootr/magicbean/tracking"
	"sudonters/libzootr/mido"
	"sudonters/libzootr/mido/ast"
	"sudonters/libzootr/mido/objects"
	"sudonters/libzootr/mido/optimizer"
	"sudonters/libzootr/mido/symbols"
	"sudonters/libzootr/table"
	"sudonters/libzootr/table/ocm"
)

var kind2tag = map[symbols.Kind]objects.PtrTag{
	symbols.REGION:  objects.PtrRegion,
	symbols.TRANSIT: objects.PtrTrans,
	symbols.TOKEN:   objects.PtrToken,
}

func createptrs(entities *ocm.Entities, syms *symbols.Table, objs *objects.Builder) {
	rows, err := entities.Query(
		table.Load[symbols.Kind],
		table.Load[magicbean.Name],
		table.NotExists[magicbean.Ptr],
	)

	slipup.PanicOnError(err)

	for ent, tup := range rows.All {
		kind := tup.Values[0].(symbols.Kind)
		tag, exists := kind2tag[kind]
		if !exists {
			continue
		}
		name := tup.Values[1].(magicbean.Name)
		symbol := syms.LookUpByName(string(name))

		if symbol == nil {
			panic(fmt.Errorf("found %s in ocm but not in symbols", name))
		}

		entity, _ := entities.Proxy(ent)
		ptr := objects.PackPtr32(objects.Ptr32{Tag: tag, Addr: objects.Addr32(ent)})
		objs.AssociateSymbol(symbol, ptr)
		entity.Attach(magicbean.Ptr(ptr))
	}
}

func loadsymbols(entities *ocm.Entities, syms *symbols.Table) error {
	batches := []tagging{
		{kind: symbols.REGION, q: []table.Q{table.Exists[magicbean.Region]}},
		{kind: symbols.TRANSIT, q: []table.Q{table.Exists[magicbean.Connection]}},
		{kind: symbols.TOKEN, q: []table.Q{table.Exists[magicbean.Token]}},
		{kind: symbols.SCRIPTED_FUNC, q: []table.Q{table.Exists[magicbean.ScriptDecl]}},
	}

	for _, batch := range batches {
		batch.tagall(entities, syms)
	}

	return nil
}

type tagging struct {
	kind symbols.Kind
	q    []table.Q
}

func (this tagging) tagall(entities *ocm.Entities, syms *symbols.Table) {
	q := []table.Q{table.Load[name]}
	q = slices.Concat(q, this.q)
	rows, err := entities.Query(q...)

	slipup.PanicOnError(err)
	for ent, tup := range rows.All {
		entity, _ := entities.Proxy(ent)
		name := string(tup.Values[0].(name))
		syms.Declare(name, this.kind)
		slipup.PanicOnError(entity.Attach(this.kind))
	}
}

func loadscripts(entities *ocm.Entities, env *mido.CompileEnv) error {
	rows, rowErr := entities.Query(
		table.Load[name],
		table.Load[magicbean.ScriptDecl],
		table.Load[magicbean.ScriptSource],
		table.NotExists[magicbean.RuleParsed],
	)

	slipup.PanicOnError(rowErr)
	decls := make(map[string]string, rows.Len())

	for _, tup := range rows.All {
		decl := tup.Values[1].(magicbean.ScriptDecl)
		body := tup.Values[2].(magicbean.ScriptSource)
		decls[string(decl)] = string(body)
	}

	slipup.PanicOnError(env.BuildScriptedFuncs(decls))

	for entity, tup := range rows.All {
		name := tup.Values[0].(name)
		script, exists := env.ScriptedFuncs.Get(string(name))
		if !exists {
			panic(fmt.Errorf("somehow scripted func %s is missing, a mystery", name))
		}
		p, _ := entities.Proxy(entity)
		slipup.PanicOnError(p.Attach(magicbean.ScriptParsed{Node: script.Body}))
	}

	return nil
}

func aliassymbols(entities *ocm.Entities, syms *symbols.Table) error {
	rows, err := entities.Query(table.Load[name], table.Exists[magicbean.Token])
	slipup.PanicOnError(err)

	for id, tup := range rows.All {
		name := string(tup.Values[0].(name))
		original := syms.LookUpByName(name)

		switch original.Kind {
		case symbols.FUNCTION, symbols.BUILT_IN_FUNCTION, symbols.COMPILER_FUNCTION, symbols.SCRIPTED_FUNC:
			continue
		case symbols.TOKEN:
			alias := escape(name)
			syms.Alias(original, alias)
			proxy, _ := entities.Proxy(id)
			slipup.PanicOnError(proxy.Attach(magicbean.AliasingName(alias)))
		default:
			panic(fmt.Errorf("expected to only alias function or token: %s", original))
		}

	}

	return nil
}

func installCompilerFunctions(these *settings.Zootr) mido.ConfigureCompiler {

	return func(env *mido.CompileEnv) {
		hasNotesForSong := env.Symbols.Declare("has_notes_for_song", symbols.BUILT_IN_FUNCTION)
		needsHeartForDamageMult := env.Symbols.Declare("needs_hearts_for_damage_multipler", symbols.BUILT_IN_FUNCTION)
		checkTod := env.Symbols.Declare("check_tod", symbols.BUILT_IN_FUNCTION)
		isGlitchEnabled := func(args []ast.Node, _ ast.Rewriting) (ast.Node, error) {
			switch arg := args[0].(type) {
			case ast.String:
				return ast.Boolean(these.Skills.Glitches[string(arg)]), nil
			default:
				return nil, fmt.Errorf("is_glitch_enabled expects string as first argument got %#v", arg)
			}
		}

		isTrickEnabled := func(args []ast.Node, _ ast.Rewriting) (ast.Node, error) {
			switch arg := args[0].(type) {
			case ast.String:
				return ast.Boolean(these.Skills.Tricks[string(arg)]), nil
			default:
				return nil, fmt.Errorf("is_trick_enabled expects string as first argument got %#v", arg)
			}
		}

		hasAllNotesForSong := func(args []ast.Node, _ ast.Rewriting) (ast.Node, error) {
			if !these.Shuffling.OcarinaNotes {
				return ast.Boolean(true), nil
			}

			return ast.Invoke{
				Target: ast.IdentifierFrom(hasNotesForSong),
				Args:   args,
			}, nil
		}

		canLiveDmg := func(args []ast.Node, rewriting ast.Rewriting) (ast.Node, error) {
			invokes := make([]ast.Node, 3)
			invokes[0] = ast.Invoke{
				Target: ast.IdentifierFrom(needsHeartForDamageMult),
				Args:   []ast.Node{args[0]},
			}
			invokes[1] = ast.Invoke{
				Target: ast.IdentifierFrom(env.Symbols.LookUpByName("Fairy")),
			}
			invokes[2] = ast.Invoke{
				Target: ast.IdentifierFrom(env.Symbols.LookUpByName("can_use")),
				Args:   []ast.Node{ast.IdentifierFrom(env.Symbols.LookUpByName("Nayrus_Love"))},
			}

			switch len(args) {
			case 3:
				if !args[1].(ast.Boolean) {
					invokes[1] = ast.Boolean(false)
				}
				if !args[2].(ast.Boolean) {
					invokes[2] = ast.Boolean(false)
				}
			case 2:
				if !args[1].(ast.Boolean) {
					invokes[1] = ast.Boolean(false)
				}
			case 1:
				break
			default:
				return nil, fmt.Errorf("can_live_dmg expects between 1 and 3 args, received %v", len(args))
			}

			return ast.AnyOf(invokes), nil
		}

		isTrialSkipped := func(args []ast.Node, _ ast.Rewriting) (ast.Node, error) {
			return ast.Boolean(false), nil
		}

		regionHasShortcuts := func(args []ast.Node, _ ast.Rewriting) (ast.Node, error) {
			return ast.Boolean(false), nil
		}

		needsTodChecks := func(tod string) optimizer.CompilerFunction {
			return func(args []ast.Node, _ ast.Rewriting) (ast.Node, error) {
				if !these.Entrances.AffectedTodChecks() {
					return ast.Boolean(true), nil
				}

				return ast.Invoke{
					Target: ast.IdentifierFrom(checkTod),
					Args:   []ast.Node{ast.String(tod)},
				}, nil
			}
		}

		hadNightStart := ConstCompileFunc(these.Starting.TimeOfDay.IsNight())
		for i, name := range settings.Names() {
			symbol := env.Symbols.Declare(name, symbols.SETTING)
			env.Objects.AssociateSymbol(
				symbol,
				objects.PackPtr32(objects.Ptr32{Tag: objects.PtrSetting, Addr: objects.Addr32(i)}),
			)
		}

		mido.WithCompilerFunctions(func(*mido.CompileEnv) optimizer.CompilerFunctionTable {
			return optimizer.CompilerFunctionTable{
				"region_has_shortcuts":   regionHasShortcuts,
				"is_glitch_enabled":      isGlitchEnabled,
				"is_trick_enabled":       isTrickEnabled,
				"had_night_start":        hadNightStart,
				"has_all_notes_for_song": hasAllNotesForSong,
				"at_dampe_time":          needsTodChecks("dampe"),
				"at_day":                 needsTodChecks("day"),
				"at_night":               needsTodChecks("night"),
				"is_trial_skipped":       isTrialSkipped,
				"has_soul":               ConstCompileFunc(true),
				"can_live_dmg":           canLiveDmg,
			}
		})(env)
	}
}

func installConnectionGenerator(entities *ocm.Entities) mido.ConfigureCompiler {
	return func(env *mido.CompileEnv) {
		env.Optimize.AddOptimizer(func(ce *mido.CompileEnv) ast.Rewriter {
			var conngen ConnectionGenerator
			var err error
			conngen.Nodes, err = tracking.NewNodes(entities)
			slipup.PanicOnError(err)
			conngen.Tokens, err = tracking.NewTokens(entities)
			slipup.PanicOnError(err)
			conngen.Symbols = ce.Symbols
			conngen.Objects = ce.Objects

			return optimizer.NewConnectionGeneration(ce.Optimize.Context, ce.Symbols, conngen)
		})

	}
}

var escaping = regexp.MustCompile(`['()[\]-]`)

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
	tokenName := magicbean.NameF("Token%s", suffix)

	if symbol := this.Symbols.LookUpByName(string(tokenName)); symbol != nil {
		return symbol, nil
	}

	token, tokenErr := this.Tokens.Named(tokenName)
	slipup.PanicOnError(tokenErr)
	placement := this.Nodes.Placement(magicbean.NameF("Place%s", suffix))

	placement.Fixed(token)
	ptr := objects.PackPtr32(objects.Ptr32{Tag: objects.PtrToken, Addr: objects.Addr32(token.Entity())})
	token.Attach(magicbean.Event{}, ptr)

	node := this.Nodes.Region(magicbean.Name(region))
	edge := node.Has(placement)
	edge.Proxy.Attach(magicbean.RuleParsed{Node: rule})

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
