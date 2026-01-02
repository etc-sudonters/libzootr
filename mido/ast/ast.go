package ast

import (
	"slices"
	"sudonters/libzootr/internal/ruleparser"
	"sudonters/libzootr/mido/symbols"

	"github.com/etc-sudonters/substrate/peruse"
)

type Kind string
type CompareOp uint8

const (
	_              Kind = ""
	KindAnyOf      Kind = "anyof"
	KindBool       Kind = "bool"
	KindCompare    Kind = "compare"
	KindEvery      Kind = "every"
	KindIdentifier Kind = "identifer"
	KindInvert     Kind = "invert"
	KindInvoke     Kind = "invoke"
	KindNumber     Kind = "number"
	KindString     Kind = "string"

	CompareEq = 1
	CompareNq = 2
	CompareLt = 3
)

type AnyOf []Node
type Boolean bool
type Compare struct {
	LHS, RHS Node
	Op       CompareOp
}
type Every []Node
type Identifier symbols.Index
type Invert struct {
	Inner Node
}
type Invoke struct {
	Target Node
	Args   []Node
}
type Number float64
type String string

type Node interface {
	Kind() Kind
}

func (a AnyOf) Kind() Kind      { return KindAnyOf }
func (a Boolean) Kind() Kind    { return KindBool }
func (a Compare) Kind() Kind    { return KindCompare }
func (a Every) Kind() Kind      { return KindEvery }
func (a Identifier) Kind() Kind { return KindIdentifier }
func (a Invert) Kind() Kind     { return KindInvert }
func (a Invoke) Kind() Kind     { return KindInvoke }
func (a Number) Kind() Kind     { return KindNumber }
func (a String) Kind() Kind     { return KindString }

func (every Every) Flatten() Every {
	var collected []Node

	for i := range every {
		switch node := every[i].(type) {
		case Every:
			collected = slices.Concat(collected, node.Flatten())
		default:
			collected = append(collected, node)
		}
	}

	return Every(collected)
}

func (every Every) Reduce() Node {
	var nodes []Node

	for i := range every {
		switch node := every[i].(type) {
		case Boolean:
			if !node {
				return Boolean(false)
			}
		default:
			nodes = append(nodes, node)
		}
	}

	switch len(nodes) {
	case 0:
		return Boolean(true)
	case 1:
		return nodes[0]
	default:
		return Every(nodes)
	}
}

func (anyOf AnyOf) Flatten() AnyOf {
	var collected []Node

	for i := range anyOf {
		switch ast := anyOf[i].(type) {
		case AnyOf:
			collected = slices.Concat(collected, ast.Flatten())
		default:
			collected = append(collected, ast)
		}
	}

	return AnyOf(collected)
}

func (anyof AnyOf) Reduce() Node {
	var nodes []Node

	for i := range anyof {
		switch node := anyof[i].(type) {
		case Boolean:
			if node {
				return Boolean(true)
			}
		default:
			nodes = append(nodes, node)
		}
	}

	switch len(nodes) {
	case 0:
		return Boolean(false)
	case 1:
		return nodes[0]
	default:
		return AnyOf(nodes)
	}
}

func (ident Identifier) AsIndex() symbols.Index {
	return symbols.Index(ident)
}

func IdentifierFrom(s *symbols.Sym) Identifier {
	return Identifier(s.Index)
}

func LookUpNodeInTable(tbl *symbols.Table, node Node) *symbols.Sym {
	switch node := node.(type) {
	case Identifier:
		return tbl.LookUpByIndex(node.AsIndex())
	default:
		return nil
	}
}

func Parse(input string, symbols *symbols.Table, grammar peruse.Grammar[ruleparser.Tree]) (Node, error) {
	parser := peruse.NewParser(grammar, ruleparser.NewRulesLexer(input))
	pt, parseErr := parser.ParseAt(ruleparser.LOWEST)
	if parseErr != nil {
		return nil, parseErr
	}

	return Lower(symbols, pt)
}
