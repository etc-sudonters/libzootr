package magicbean

import (
	"fmt"
	"sudonters/libzootr/components"
	"sudonters/libzootr/internal"
	"sudonters/libzootr/mido"
	"sudonters/libzootr/mido/code"
	"sudonters/libzootr/mido/compiler"
	"sudonters/libzootr/mido/objects"
	"sudonters/libzootr/zecs"

	"github.com/etc-sudonters/substrate/skelly/bitset32"
	"github.com/etc-sudonters/substrate/skelly/graph32"
)

type ExplorableEdge struct {
	Entity zecs.Entity
	Kind   components.EdgeKind
	Rule   components.RuleCompiled
	Conn   components.Connection
}

type ExplorableWorld struct {
	Graph graph32.Directed
	Edges map[components.Connection]ExplorableEdge
}

func (this ExplorableWorld) Edge(from, to graph32.Node) (ExplorableEdge, bool) {
	conn := components.Connection{From: zecs.Entity(from), To: zecs.Entity(to)}
	edge, exists := this.Edges[conn]
	return edge, exists
}

type Exploration struct {
	VM      *mido.VM
	Visited *bitset32.Bitset
	Pending *bitset32.Bitset
}

func (this *Exploration) evaluateRule(bytecode compiler.Bytecode) bool {
	if len(bytecode.Tape) == 1 {
		switch bytecode.Tape[0] {
		case code.PUSH_T:
			return true
		case code.PUSH_F:
			return false
		}
	}

	answer, vmErr := this.VM.Execute(bytecode)
	if vmErr != nil {
		fmt.Fprintln(this.VM.Std.Err, vmErr)
		answer = objects.PackedFalse
	}

	return this.VM.Truthy(answer)
}

func (this *Exploration) CanTransit(world *ExplorableWorld, from, to graph32.Node) bool {
	edge, exists := world.Edge(from, to)
	if !exists {
		panic(fmt.Errorf("no edge registered between %d %d", from, to))
	}
	bytecode := compiler.Bytecode(edge.Rule)
	result := this.evaluateRule(bytecode)
	return result
}

type ExplorationResults struct {
	Pending bitset32.Bitset
	Reached bitset32.Bitset
	Edges   []EdgeHandle
}

type EdgeHandle struct {
	Id  zecs.Entity
	Def components.Connection
}

func (this *ExplorableWorld) ExploreAvailableEdges(xplr *Exploration) ExplorationResults {
	var results ExplorationResults
	for current := range nodeiter(xplr.Pending).UntilEmpty {
		neighbors, err := this.Graph.Successors(current)
		internal.PanicOnError(err)
		neighbors = neighbors.Difference(*xplr.Visited)

		for neighbor := range nodeiter(&neighbors).All {
			if xplr.CanTransit(this, current, neighbor) {
				bitset32.Unset(&neighbors, neighbor)
				bitset32.Set(xplr.Pending, neighbor)
				bitset32.Set(&results.Reached, neighbor)
				bitset32.Set(xplr.Visited, neighbor)
			}

			edge, _ := this.Edge(current, neighbor)
			results.Edges = append(results.Edges, EdgeHandle{
				Def: edge.Conn,
				Id:  edge.Entity,
			})
		}

		if !neighbors.IsEmpty() {
			bitset32.Set(&results.Pending, current)
		}
	}

	return results
}

func nodeiter(set *bitset32.Bitset) bitset32.IterOf[graph32.Node] {
	return bitset32.IterT[graph32.Node](set)
}
