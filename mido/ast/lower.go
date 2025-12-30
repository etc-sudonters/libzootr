package ast

import (
	"errors"
	"fmt"
	"strings"
	"sudonters/libzootr/internal/ruleparser"
	"sudonters/libzootr/mido/symbols"
)

var (
	UnknownNode     = errors.New("unknown node")
	UnknownOperator = errors.New("unknown operator")
	UnknownLiteral  = errors.New("unknown literal type")

	compareOps = map[ruleparser.BinOpKind]CompareOp{
		ruleparser.BinOpEq:    CompareEq,
		ruleparser.BinOpNotEq: CompareNq,
		ruleparser.BinOpLt:    CompareLt,
	}
)

const (
	isTrickEnabledPrefix  = "logic_"
	isGlitchEnabledPrefix = "glitch_"
)

type CouldNotLowerTree struct {
	Node  ruleparser.Tree
	Cause error
}

func (err CouldNotLowerTree) Error() string {
	return fmt.Sprintf("could not lower parse tree: %s", err.Node.Type())
}

func Lower(tbl *symbols.Table, node ruleparser.Tree) (Node, error) {
	switch node := node.(type) {
	case *ruleparser.BinOp:
		switch node.Op {
		case ruleparser.BinOpContains:
			if ident, isIdent := node.Right.(*ruleparser.Identifier); isIdent && ident.Value == "dungeon_shortcuts" {
				if literal, isLiteral := node.Left.(*ruleparser.Literal); isLiteral && literal.Kind == ruleparser.LiteralStr {
					return createCall(tbl, "region_has_shortcuts", literal)
				}
			}

			return nil, fmt.Errorf("invalid contains construction: %#v", node)
		case ruleparser.BinOpEq, ruleparser.BinOpNotEq, ruleparser.BinOpLt:
			lhs, lhsErr := Lower(tbl, node.Left)
			rhs, rhsErr := Lower(tbl, node.Right)
			if lhsErr != nil || rhsErr != nil {
				return nil, CouldNotLowerTree{node, errors.Join(lhsErr, rhsErr)}
			}
			op := compareOps[node.Op]

			return Compare{
				LHS: lhs,
				RHS: rhs,
				Op:  op,
			}, nil
		default:
			return nil, CouldNotLowerTree{node, UnknownOperator}
		}
	case *ruleparser.BoolOp:
		switch node.Op {
		case ruleparser.BoolOpAnd:
			lhs, lhsErr := Lower(tbl, node.Left)
			rhs, rhsErr := Lower(tbl, node.Right)
			if lhsErr != nil || rhsErr != nil {
				return nil, CouldNotLowerTree{node, errors.Join(lhsErr, rhsErr)}
			}
			every := Every{lhs, rhs}
			return every.Flatten(), nil
		case ruleparser.BoolOpOr:
			lhs, lhsErr := Lower(tbl, node.Left)
			rhs, rhsErr := Lower(tbl, node.Right)
			if lhsErr != nil || rhsErr != nil {
				return nil, CouldNotLowerTree{node, errors.Join(lhsErr, rhsErr)}
			}
			anyOf := AnyOf{lhs, rhs}
			return anyOf.Flatten(), nil
		default:
			return nil, CouldNotLowerTree{node, UnknownOperator}
		}
	case *ruleparser.Call:
		var invoke Invoke
		var err error

		invoke.Target, err = Lower(tbl, node.Callee)
		if err != nil {
			return nil, CouldNotLowerTree{node, err}
		}

		invoke.Args = make([]Node, len(node.Args))
		for i := range node.Args {
			var argErr error
			invoke.Args[i], argErr = Lower(tbl, node.Args[i])
			if argErr != nil {
				err = errors.Join(err, argErr)
			}
		}

		return invoke, err
	case *ruleparser.Identifier:
		if trimmed, didTrim := strings.CutPrefix(node.Value, isTrickEnabledPrefix); didTrim {
			//TODO how to not special case
			if node.Value != "logic_rules" {
				return createCall(tbl, "is_trick_enabled", ruleparser.StringLiteral(trimmed))
			}
		}
		if trimmed, didTrim := strings.CutPrefix(node.Value, isGlitchEnabledPrefix); didTrim {
			//TODO hmmmm at leat there's more info now
			return createCall(tbl, "is_glitch_enabled", ruleparser.StringLiteral(trimmed))
		}
		symbol := tbl.Declare(node.Value, symbols.UNKNOWN)
		return IdentifierFrom(symbol), nil
	case *ruleparser.Literal:
		switch value := node.Value.(type) {
		case float64:
			return Number(value), nil
		case bool:
			return Boolean(value), nil
		case string:
			return String(value), nil
		default:
			return nil, CouldNotLowerTree{node, UnknownLiteral}
		}
	case *ruleparser.Subscript:
		if target, isIdent := node.Target.(*ruleparser.Identifier); isIdent && target.Value == "skipped_trials" {
			if trial, isIdent := node.Index.(*ruleparser.Identifier); isIdent {
				return createCall(tbl, "is_trial_skipped", ruleparser.StringLiteral(trial.Value))
			}
		}

		return nil, fmt.Errorf("invalid subscript construction %#v", node)
	case *ruleparser.Tuple:
		return createCall(tbl, "has", node.Elems...)
	case *ruleparser.UnaryOp:
		switch node.Op {
		case ruleparser.UnaryNot:
			body, err := Lower(tbl, node.Target)
			if err != nil {
				return nil, CouldNotLowerTree{node, err}
			}
			return Invert{body}, nil
		default:
			return nil, CouldNotLowerTree{node, UnknownOperator}
		}
	}
	return nil, CouldNotLowerTree{node, UnknownNode}
}

func createCall(tbl *symbols.Table, name string, args ...ruleparser.Tree) (Node, error) {
	symbol := tbl.Declare(name, symbols.FUNCTION)
	invoke := Invoke{
		Target: IdentifierFrom(symbol),
		Args:   make([]Node, len(args)),
	}

	var err error
	for i := range args {
		var argErr error
		invoke.Args[i], argErr = Lower(tbl, args[i])
		if argErr != nil {
			err = errors.Join(err, argErr)
		}
	}

	return invoke, err
}
