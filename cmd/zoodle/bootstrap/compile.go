package bootstrap

import (
	"errors"
	"sudonters/libzootr/internal"
	"sudonters/libzootr/magicbean"
	"sudonters/libzootr/mido"
	"sudonters/libzootr/mido/optimizer"
	"sudonters/libzootr/table"
	"sudonters/libzootr/table/ocm"

	"github.com/etc-sudonters/substrate/slipup"
)

func parseall(entities *ocm.Entities, codegen *mido.CodeGen) error {
	rows, err := entities.Query(
		table.Load[magicbean.RuleSource],
		table.Exists[magicbean.Connection],
		table.NotExists[magicbean.RuleParsed],
	)

	if err != nil {
		return slipup.Describe(err, "failed to find rules to parse")
	}

	var parseErr error

	for row, tup := range rows.All {
		entity, _ := entities.Proxy(row)
		source := tup.Values[0].(magicbean.RuleSource)

		parsed, err := codegen.Parse(string(source))
		if err != nil {
			parseErr = errors.Join(parseErr, slipup.Describef(err, "failed to parse %q", source))
		}

		entity.Attach(magicbean.RuleParsed{Node: parsed})
	}

	return parseErr
}

func optimizeall(entities *ocm.Entities, codegen *mido.CodeGen) error {
	rows, err := entities.Query(
		table.Load[magicbean.RuleParsed],
		table.Load[magicbean.Connection],
		table.NotExists[magicbean.RuleOptimized],
	)

	if err != nil {
		return slipup.Describe(err, "failed to find rules to optimize")
	}
	if rows.Len() == 0 {
		return nil
	}

	for ent, tup := range rows.All {
		entity, _ := entities.Proxy(ent)
		parsed := tup.Values[0].(magicbean.RuleParsed)
		edge := tup.Values[1].(magicbean.Connection)

		fromEntity, _ := entities.Proxy(edge.From)
		parent, parentErr := fromEntity.Values(table.ColumnIdFor[magicbean.Name])
		PanicWhenErr(parentErr)
		optimizer.SetCurrentLocation(codegen.Context, string(parent.Values[0].(magicbean.Name)))
		optimized, optimizeErr := codegen.Optimize(parsed.Node)
		PanicWhenErr(optimizeErr)
		entity.Attach(magicbean.RuleOptimized{Node: optimized})
	}

	return nil
}

func compileall(entities *ocm.Entities, codegen *mido.CodeGen) error {
	rows, err := entities.Query(
		table.Load[magicbean.RuleOptimized],
		table.Exists[magicbean.Connection],
		table.NotExists[magicbean.RuleCompiled],
	)

	if err != nil {
		return slipup.Describe(err, "failed to find rules to compile")
	}

	for ent, tup := range rows.All {
		entity, _ := entities.Proxy(ent)
		compiling := tup.Values[0].(magicbean.RuleOptimized)
		bytecode, err := codegen.Compile(compiling.Node)
		if err != nil {
			return slipup.Describe(err, "failed while compiling rule")
		}
		internal.PanicOnError(entity.Attach(magicbean.RuleCompiled(bytecode)))
	}

	return nil
}
