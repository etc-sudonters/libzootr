package boot

import (
	"errors"
	"github.com/etc-sudonters/substrate/skelly/graph32"
	"sudonters/libzootr/components"
	"sudonters/libzootr/magicbean"
	"sudonters/libzootr/zecs"
)

func explorableworldfrom(ocm *zecs.Ocm) (magicbean.ExplorableWorld, error) {
	var world magicbean.ExplorableWorld
	q := ocm.Query()
	q.Build(
		zecs.Load[components.RuleCompiled],
		zecs.Load[components.EdgeKind],
		zecs.Load[components.Connection],
	)

	rows, err := q.Execute()
	if err != nil {
		return world, err
	}

	world.Edges = make(map[components.Connection]magicbean.ExplorableEdge, rows.Len())
	world.Graph = graph32.WithCapacity(rows.Len() * 2)
	directed := graph32.Builder{Graph: &world.Graph}
	roots := zecs.SliceMatching(ocm, zecs.With[components.WorldGraphRoot])
	if len(roots) == 0 {
		return world, errors.New("no graph roots loaded")
	}
	for _, root := range roots {
		directed.AddRoot(graph32.Node(root))
	}

	for entity, tup := range rows.All {
		trans := tup.Values[2].(components.Connection)
		directed.AddEdge(graph32.Node(trans.From), graph32.Node(trans.To))
		edge := magicbean.ExplorableEdge{
			Entity: entity,
			Rule:   tup.Values[0].(components.RuleCompiled),
			Kind:   tup.Values[1].(components.EdgeKind),
		}
		world.Edges[trans] = edge
	}

	return world, nil
}
