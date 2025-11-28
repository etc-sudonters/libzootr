package boot

import (
	"fmt"
	"sudonters/libzootr/components"
	"sudonters/libzootr/internal/query"
	"sudonters/libzootr/internal/table"
	"sudonters/libzootr/mido"
	"sudonters/libzootr/mido/optimizer"
	"sudonters/libzootr/zecs"
)

func parseall(ocm *zecs.Ocm, codegen *mido.CodeGen) error {
	q := ocm.Query()
	q.Build(
		zecs.Load[components.RuleSource],
		zecs.With[components.Connection],
		zecs.WithOut[components.RuleParsed],
	)

	for ent, tup := range q.Rows {
		entity := ocm.Proxy(ent)
		source := tup.Values[0].(components.RuleSource)

		parsed, err := codegen.Parse(string(source))
		if err != nil {
			return err
		}
		entity.Attach(components.RuleParsed{parsed})
	}

	return nil
}

func optimizeall(ocm *zecs.Ocm, codegen *mido.CodeGen) error {
	eng := ocm.Engine()
	unoptimized := ocm.Query()
	unoptimized.Build(
		zecs.Load[components.RuleParsed],
		zecs.Load[components.Connection],
		zecs.WithOut[components.RuleOptimized],
	)

	for {
		rows, err := unoptimized.Execute()
		if err != nil {
			return err
		}
		if rows.Len() == 0 {
			break
		}

		for ent, tup := range rows.All {
			entity := ocm.Proxy(ent)
			parsed := tup.Values[0].(components.RuleParsed)
			edge := tup.Values[1].(components.Connection)

			parent, parentErr := eng.GetValues(
				edge.From, table.ColumnIds{
					query.MustAsColumnId[components.Name](eng),
				},
			)
			if parentErr != nil {
				return fmt.Errorf("while looking for parents name: %w", parentErr)
			}
			optimizer.SetCurrentLocation(codegen.Context, string(parent.Values[0].(components.Name)))
			optimized, optimizeErr := codegen.Optimize(parsed.Node)
			if (optimizeErr) != nil {
				return fmt.Errorf("while optimizing: %w", optimizeErr)
			}
			entity.Attach(components.RuleOptimized{optimized})
		}
	}

	return nil
}

func compileall(ocm *zecs.Ocm, codegen *mido.CodeGen) error {
	uncompiled := ocm.Query()
	uncompiled.Build(
		zecs.Load[components.RuleOptimized],
		zecs.With[components.Connection],
		zecs.WithOut[components.RuleCompiled],
	)

	for ent, tup := range uncompiled.Rows {
		entity := ocm.Proxy(ent)
		compiling := tup.Values[0].(components.RuleOptimized)

		bytecode, err := codegen.Compile(compiling.Node)
		if err != nil {
			return fmt.Errorf("while compiling/codegen: %w", err)
		}
		entity.Attach(components.RuleCompiled(bytecode))
	}

	return nil
}
