package ast

import (
	"errors"
	"fmt"
)

type Visiting func(Node) error

func (v Visiting) All(ast []Node) error {
	var err error
	for i := range ast {
		thisErr := v(ast[i])
		if thisErr != nil {
			err = errors.Join(thisErr)
		}
	}
	return err
}

func DontVisit[N Node]() VisitFunc[N] {
	return func(n N, _ Visiting) error {
		return nil
	}
}

type VisitFunc[T Node] func(T, Visiting) error

type Visitor struct {
	AnyOf      VisitFunc[AnyOf]
	Boolean    VisitFunc[Boolean]
	Compare    VisitFunc[Compare]
	Every      VisitFunc[Every]
	Identifier VisitFunc[Identifier]
	Invert     VisitFunc[Invert]
	Invoke     VisitFunc[Invoke]
	Number     VisitFunc[Number]
	String     VisitFunc[String]
	filled     bool
}

func (v *Visitor) fillIn() {
	if v.filled {
		return
	}

	if v.AnyOf == nil {
		v.AnyOf = VisitAnyOf
	}
	if v.Boolean == nil {
		v.Boolean = VisitBoolean
	}
	if v.Compare == nil {
		v.Compare = VisitCompare
	}
	if v.Every == nil {
		v.Every = VisitEvery
	}
	if v.Identifier == nil {
		v.Identifier = VisitIdentifier
	}
	if v.Invert == nil {
		v.Invert = VisitInvert
	}
	if v.Invoke == nil {
		v.Invoke = VisitInvoke
	}
	if v.Number == nil {
		v.Number = VisitNumber
	}
	if v.String == nil {
		v.String = VisitString
	}

}

func (v *Visitor) Visit(node Node) error {
	v.fillIn()
	return v.visit(node)
}

func (v *Visitor) visit(node Node) error {
	visit := v.visit
	switch node := node.(type) {
	case AnyOf:
		return v.AnyOf(node, visit)
	case Boolean:
		return v.Boolean(node, visit)
	case Compare:
		return v.Compare(node, visit)
	case Every:
		return v.Every(node, visit)
	case Identifier:
		return v.Identifier(node, visit)
	case Invert:
		return v.Invert(node, visit)
	case Invoke:
		return v.Invoke(node, visit)
	case Number:
		return v.Number(node, visit)
	case String:
		return v.String(node, visit)
	default:
		if node == nil {
			panic("visited nil node")
		}
		panic(fmt.Errorf("unknown node type: %T", node))
	}
}

func VisitAnyOf(anyof AnyOf, visit Visiting) error {
	return visit.All(anyof)
}

func VisitBoolean(_ Boolean, visit Visiting) error {
	return nil
}

func VisitCompare(compare Compare, visit Visiting) error {
	return visit.All([]Node{compare.LHS, compare.RHS})
}

func VisitEvery(every Every, visit Visiting) error {
	return visit.All(every)
}

func VisitIdentifier(_ Identifier, _ Visiting) error {
	return nil
}

func VisitInvert(invert Invert, visit Visiting) error {
	return visit(invert.Inner)
}

func VisitInvoke(invoke Invoke, visit Visiting) error {
	err := visit(invoke.Target)
	if err != nil {
		return err
	}

	for i := range invoke.Args {
		argErr := visit(invoke.Args[i])
		if argErr != nil {
			err = errors.Join(argErr)
		}
	}
	return err
}

func VisitNumber(_ Number, _ Visiting) error {
	return nil
}

func VisitString(_ String, _ Visiting) error {
	return nil
}
