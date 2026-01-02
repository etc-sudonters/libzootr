package ruleparser

import (
	"math"

	"github.com/etc-sudonters/substrate/mirrors"
	"github.com/etc-sudonters/substrate/slipup"
)

type BinOpKind string
type BoolOpKind string
type UnaryOpKind string

type AnyNumeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

func Identify(name string) *Identifier {
	return &Identifier{Value: name}
}

// either a 'quoted literal' or Normalied_Identifier
func TokenLiteral(t string) *Literal {
	return &Literal{
		Kind:  LiteralToken,
		Value: t,
	}
}

func BoolLiteral(b bool) *Literal {
	return &Literal{
		Kind:  LiteralBool,
		Value: b,
	}
}

func NumberLiteral[N AnyNumeric](n N) *Literal {
	return &Literal{
		Kind:  LiteralNum,
		Value: float64(n),
	}
}

func StringLiteral(s string) *Literal {
	return &Literal{
		Kind:  LiteralStr,
		Value: s,
	}
}

func MakeCall(callee Tree, args []Tree) *Call {
	return &Call{
		Callee: callee,
		Args:   args,
	}
}

func MakeCallSplat(callee Tree, args ...Tree) *Call {
	return MakeCall(callee, args)
}

func AssertAs[T Tree](pt Tree) (T, error) {
	if cast, casted := pt.(T); casted {
		return cast, nil
	}

	return mirrors.Empty[T](), slipup.Createf("could not cast %+v to %s", pt, mirrors.TypeOf[T]().Name())
}

func MustAssertAs[T Tree](ast Tree) T {
	t, err := AssertAs[T](ast)
	if err != nil {
		panic(err)
	}
	return t
}

func Unify[A Tree, B Tree, C any](pt Tree, a func(A) (C, error), b func(B) (C, error)) (C, error) {
	switch pt := pt.(type) {
	case A:
		return a(pt)
	case B:
		return b(pt)
	default:
		return mirrors.Empty[C](), slipup.Createf("could not cast %+v to %s or %s", pt, mirrors.T[A]().Name(), mirrors.T[B]().Name())
	}
}

var (
	BinOpEq       BinOpKind   = "=="
	BinOpNotEq    BinOpKind   = "!="
	BinOpLt       BinOpKind   = "<"
	BinOpContains BinOpKind   = "in"
	BoolOpAnd     BoolOpKind  = "and"
	BoolOpOr      BoolOpKind  = "or"
	UnaryNot      UnaryOpKind = "not"
)

type ExprType string

const (
	ExprBinOp      = "BinOp"
	ExprBoolOp     = "BoolOp"
	ExprCall       = "Call"
	ExprIdentifier = "Identifier"
	ExprSubscript  = "Subscript"
	ExprTuple      = "Tuple"
	ExprUnaryOp    = "UnaryOp"
	ExprLiteral    = "Literal"
)

type LiteralKind string

const (
	LiteralBool  LiteralKind = "Boolean"
	LiteralNum   LiteralKind = "Number"
	LiteralStr   LiteralKind = "String"
	LiteralToken LiteralKind = "Token"
)

type (
	Tree interface {
		Type() ExprType
		exprNode()
	}
)

type FunctionDecl struct {
	Identifier string
	Body       Tree
	Parameters []string
}

type (
	BoolOp struct {
		Left  Tree
		Op    BoolOpKind
		Right Tree
	}

	Literal struct {
		Kind  LiteralKind
		Value any
	}

	Identifier struct {
		Value string
	}

	BinOp struct {
		Left  Tree
		Op    BinOpKind
		Right Tree
	}

	Call struct {
		Callee Tree
		Args   []Tree
	}

	Subscript struct {
		Target Tree
		Index  Tree
	}

	Tuple struct {
		Elems []Tree
	}

	UnaryOp struct {
		Op     UnaryOpKind
		Target Tree
	}
)

func (b *BinOp) exprNode()      {}
func (b *BoolOp) exprNode()     {}
func (c *Call) exprNode()       {}
func (i *Identifier) exprNode() {}
func (s *Subscript) exprNode()  {}
func (t *Tuple) exprNode()      {}
func (u *UnaryOp) exprNode()    {}
func (l *Literal) exprNode()    {}

func (expr *BinOp) Type() ExprType      { return ExprBinOp }
func (expr *BoolOp) Type() ExprType     { return ExprBoolOp }
func (expr *Call) Type() ExprType       { return ExprCall }
func (expr *Identifier) Type() ExprType { return ExprIdentifier }
func (expr *Subscript) Type() ExprType  { return ExprSubscript }
func (expr *Tuple) Type() ExprType      { return ExprTuple }
func (expr *UnaryOp) Type() ExprType    { return ExprUnaryOp }
func (expr *Literal) Type() ExprType    { return ExprLiteral }

func (expr *Literal) AsBool() (bool, bool) {
	if expr.Kind == LiteralBool {
		return expr.Value.(bool), true
	}
	return false, false
}

func (expr *Literal) AsNumber() (float64, bool) {
	if expr.Kind == LiteralNum {
		return expr.Value.(float64), true
	}

	return math.NaN(), false
}

func (expr *Literal) AsString() (string, bool) {
	if expr.Kind == LiteralStr {
		return expr.Value.(string), true
	}
	return "", false
}
