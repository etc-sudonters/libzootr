package bootstrap

import (
	"slices"
	"sudonters/libzootr/magicbean"
	"sudonters/libzootr/table"
	"sudonters/libzootr/table/ocm"

	"github.com/etc-sudonters/substrate/skelly/graph32"
	"github.com/etc-sudonters/substrate/slipup"
)

func explorableworldfrom(entities *ocm.Entities) magicbean.ExplorableWorld {
	var world magicbean.ExplorableWorld
	rows, err := entities.Query(
		table.Load[magicbean.RuleCompiled],
		table.Load[magicbean.EdgeKind],
		table.Load[magicbean.Connection],
		table.Load[magicbean.Name],
		table.Optional[magicbean.RuleSource],
	)
	slipup.PanicOnError(err)

	world.Edges = make(map[magicbean.Connection]magicbean.ExplorableEdge, rows.Len())
	world.Graph = graph32.WithCapacity(rows.Len() * 2)
	directed := graph32.Builder{Graph: &world.Graph}
	matching, rootsErr := entities.Matching(table.Exists[magicbean.WorldGraphRoot])
	slipup.PanicOnError(rootsErr)
	roots := slices.Collect(matching)

	if len(roots) == 0 {
		panic("no graph roots loaded")
	}
	for _, root := range roots {
		directed.AddRoot(graph32.Node(root))
	}

	for entity, tup := range rows.All {
		trans := tup.Values[2].(magicbean.Connection)
		directed.AddEdge(graph32.Node(trans.From), graph32.Node(trans.To))
		edge := magicbean.ExplorableEdge{
			Entity: entity,
			Rule:   tup.Values[0].(magicbean.RuleCompiled),
			Kind:   tup.Values[1].(magicbean.EdgeKind),
			Name:   tup.Values[3].(magicbean.Name),
		}

		src := tup.Values[4]
		if src != nil {
			edge.Src = src.(magicbean.RuleSource)
		}

		world.Edges[trans] = edge
	}

	return world
}
