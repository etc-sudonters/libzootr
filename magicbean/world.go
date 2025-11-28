package magicbean

import (
	"fmt"
	"sudonters/libzootr/components"
	"sudonters/libzootr/magicbean/tracking"
	"sudonters/libzootr/zecs"
)

type World struct {
	inventory Inventory
}

func PlaceAlwaysItems(set *tracking.Set) {
	timeTravel := set.Tokens.MustGet("Time Travel")
	pedestal := set.Nodes.Placement("Master Sword Pedestal")
	pedestal.Holding(timeTravel)
}

func PromoteRemainingDefaultTokens(ocm *zecs.Ocm) error {
	q := ocm.Query()
	q.Build(
		zecs.Load[components.DefaultPlacement],
		zecs.WithOut[components.HoldsToken],
	)

	rows, err := q.Execute()
	if err != nil {
		return err
	}

	for entity, tup := range rows.All {
		token := tup.Values[0].(components.DefaultPlacement)
		proxy := ocm.Proxy(entity)
		err := proxy.Attach(components.HoldsToken(token))
		if err != nil {
			return fmt.Errorf("while promoting default token for %v: %w", entity, err)
		}
	}

	return nil
}
