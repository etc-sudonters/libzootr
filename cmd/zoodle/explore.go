package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"slices"
	"sudonters/libzootr/cmd/zoodle/bootstrap"
	"sudonters/libzootr/internal"
	"sudonters/libzootr/internal/settings"
	"sudonters/libzootr/internal/shufflequeue"
	"sudonters/libzootr/magicbean"
	"sudonters/libzootr/magicbean/tracking"
	"sudonters/libzootr/mido"
	"sudonters/libzootr/mido/objects"
	"sudonters/libzootr/table"
	"sudonters/libzootr/table/ocm"

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
	pockets := magicbean.NewPockets(&generation.Inventory, generation.Entities)

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

func PtrsMatching(entities *ocm.Entities, query ...table.Q) []objects.Object {
	q := []table.Q{table.Load[magicbean.Ptr], table.Exists[magicbean.Token]}
	q = slices.Concat(q, query)
	rows, err := entities.Query(q...)
	bootstrap.PanicWhenErr(err)
	ptrs := make([]objects.Object, 0, rows.Len())

	for _, tup := range rows.All {
		ptr := tup.Values[0].(magicbean.Ptr)
		ptrs = append(ptrs, objects.Object(ptr))
	}

	return ptrs
}

func CollectStartingItems(generation *magicbean.Generation) {
	entities := generation.Entities
	rng := &generation.Rng
	these := &generation.Settings

	type collecting struct {
		entity ocm.Entity
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

	tokens, err := tracking.NewTokens(entities)
	internal.PanicOnError(err)

	if these.Locations.OpenDoorOfTime {
		collect(tokens.MustGet("Time Travel"), 1)
	}

	collectOneEach(
		tokens.MustGet("Ocarina"),
		tokens.MustGet("Deku Shield"),
	)

	collect(tokens.MustGet("Deku Stick (1)"), 10)

	starting = append(starting, collecting{OneOfRandomly(entities, rng, table.Exists[magicbean.Song]), 1})
	starting = append(starting, collecting{OneOfRandomly(entities, rng, table.Exists[magicbean.DungeonReward]), 1})

	for _, collect := range starting {
		proxy, _ := entities.Proxy(collect.entity)
		values, err := proxy.Values(table.ColumnIdFor[magicbean.Name])
		internal.PanicOnError(err)
		fmt.Printf("starting with %f %s\n", collect.qty, values.Values[0].(magicbean.Name))
		generation.Inventory.Collect(collect.entity, collect.qty)
	}
}

func OneOfRandomly(entities *ocm.Entities, rng *rand.Rand, query ...table.Q) ocm.Entity {
	matched, err := entities.Matching(query...)
	internal.PanicOnError(err)

	matching := shufflequeue.From(rng, slices.Collect(matched))
	randomly, err := matching.Dequeue()
	bootstrap.PanicWhenErr(err)
	return *randomly
}
