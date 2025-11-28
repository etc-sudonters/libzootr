package fill

import (
	"context"
	"errors"
	"fmt"
	"sudonters/libzootr/internal"
	"sudonters/libzootr/magicbean"
	"sudonters/libzootr/zecs"
)

var _ Algorithm = (*Assumed)(nil)

type Assumed struct {
	Adult     magicbean.Exploration
	Child     magicbean.Exploration
	Inventory magicbean.Inventory
	World     magicbean.ExplorableWorld
}

func (this Assumed) Fill(ctx context.Context, pools Pools) (Ledger, error) {
	var placingErr error = nil
	var ledger = LedgerFor(pools)
	var assertRequeuedBoth = func(t Token, l Location) {
		internal.PanicOnError(pools.RequeuePair(t, l))
	}
placing:
	for {
		select {
		case <-ctx.Done():
			placingErr = errors.Join(PlacementCanceled, ctx.Err())
			break placing

		default:
			token, location, err := pools.DequeuePair()
			if err != nil {
				placingErr = err
				break placing
			}

			if removed := this.Inventory.Remove(zecs.Entity(token), 1); removed != 1 {
				assertRequeuedBoth(token, location)
				placingErr = fmt.Errorf("attempted to remove %v from inventory but none remain", token)
				break placing
			}

			ledger.Add(location, token)
		}
	}

	return ledger, placingErr
}
