package ast

import (
	"errors"
	"fmt"
)

type CouldNotRewrite struct {
	Node  Node
	Cause error
}

func (err CouldNotRewrite) Error() string {
	if err.Cause != nil {
		return fmt.Sprintf("could not rewrite %T: %s", err.Node, err.Cause)
	}

	return fmt.Sprintf("could not rewrite: %#v", err.Node)
}

func DontRewrite[N Node]() RewriteFunc[N] {
	return func(n N, _ Rewriting) (Node, error) {
		return n, nil
	}
}

func RewriteWithEvery(original Node, rw []Rewriter) (Node, error) {
	var err error
	node := original
	for i := range rw {
		node, err = rw[i].Rewrite(node)
		if node == nil {
			break
		}
		if err != nil {
			break
		}
	}

	return node, err
}

type Rewriting func(Node) (Node, error)
type RewriteFunc[T Node] func(T, Rewriting) (Node, error)

func (this Rewriting) All(ast []Node) ([]Node, error) {
	var err error
	rewritten := make([]Node, len(ast))
	for i := range ast {
		rewritten[i], err = this(ast[i])
		if err != nil {
			err = errors.Join(err)
		}
	}
	return rewritten, err
}

func (this Rewriting) AnyOf(node AnyOf) (Node, error) {
	return RewriteAnyOf(node, this)
}

func (this Rewriting) Compare(node Compare) (Node, error) {
	return RewriteCompare(node, this)
}

func (this Rewriting) Every(node Every) (Node, error) {
	return RewriteEvery(node, this)
}

func (this Rewriting) Invert(node Invert) (Node, error) {
	return RewriteInvert(node, this)
}

func (this Rewriting) Invoke(node Invoke) (Node, error) {
	return RewriteInvoke(node, this)
}

type Rewriter struct {
	AnyOf      RewriteFunc[AnyOf]
	Boolean    RewriteFunc[Boolean]
	Compare    RewriteFunc[Compare]
	Every      RewriteFunc[Every]
	Identifier RewriteFunc[Identifier]
	Invert     RewriteFunc[Invert]
	Invoke     RewriteFunc[Invoke]
	Number     RewriteFunc[Number]
	String     RewriteFunc[String]
	filled     bool
}

func (this *Rewriter) fill() {
	if this.filled {
		return
	}
	if this.AnyOf == nil {
		this.AnyOf = RewriteAnyOf
	}
	if this.Boolean == nil {
		this.Boolean = RewriteBoolean
	}
	if this.Compare == nil {
		this.Compare = RewriteCompare
	}
	if this.Every == nil {
		this.Every = RewriteEvery
	}
	if this.Identifier == nil {
		this.Identifier = RewriteIdentifier
	}
	if this.Invert == nil {
		this.Invert = RewriteInvert
	}
	if this.Invoke == nil {
		this.Invoke = RewriteInvoke
	}
	if this.Number == nil {
		this.Number = RewriteNumber
	}
	if this.String == nil {
		this.String = RewriteString
	}

}

func (this *Rewriter) Rewrite(ast Node) (Node, error) {
	this.fill()
	if ast == nil {
		panic("rewriting nil node")
	}
	switch ast := ast.(type) {
	case AnyOf:
		return this.AnyOf(ast, this.Rewrite)
	case Boolean:
		return this.Boolean(ast, this.Rewrite)
	case Compare:
		return this.Compare(ast, this.Rewrite)
	case Every:
		return this.Every(ast, this.Rewrite)
	case Identifier:
		return this.Identifier(ast, this.Rewrite)
	case Invert:
		return this.Invert(ast, this.Rewrite)
	case Invoke:
		return this.Invoke(ast, this.Rewrite)
	case Number:
		return this.Number(ast, this.Rewrite)
	case String:
		return this.String(ast, this.Rewrite)
	default:
		return nil, CouldNotRewrite{ast, ErrUnknownNode}
	}
}

func RewriteAnyOf(anyof AnyOf, rewrite Rewriting) (Node, error) {
	items, err := rewrite.All(anyof)
	return AnyOf(items).Flatten().Reduce(), err
}

func RewriteBoolean(b Boolean, _ Rewriting) (Node, error) {
	return b, nil
}

func RewriteCompare(compare Compare, rewrite Rewriting) (Node, error) {
	operands, err := rewrite.All([]Node{compare.LHS, compare.RHS})
	return Compare{
		LHS: operands[0],
		RHS: operands[1],
		Op:  compare.Op,
	}, err
}

func RewriteEvery(every Every, rewrite Rewriting) (Node, error) {
	items, err := rewrite.All(every)
	return Every(items).Flatten().Reduce(), err
}

func RewriteIdentifier(i Identifier, _ Rewriting) (Node, error) {
	return i, nil
}

func RewriteInvert(invert Invert, rewrite Rewriting) (Node, error) {
	return rewrite(invert.Inner)
}

func RewriteInvoke(invoke Invoke, rewrite Rewriting) (Node, error) {
	var rewritten Invoke
	var err error
	rewritten.Target, err = rewrite(invoke.Target)
	if err != nil {
		return nil, err
	}
	rewritten.Args, err = rewrite.All(invoke.Args)
	if err != nil {
		return nil, err
	}

	return rewritten, err
}

func RewriteNumber(n Number, _ Rewriting) (Node, error) {
	return n, nil
}

func RewriteString(s String, _ Rewriting) (Node, error) {
	return s, nil
}
