package boot

import (
	"fmt"
	"log/slog"
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
	eng := ocm.Engine()
	uncompiled := ocm.Query()
	uncompiled.Build(
		zecs.Load[components.RuleOptimized],
		zecs.Load[components.Connection],
		zecs.WithOut[components.RuleCompiled],
	)

	for ent, tup := range uncompiled.Rows {
		entity := ocm.Proxy(ent)
		compiling := tup.Values[0].(components.RuleOptimized)
		edge := tup.Values[1].(components.Connection)

		connFrom, connFromErr := eng.GetValues(
			edge.From, table.ColumnIds{
				query.MustAsColumnId[components.Name](eng),
			},
		)
		if connFromErr != nil {
			return fmt.Errorf("while looking for connection.From name: %w", connFromErr)
		}

		connTo, connToErr := eng.GetValues(
			edge.To, table.ColumnIds{
				query.MustAsColumnId[components.Name](eng),
			},
		)
		if connToErr != nil {
			return fmt.Errorf("while looking for connection.To name: %w", connToErr)
		}

		slog.Debug("compiling connection rule", "from", connFrom.Values[0], "to", connTo.Values[0])
		bytecode, err := codegen.Compile(compiling.Node)
		if err != nil {
			return fmt.Errorf("while compiling/codegen: %w", err)
		}
		entity.Attach(components.RuleCompiled(bytecode))
	}

	return nil
}
