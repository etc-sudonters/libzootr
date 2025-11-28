package playthrough

import (
	"sudonters/libzootr/magicbean"

	"github.com/etc-sudonters/substrate/skelly/bitset32"
)

type NoProgressMade struct{}

func (_ NoProgressMade) Error() string { return "no progress made" }

func SearchAndCollect(searches Searches, gen *magicbean.Generation) Sphere {
	var sphere Sphere
	sphere.AdultSearch = searches.Adult.Explore()
	sphere.ChildSearch = searches.Child.Explore()
	sphere.Collected = magicbean.Inventory{}

	reached := sphere.AdultSearch.Nodes.Reached.Union(sphere.ChildSearch.Nodes.Reached)

	if reached.Len() == 0 {
		sphere.Err = NoProgressMade{}
	} else {
		sphere.Err = magicbean.CollectTokensFrom(
			&gen.Ocm,
			reached,
			sphere.Collected,
		)
		gen.Inventory.AddFrom(sphere.Collected)
	}
	return sphere
}

type Sphere struct {
	Err         error
	AdultSearch SearchSphere
	ChildSearch SearchSphere
	Collected   magicbean.Inventory
}

type SearchSphere struct {
	Nodes NodeSet
	Edges EdgeSet
}

type NodeSet struct {
	Reached, Pended bitset32.Bitset
}

func (this NodeSet) All() bitset32.Bitset {
	return this.Reached.Union(this.Pended)
}

type EdgeSet struct {
	Crossed, Pended, Total bitset32.Bitset
}

func (this EdgeSet) All() bitset32.Bitset {
	return this.Crossed.Union(this.Pended)
}
