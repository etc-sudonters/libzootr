package fill

import (
	"context"
	"errors"
)

var _ Algorithm = (*Random)(nil)

// randomly pairs items and locations w/o any verification
// applications include placing junk/non-progression pools and no logic seeds
type Random struct{}

func (this Random) Fill(ctx context.Context, pools Pools) (Ledger, error) {
	var placingErr error = nil
	var ledger = LedgerFor(pools)
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
			ledger.Add(location, token)
		}
	}

	return ledger, placingErr
}
