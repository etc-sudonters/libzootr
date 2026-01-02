package bootstrap

import (
	"github.com/etc-sudonters/substrate/skelly/graph32"
	"sudonters/libzootr/magicbean"
	"sudonters/libzootr/zecs"
)

func explorableworldfrom(ocm *zecs.Ocm) magicbean.ExplorableWorld {
	var world magicbean.ExplorableWorld
	q := ocm.Query()
	q.Build(
		zecs.Load[magicbean.RuleCompiled],
		zecs.Load[magicbean.EdgeKind],
		zecs.Load[magicbean.Connection],
		zecs.Load[magicbean.Name],
		zecs.Optional[magicbean.RuleSource],
	)

	rows, err := q.Execute()
	PanicWhenErr(err)

	world.Edges = make(map[magicbean.Connection]magicbean.ExplorableEdge, rows.Len())
	world.Graph = graph32.WithCapacity(rows.Len() * 2)
	directed := graph32.Builder{Graph: &world.Graph}
	roots := zecs.EntitiesMatching(ocm, zecs.With[magicbean.WorldGraphRoot])
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
