package fill

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"sudonters/libzootr/components"
	"sudonters/libzootr/internal/shuffle"
	"sudonters/libzootr/zecs"
)

// see https://github.com/cjohnson57/RandomizerAlgorithms
type Algorithm interface {
	// all algorithms assume they have prefiltered pools
	// for example, when placings songs on dungeon rewards pools.Tokens is all
	// songs and pools.Locations are the locations that songs are allowed to be
	// placed on
	Fill(context.Context, Pools) (Ledger, error)
}

// not _really_ errors but do communicate why we stopped placing
var TokenPoolEmpty = errors.New("token pool empty")
var LocationPoolEmpty = errors.New("location pool empty")
var PlacementCanceled = errors.New("placement canceled")
var FailedTokenRequeue = errors.New("failed to requeue token")
var FailedLocationRequque = errors.New("failed to requeue location")

type Token zecs.Entity
type Location zecs.Entity

type Placeable interface {
	Token | Location
}

type TokenPool = Pool[Token]
type LocationPool = Pool[Location]

type Pools struct {
	Tokens    TokenPool
	Locations LocationPool
}

func (this Pools) DequeuePair() (Token, Location, error) {
	if this.Tokens.Len() == 0 {
		return 0, 0, TokenPoolEmpty
	}

	if this.Locations.Len() == 0 {
		return 0, 0, LocationPoolEmpty
	}

	token, _ := this.Tokens.Dequeue()
	location, _ := this.Locations.Dequeue()
	return token, location, nil
}

func (this Pools) RequeuePair(t Token, l Location) error {
	if err := this.Tokens.Requeue(t); err != nil {
		return errors.Join(FailedTokenRequeue, err)
	}

	if err := this.Locations.Requeue(l); err != nil {
		return errors.Join(FailedLocationRequque, err)
	}

	return nil
}

type Pool[P Placeable] struct {
	*shuffle.Q[P]
}

func LedgerFor(pools Pools) Ledger {
	return Ledger{make([]pair, 0, min(pools.Locations.Len(), pools.Tokens.Len()))}
}

type pair struct {
	Location
	Token
}

type Ledger struct {
	inner []pair
}

func (this *Ledger) Add(l Location, t Token) {
	this.inner = append(this.inner, pair{l, t})
}

func (this Ledger) All(yield func(Location, Token) bool) {
	for _, pair := range this.inner {
		if !yield(pair.Location, pair.Token) {
			return
		}
	}
}

// this should be someone else's job, the struct has public fields for a reason
func NewPlacing[P Placeable](ocm *zecs.Ocm, rng *rand.Rand) (Pool[P], error) {
	var placing Pool[P]
	q := ocm.Query()
	var p P
	switch any(&p).(type) {
	case *Token:
		q.Build(zecs.With[components.TokenMarker])
	case *Location:
		q.Build(zecs.With[components.PlacementLocationMarker])
	default:
		panic("unreachable")
	}

	rows, err := q.Execute()
	if err != nil {
		return placing, fmt.Errorf("could not initialize %T: %w", p, err)
	}

	entities := make([]P, 0, rows.Len())

	for row, _ := range rows.All {
		p := P(row)
		entities = append(entities, p)
	}

	placing.Q = shuffle.From(rng, entities)
	return placing, nil
}
