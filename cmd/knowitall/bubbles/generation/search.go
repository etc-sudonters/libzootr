package generation

import (
	"slices"
	"strings"
	"sudonters/libzootr/cmd/knowitall/bubbles/explore"
	"sudonters/libzootr/magicbean"
	"sudonters/libzootr/magicbean/tracking"
	"sudonters/libzootr/playthrough"
	"sudonters/libzootr/zecs"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/etc-sudonters/substrate/skelly/bitset32"
)

func runSphere(msg explore.ExploreSphere, searches playthrough.Searches, nameTable tracking.NameTable, gen *magicbean.Generation) tea.Cmd {
	return func() tea.Msg {
		sphere := playthrough.SearchAndCollect(searches, gen)
		named := explore.NamedSphere{Error: sphere.Err}
		nameEdges(sphere, nameTable, &named)
		nameNodes(sphere, nameTable, &named)
		nameSphereTokens(sphere, nameTable, &named)
		nameAllTokens(gen.Inventory, nameTable, &named)
		slices.SortFunc(named.Tokens, func(a, b explore.NamedToken) int {
			return strings.Compare(string(a.Name), string(b.Name))
		})
		slices.SortFunc(named.AllTokens, func(a, b explore.NamedToken) int {
			return strings.Compare(string(a.Name), string(b.Name))
		})
		return explore.SphereExplored{Sphere: named}
	}
}

func nameEdges(sphere playthrough.Sphere, names tracking.NameTable, named *explore.NamedSphere) {
	adult := &sphere.AdultSearch
	child := &sphere.ChildSearch
	edges := adult.Edges.All().Union(child.Edges.All())
	named.Edges = make([]explore.NamedEdge, 0, edges.Len())
	for edge := range bitset32.IterT[zecs.Entity](&edges).UntilEmpty {
		name := names[edge]
		index := uint32(len(named.Edges))
		named.Edges = append(named.Edges, explore.NamedEdge{edge, name})

		if bitset32.IsSet(&adult.Edges.Crossed, edge) {
			bitset32.Set(&named.Adult.Edges.Crossed, index)
		}
		if bitset32.IsSet(&adult.Edges.Pended, edge) {
			bitset32.Set(&named.Adult.Edges.Pended, index)
		}
		if bitset32.IsSet(&child.Edges.Crossed, edge) {
			bitset32.Set(&named.Child.Edges.Crossed, index)
		}
		if bitset32.IsSet(&child.Edges.Pended, edge) {
			bitset32.Set(&named.Child.Edges.Pended, index)
		}
	}

	named.Adult.Edges.Total = sphere.AdultSearch.Edges.Total
	named.Child.Edges.Total = sphere.ChildSearch.Edges.Total
}

func nameNodes(sphere playthrough.Sphere, names tracking.NameTable, named *explore.NamedSphere) {
	adult := &sphere.AdultSearch
	child := &sphere.ChildSearch
	nodes := adult.Nodes.All().Union(child.Nodes.All())
	for node := range bitset32.IterT[zecs.Entity](&nodes).UntilEmpty {
		name := names[node]
		index := uint32(len(named.Nodes))
		named.Nodes = append(named.Nodes, explore.NamedNode{node, name})

		if bitset32.IsSet(&adult.Nodes.Reached, node) {
			bitset32.Set(&named.Adult.Nodes.Reached, index)
		}
		if bitset32.IsSet(&adult.Nodes.Pended, node) {
			bitset32.Set(&named.Adult.Nodes.Pended, index)
		}
		if bitset32.IsSet(&child.Nodes.Reached, node) {
			bitset32.Set(&named.Child.Nodes.Reached, index)
		}
		if bitset32.IsSet(&child.Nodes.Pended, node) {
			bitset32.Set(&named.Child.Nodes.Pended, index)
		}
	}
}

func nameSphereTokens(sphere playthrough.Sphere, names tracking.NameTable, named *explore.NamedSphere) {
	named.Tokens = make([]explore.NamedToken, 0, len(sphere.Collected))
	for id, qty := range sphere.Collected {
		token := explore.NamedToken{Id: id, Qty: qty, Name: names[id]}
		named.Tokens = append(named.Tokens, token)
	}
}

func nameAllTokens(inventory magicbean.Inventory, names tracking.NameTable, named *explore.NamedSphere) {
	named.TokenMap = make(map[zecs.Entity]explore.NamedToken, len(inventory))
	named.AllTokens = make([]explore.NamedToken, 0, len(inventory))
	for id, qty := range inventory {
		token := explore.NamedToken{Id: id, Qty: qty, Name: names[id]}
		named.TokenMap[id] = token
		named.AllTokens = append(named.AllTokens, token)
	}
}
