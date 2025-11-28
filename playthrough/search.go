package playthrough

import (
	"sudonters/libzootr/magicbean"
	"sudonters/libzootr/mido"
	"sudonters/libzootr/mido/compiler"
	"sudonters/libzootr/mido/objects"

	"github.com/etc-sudonters/substrate/skelly/bitset32"
	"github.com/etc-sudonters/substrate/skelly/graph32"
)

type Searches struct {
	Adult *Search
	Child *Search
}

func SearchFromRoots(vm *mido.VM, world *magicbean.ExplorableWorld) Search {
	search := Search{vm: vm, world: world}
	search.pended = world.Graph.Roots()
	return search
}

type Search struct {
	vm    *mido.VM
	world *magicbean.ExplorableWorld

	visited, pended bitset32.Bitset

	crossed bitset32.Bitset
}

func (this *Search) Pending() bitset32.Bitset {
	return bitset32.Copy(this.pended)
}

func (this *Search) Visited() bitset32.Bitset {
	return bitset32.Copy(this.visited)
}

func (this *Search) Crossed() bitset32.Bitset {
	return bitset32.Copy(this.crossed)
}

func (this *Search) Explore() SearchSphere {
	var sphere SearchSphere
	for current := range nodeiter(&this.pended).UntilEmpty {
		neighbors, _ := this.world.Graph.Successors(current)
		neighbors = neighbors.Difference(this.visited)
		for neighbor := range nodeiter(&neighbors).All {
			edge, _ := this.world.Edge(current, neighbor)
			crossed := evaluateEdge(this.vm, compiler.Bytecode(edge.Rule))
			if crossed {
				bitset32.Unset(&neighbors, neighbor)
				bitset32.Set(&this.pended, neighbor)
				bitset32.Set(&this.visited, neighbor)
				bitset32.Set(&this.crossed, edge.Entity)
				bitset32.Set(&sphere.Nodes.Reached, neighbor)
				bitset32.Set(&sphere.Edges.Crossed, edge.Entity)
			} else {
				bitset32.Set(&sphere.Edges.Pended, edge.Entity)
			}
		}

		if !neighbors.IsEmpty() {
			bitset32.Set(&sphere.Nodes.Pended, current)
		}
	}

	sphere.Edges.Total = this.Crossed()
	this.pended = bitset32.Copy(sphere.Nodes.Pended)
	return sphere
}

func evaluateEdge(vm *mido.VM, bytecode compiler.Bytecode) bool {
	answer, vmErr := vm.Execute(bytecode)
	if vmErr != nil {
		answer = objects.PackedFalse
	}
	return vm.Truthy(answer)
}

func nodeiter(set *bitset32.Bitset) bitset32.IterOf[graph32.Node] {
	return bitset32.IterT[graph32.Node](set)
}
