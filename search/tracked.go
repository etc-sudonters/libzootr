package search

import (
	"sudonters/libzootr/magicbean"
	"sudonters/libzootr/mido"

	"github.com/etc-sudonters/substrate/skelly/bitset32"
)

type Tracked struct {
	Age magicbean.Age
	vm  *mido.VM

	VisitedNodes bitset32.Bitset
	PendingNodes bitset32.Bitset
}

func NewTracked(age magicbean.Age, vm *mido.VM) Tracked {
	return Tracked{
		Age: age,
		vm:  vm,
	}
}

func (this *Tracked) Search(world *magicbean.ExplorableWorld) (Result, error) {
	var searchResult Result
	xplr := magicbean.Exploration{
		VM:      this.vm,
		Visited: &this.VisitedNodes,
		Pending: &this.PendingNodes,
	}

	exploration := world.ExploreAvailableEdges(&xplr)
	this.PendingNodes = exploration.Pending
	searchResult.VisitedNodes = bitset32.Copy(this.VisitedNodes)
	searchResult.PendingNodes = bitset32.Copy(this.PendingNodes)
	searchResult.ReachedNodes = exploration.Reached

	searchResult.VisitedEdges = exploration.Edges
	for _, handle := range searchResult.VisitedEdges {
		if bitset32.IsSet(&this.VisitedNodes, handle.Def.To) {
			bitset32.Set(&searchResult.CrossedEdges, handle.Id)
		}
	}

	if searchResult.ReachedNodes.Len() == 0 {
		return searchResult, ErrNoProgress
	}

	return searchResult, nil
}

func MakeSphere(tracker *Tracked, result Result) Result {
	var sphere Result
	sphere.ReachedNodes = result.ReachedNodes
	sphere.PendingNodes = result.PendingNodes.Intersect(result.ReachedNodes)
	sphere.VisitedNodes = bitset32.Copy(sphere.ReachedNodes)
	sphere.VisitedEdges = make([]magicbean.EdgeHandle, 0, len(result.VisitedEdges)/2)

	for _, handle := range result.VisitedEdges {
		if !bitset32.IsSet(&sphere.VisitedNodes, handle.Def.From) {
			continue
		}
		if bitset32.IsSet(&result.CrossedEdges, handle.Id) {
			bitset32.Set(&sphere.CrossedEdges, handle.Id)
		}
		sphere.VisitedEdges = append(sphere.VisitedEdges, handle)
	}

	return sphere
}
