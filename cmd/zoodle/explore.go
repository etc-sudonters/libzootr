package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"sudonters/libzootr/cmd/zoodle/bootstrap"
	"sudonters/libzootr/internal/query"
	"sudonters/libzootr/internal/settings"
	"sudonters/libzootr/internal/shufflequeue"
	"sudonters/libzootr/internal/table"
	"sudonters/libzootr/magicbean"
	"sudonters/libzootr/magicbean/tracking"
	"sudonters/libzootr/mido"
	"sudonters/libzootr/mido/objects"
	"sudonters/libzootr/zecs"

	"github.com/etc-sudonters/substrate/dontio"
)

type Age bool

const AgeAdult Age = true
const AgeChild Age = false

func fromStartingAge(start settings.StartingAge) Age {
	switch start {
	case settings.StartAgeAdult:
		return AgeAdult
	case settings.StartAgeChild:
		return AgeChild
	default:
		panic("unknown starting age")

	}
}

func explore(ctx context.Context, xplr *magicbean.Exploration, generation *magicbean.Generation, age Age) magicbean.ExplorationResults {
	pockets := magicbean.NewPockets(&generation.Inventory, &generation.Ocm)

	funcs := magicbean.BuiltIns{}
	magicbean.CreateBuiltInHasFuncs(&funcs, &pockets, &generation.Settings)
	funcs.CheckTodAccess = magicbean.ConstBool(true)
	funcs.IsAdult = magicbean.ConstBool(age == AgeAdult)
	funcs.IsChild = magicbean.ConstBool(age == AgeChild)
	funcs.IsStartingAge = magicbean.ConstBool(age == fromStartingAge(generation.Settings.Spawns.StartingAge))

	std, noStd := dontio.StdFromContext(ctx)
	if noStd != nil {
		panic("no std found in context")
	}

	vm := mido.VM{
		Objects: &generation.Objects,
		Funcs:   funcs.Table(),
		Std:     std,
		ChkQty:  funcs.Has,
	}

	xplr.VM = vm
	xplr.Objects = &generation.Objects

	return generation.World.ExploreAvailableEdges(ctx, xplr)
}

func PtrsMatching(ocm *zecs.Ocm, query ...zecs.BuildQuery) []objects.Object {
	q := ocm.Query()
	q.Build(zecs.Load[magicbean.Ptr], zecs.With[magicbean.Token])
	rows, err := q.Execute()
	bootstrap.PanicWhenErr(err)
	ptrs := make([]objects.Object, 0, rows.Len())

	for _, tup := range rows.All {
		ptr := tup.Values[0].(magicbean.Ptr)
		ptrs = append(ptrs, objects.Object(ptr))
	}

	return ptrs
}

func CollectStartingItems(generation *magicbean.Generation) {
	ocm := &generation.Ocm
	rng := &generation.Rng
	these := &generation.Settings
	eng := ocm.Engine()

	type collecting struct {
		entity zecs.Entity
		qty    float64
	}
	var starting []collecting

	collect := func(token tracking.Token, qty float64) {
		starting = append(starting, collecting{token.Entity(), qty})
	}

	collectOneEach := func(token ...tracking.Token) {
		new := make([]collecting, len(starting)+len(token))
		copy(new[len(token):], starting)
		for i, t := range token {
			new[i] = collecting{t.Entity(), 1}
		}

		starting = new
	}

	tokens := tracking.NewTokens(ocm)

	if these.Locations.OpenDoorOfTime {
		collect(tokens.MustGet("Time Travel"), 1)
	}

	collectOneEach(
		tokens.MustGet("Ocarina"),
		tokens.MustGet("Deku Shield"),
	)

	collect(tokens.MustGet("Deku Stick (1)"), 10)

	starting = append(starting, collecting{OneOfRandomly(ocm, rng, zecs.With[magicbean.Song]), 1})
	starting = append(starting, collecting{OneOfRandomly(ocm, rng, zecs.With[magicbean.DungeonReward]), 1})

	for _, collect := range starting {
		selected, err := eng.GetValues(collect.entity, table.ColumnIds{query.MustAsColumnId[magicbean.Name](eng)})
		bootstrap.PanicWhenErr(err)
		fmt.Printf("starting with %f %s\n", collect.qty, selected.Values[0].(magicbean.Name))
		generation.Inventory.Collect(collect.entity, collect.qty)
	}
}

func OneOfRandomly(ocm *zecs.Ocm, rng *rand.Rand, query ...zecs.BuildQuery) zecs.Entity {
	matching := shufflequeue.From(rng, zecs.EntitiesMatching(ocm, query...))
	randomly, err := matching.Dequeue()
	bootstrap.PanicWhenErr(err)
	return *randomly
}
