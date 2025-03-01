package script

import (
	"fmt"

	"github.com/etc-sudonters/substrate/skelly/bitset64"
)

type Substitutions map[TypeVar]Type

func (this Substitutions) Combine(other Substitutions) (Substitutions, error) {
	new := make(Substitutions, max(len(this), len(other)))
	for tv, ty := range this {
		new[tv] = ty
	}
	for tv, ty := range other {
		ty, err := Substitute(ty, this)
		if err != nil {
			return new, err
		}
		new[tv] = ty
	}

	return new, nil
}

func Substitute(t Type, subs Substitutions) (Type, error) {
	if t == nil {
		panic("nil type")
	}
	switch t := t.(type) {
	case TypeVar:
		var terminal Type
		set := bitset64.Bitset{}
		seen := []TypeVar{}
		for {
			if !set.Set(uint64(t)) {
				// todo, return err instead
				return nil, fmt.Errorf("%v is a cycle: %v", t, seen)
			}
			seen = append(seen, t)
			typ, exists := subs[t]
			if !exists {
				terminal = t
				break
			}
			tv, isTv := typ.(TypeVar)
			if !isTv {
				terminal = typ
				break
			}
			t = tv
		}
		termVar, _ := terminal.(TypeVar)
		for _, tv := range seen {
			if tv == termVar {
				continue
			}
			subs[tv] = terminal
		}
		if tt, isTT := terminal.(TypeTuple); isTT {
			return Substitute(tt, subs)
		}
		return terminal, nil
	case TypeTuple:
		tt := TypeTuple{make([]Type, len(t.Types))}
		for i := range t.Types {
			ty, err := Substitute(t.Types[i], subs)
			if err != nil {
				return nil, err
			}
			tt.Types[i] = ty
		}
		return tt, nil
	case TypeString, TypeNumber, TypeBool:
		return t, nil
	default:
		return nil, fmt.Errorf("unknown type %#v", t)

	}
}

func SubstituteTypes(node AstNode, subs Substitutions) error {
	switch node := node.(type) {
	case *FindNode:
		for i := range node.Finding {
			if err := SubstituteTypes(node.Finding[i], subs); err != nil {
				return fmt.Errorf("in find select clause: %w", err)
			}
		}
		for i := range node.Clauses {
			if err := SubstituteTypes(node.Clauses[i], subs); err != nil {
				return fmt.Errorf("in find predicate clause: %w", err)
			}

		}
		for i := range node.Rules {
			if err := SubstituteTypes(node.Rules[i], subs); err != nil {
				return fmt.Errorf("in find rules clause: %w", err)
			}
		}

		ty, err := Substitute(node.Type, subs)
		if err != nil {
			return fmt.Errorf("on find: %w", err)
		}
		node.Type = ty
	case *InsertNode:
		for i := range node.Inserting {
			if err := SubstituteTypes(node.Inserting[i], subs); err != nil {
				return fmt.Errorf("in insert select clause: %w", err)
			}
		}
		for i := range node.Clauses {
			if err := SubstituteTypes(node.Clauses[i], subs); err != nil {
				return fmt.Errorf("in insert predicate clause: %w", err)
			}

		}
		for i := range node.Rules {
			if err := SubstituteTypes(node.Rules[i], subs); err != nil {
				return fmt.Errorf("in insert rules clause: %w", err)
			}
		}

		ty, err := Substitute(node.Type, subs)
		if err != nil {
			return fmt.Errorf("on insert: %w", err)
		}
		node.Type = ty
	case *RuleDeclNode:
		for i := range node.Params {
			if err := SubstituteTypes(node.Params[i], subs); err != nil {
				return fmt.Errorf("in decl params: %w", err)
			}
		}
		for i := range node.Clauses {
			if err := SubstituteTypes(node.Clauses[i], subs); err != nil {
				return fmt.Errorf("in decl clauses: %w", err)
			}
		}
		ty, err := Substitute(node.Type, subs)
		if err != nil {
			return fmt.Errorf("on decl: %w", err)
		}
		node.Type = ty
	case *InsertTripletNode:
		return SubstituteTriplet(&node.TripletNode, subs)
	case *TripletClauseNode:
		return SubstituteTriplet(&node.TripletNode, subs)
	case *RuleClauseNode:
		for i := range node.Args {
			if err := SubstituteTypes(node.Args[i], subs); err != nil {
				return fmt.Errorf("on rule clause arg: %w", err)
			}
		}
		ty, err := Substitute(node.Type, subs)
		if err != nil {
			return fmt.Errorf("on rule clause: %w", err)
		}
		node.Type = ty
	case *VarNode:
		ty, err := Substitute(node.Type, subs)
		if err != nil {
			return fmt.Errorf("on var %s: %w", node.Name, err)
		}
		node.Type = ty
	case *EntityNode:
		if node.Var != nil {
			if node.Var.Type == nil {
				return fmt.Errorf("entity %s has var but nil type", node.Var.Name)
			}
			ty, err := Substitute(node.Var.Type, subs)
			if err != nil {
				return fmt.Errorf("on entity: %w", err)
			}
			node.Type = ty
		}
	case *AttrNode:
		ty, err := Substitute(node.Type, subs)
		if err != nil {
			return fmt.Errorf("on attr %s: %w", node.Name, err)
		}
		node.Type = ty
	case *BoolNode, *NumberNode, *StringNode:
	default:
		return fmt.Errorf("unknown ast node: %#v", node)
	}
	return nil
}

func SubstituteTriplet(node *TripletNode, subs Substitutions) error {
	if err := SubstituteTypes(node.Id, subs); err != nil {
		return fmt.Errorf("on triplet id: %w", err)
	}
	if err := SubstituteTypes(node.Attribute, subs); err != nil {
		return fmt.Errorf("on triplet attribute: %w", err)
	}
	if err := SubstituteTypes(node.Predicate, subs); err != nil {
		return fmt.Errorf("on triplet predicate: %w", err)
	}
	ty, err := Substitute(node.Type, subs)
	if err != nil {
		return fmt.Errorf("on triplet: %w", err)
	}
	node.Type = ty
	return nil
}
