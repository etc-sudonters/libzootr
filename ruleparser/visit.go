package ruleparser

import (
	"fmt"
	"github.com/etc-sudonters/substrate/stageleft"
)

type VisitFunc[T Tree] func(node T, visit func(Tree) error) error

type Visitor struct {
	BinOp      VisitFunc[*BinOp]
	BoolOp     VisitFunc[*BoolOp]
	Call       VisitFunc[*Call]
	Identifier VisitFunc[*Identifier]
	Subscript  VisitFunc[*Subscript]
	Tuple      VisitFunc[*Tuple]
	UnaryOp    VisitFunc[*UnaryOp]
	Literal    VisitFunc[*Literal]
}

func (v Visitor) Visit(node Tree) error {
	var visit func(Tree) error
	visit = func(expr Tree) error {
		switch n := expr.(type) {
		case *BinOp:
			if v.BinOp == nil {
				if lhsErr := visit(n.Left); lhsErr != nil {
					return lhsErr
				}
				if rhsErr := visit(n.Right); rhsErr != nil {
					return rhsErr
				}

				return nil
			}
			return v.BinOp(n, visit)
		case *BoolOp:
			if v.BoolOp == nil {
				if lhsErr := visit(n.Left); lhsErr != nil {
					return lhsErr
				}
				if rhsErr := visit(n.Right); rhsErr != nil {
					return rhsErr
				}

				return nil
			}
			return v.BoolOp(n, visit)
		case *Call:
			if v.Call == nil {
				if calleeErr := visit(n.Callee); calleeErr != nil {
					return calleeErr
				}
				for _, arg := range n.Args {
					if argErr := visit(arg); argErr != nil {
						return argErr
					}
				}
				return nil
			}
			return v.Call(n, visit)
		case *Identifier:
			if v.Identifier == nil {
				return nil
			}
			return v.Identifier(n, visit)
		case *Subscript:
			if v.Subscript == nil {
				if trgtErr := visit(n.Target); trgtErr != nil {
					return trgtErr
				}
				if idxErr := visit(n.Index); idxErr != nil {
					return idxErr
				}
				return nil
			}
			return v.Subscript(n, visit)
		case *Tuple:
			if v.Tuple == nil {
				for _, elm := range n.Elems {
					if elmErr := visit(elm); elmErr != nil {
						return elmErr
					}
				}

			}
			return v.Tuple(n, visit)
		case *UnaryOp:
			if v.UnaryOp == nil {
				if trgtErr := visit(n.Target); trgtErr != nil {
					return trgtErr
				}
				return nil
			}
			return v.UnaryOp(n, visit)
		case *Literal:
			if v.Literal == nil {
				return nil
			}
			return v.Literal(n, visit)

		default:
			panic(stageleft.AttachExitCode(
				fmt.Errorf("unknown node type %T", expr),
				stageleft.ExitCode(90),
			))
		}
	}
	return visit(node)
}
