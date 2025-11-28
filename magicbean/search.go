package magicbean

import (
	"io"
	"sudonters/libzootr/mido"

	"github.com/etc-sudonters/substrate/dontio"
	"github.com/etc-sudonters/substrate/skelly/bitset32"
)

type Search struct {
	Visited    bitset32.Bitset
	Pending    bitset32.Bitset
	Generation *Generation
	Age        Age
}

func (this *Search) Visit() SearchResult {
	vm := CreateVMForAge(this.Generation, this.Age)
	xplr := Exploration{
		Visited: &this.Visited,
		Pending: &this.Pending,
		VM:      &vm,
	}
	results := this.Generation.World.ExploreAvailableEdges(&xplr)

	this.Pending = results.Pending
	return SearchResult{
		Reached: results.Reached,
		Edges:   results.Edges,
	}
}

type SearchResult struct {
	Reached bitset32.Bitset
	Edges   []EdgeHandle
}

type Age bool

const AgeAdult Age = true
const AgeChild Age = false

func (this Age) String() string {
	if this == AgeAdult {
		return "Adult"
	}
	return "Child"
}

func CreateVMForAge(generation *Generation, age Age) mido.VM {
	pockets := NewPockets(generation.Inventory, &generation.Ocm)

	funcs := BuiltIns{}
	CreateBuiltInHasFuncs(&funcs, &pockets, generation.Settings.Logic.Shuffling.Flags)
	funcs.CheckTodAccess = ConstBool(true)
	funcs.IsAdult = ConstBool(age == AgeAdult)
	funcs.IsChild = ConstBool(age == AgeChild)
	funcs.IsStartingAge = ConstBool(age == Age(generation.Settings.Logic.Spawns.StartAge))

	return mido.VM{
		Objects: &generation.Objects,
		Funcs:   funcs.Table(),
		Std: &dontio.Std{
			Out: io.Discard,
			Err: io.Discard,
			In:  dontio.AlwaysErrReader{io.EOF},
		},
		ChkQty: funcs.Has,
	}
}

func explore(xplr *Exploration, generation *Generation, age Age) ExplorationResults {
	panic("")
}
