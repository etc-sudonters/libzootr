package explore

import (
	"sudonters/libzootr/components"
	"sudonters/libzootr/magicbean"
	"sudonters/libzootr/mido/ast"
	"sudonters/libzootr/mido/symbols"
	"sudonters/libzootr/mido/vm"
	"sudonters/libzootr/playthrough"
	"sudonters/libzootr/zecs"

	tea "github.com/charmbracelet/bubbletea"
)

type sphereSelected int

func selectSphere(which int) tea.Cmd {
	return func() tea.Msg {
		return sphereSelected(which)
	}
}

type NamedToken struct {
	Id   zecs.Entity
	Name components.Name
	Qty  int
}

type NamedEdge struct {
	Id   zecs.Entity
	Name components.Name
}

type NamedNode struct {
	Id   zecs.Entity
	Name components.Name
}

type NamedSphere struct {
	I         int
	Error     error
	Edges     []NamedEdge
	Nodes     []NamedNode
	Tokens    []NamedToken
	AllTokens []NamedToken

	TokenMap map[zecs.Entity]NamedToken

	Adult playthrough.SearchSphere
	Child playthrough.SearchSphere
}

type SphereExplored struct {
	Err    error
	Sphere NamedSphere
}

type ExploreSphere struct{}

func RequestNextSphere() tea.Msg {
	return ExploreSphere{}
}

type DisassembleRule struct {
	Id zecs.Entity
}

type RuleDisassembled struct {
	Id          zecs.Entity
	Err         error
	Name        string
	Disassembly vm.Disassembly
	Inventory   magicbean.Inventory
	Source      string
	Ast, Opt    ast.Node
	Symbols     *symbols.Table
}

func RequestDisassembly(edge zecs.Entity) tea.Cmd {
	return func() tea.Msg {
		return DisassembleRule{edge}
	}
}

type EditRule struct {
	Source components.RuleSource
	Err    error
}

type LoadRuleSource zecs.Entity

func startEditingRule(id zecs.Entity) tea.Cmd {
	return func() tea.Msg {
		return LoadRuleSource(id)
	}
}
