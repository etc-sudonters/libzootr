package ruleparser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/etc-sudonters/substrate/peruse"
)

const (
	_ peruse.Precedence = iota
	LOWEST
	OR
	AND
	NOT
	EQ
	INDEX
	PARENS
)

func NewRulesGrammar() peruse.Grammar[Tree] {
	g := peruse.NewGrammar[Tree]()

	g.Parse(TokenTrue, parseBool)
	g.Parse(TokenFalse, parseBool)
	g.Parse(TokenIdentifier, parseIdentifierExpr)
	g.Parse(TokenNumber, parseNumber)
	g.Parse(TokenOpenParen, parseParenExpr)
	g.Parse(TokenString, parseString)
	g.Parse(TokenUnaryNot, parsePrefixNot)

	g.Infix(OR, parseBoolOpExpr, TokenOr)
	g.Infix(AND, parseBoolOpExpr, TokenAnd)
	g.Infix(EQ, parseBinOp, TokenEq, TokenNotEq, TokenLt, TokenContains)
	g.Infix(INDEX, parseSubscript, TokenOpenBracket)
	g.Infix(PARENS, parseCall, TokenOpenParen)

	return g
}

func BoolOpFromTok(t peruse.Token) BoolOpKind {
	switch s := strings.ToLower(t.Literal); s {
	case string(BoolOpAnd):
		return BoolOpAnd
	case string(BoolOpOr):
		return BoolOpOr
	default:
		panic(fmt.Errorf("invalid boolop %q", t))
	}
}

func BinOpFromTok(t peruse.Token) BinOpKind {
	switch t.Literal {
	case string(BinOpLt):
		return BinOpLt
	case string(BinOpEq):
		return BinOpEq
	case string(BinOpNotEq):
		return BinOpNotEq
	case string(BinOpContains):
		return BinOpContains
	default:
		panic(fmt.Errorf("invalid binop %q", t))
	}
}

func parseParenExpr(p *peruse.Parser[Tree]) (Tree, error) {
	p.Consume()
	e, err := p.ParseAt(LOWEST)
	if err != nil {
		return nil, err
	}

	if p.Next.Is(TokenComma) {
		e, err = parseTuple(p, e)
		if err != nil {
			return nil, err
		}
	}

	if !p.Expect(TokenCloseParen) {
		return nil, fmt.Errorf(
			"PARENEXPR: expected %q but got %q",
			TokenTypeString(TokenCloseParen),
			p.Next,
		)
	}

	return e, nil
}

func parseTuple(p *peruse.Parser[Tree], left Tree) (Tree, error) {
	elems := []Tree{left}

	for p.Expect(TokenComma) {
		p.Consume()
		elem, err := p.ParseAt(LOWEST)
		if err != nil {
			return nil, err
		}

		elems = append(elems, elem)
	}

	return &Tuple{Elems: elems}, nil
}

func parseIdentifierExpr(p *peruse.Parser[Tree]) (Tree, error) {
	return &Identifier{p.Cur.Literal}, nil
}

func parseBoolOpExpr(p *peruse.Parser[Tree], left Tree, parentPrecedence peruse.Precedence) (Tree, error) {
	thisTok := p.Cur
	p.Consume()
	right, err := p.ParseAt(parentPrecedence)
	if err != nil {
		return nil, err
	}
	b := BoolOp{
		Left:  left,
		Op:    BoolOpFromTok(thisTok),
		Right: right,
	}

	return &b, nil
}

func parseBinOp(p *peruse.Parser[Tree], left Tree, bp peruse.Precedence) (Tree, error) {
	thisTok := p.Cur
	p.Consume()
	right, err := p.ParseAt(bp)
	if err != nil {
		return nil, err
	}

	b := BinOp{
		Left:  left,
		Op:    BinOpFromTok(thisTok),
		Right: right,
	}

	return &b, nil
}

func parseCall(p *peruse.Parser[Tree], left Tree, bp peruse.Precedence) (Tree, error) {
	if p.Expect(TokenCloseParen) { // fn()
		return &Call{Callee: left}, nil
	}

	var args []Tree
	for {
		p.Consume()
		arg, err := p.ParseAt(LOWEST)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
		if !p.Expect(TokenComma) {
			break
		}
	}

	if err := p.ExpectOrError(TokenCloseParen); err != nil {
		return nil, peruse.UnexpectedAt("FNCALL", err)
	}

	c := Call{Callee: left, Args: args}
	return &c, nil
}

func parseSubscript(p *peruse.Parser[Tree], left Tree, bp peruse.Precedence) (Tree, error) {
	p.Consume()
	index, err := p.ParseAt(LOWEST)
	if err != nil {
		return nil, err
	}

	if err = p.ExpectOrError(TokenCloseBracket); err != nil {
		return nil, peruse.UnexpectedAt("SUBSCRIPT", err)
	}

	s := Subscript{Target: left, Index: index}
	return &s, nil
}

func parseString(p *peruse.Parser[Tree]) (Tree, error) {
	s := &Literal{Value: p.Cur.Literal, Kind: LiteralStr}
	return s, nil
}

func parseNumber(p *peruse.Parser[Tree]) (Tree, error) {
	n, err := strconv.ParseFloat(p.Cur.Literal, 64)
	if err != nil {
		return nil, fmt.Errorf("cannot parse %q as number", p.Cur)
	}
	return &Literal{Value: n, Kind: LiteralNum}, nil
}

func parseBool(p *peruse.Parser[Tree]) (Tree, error) {
	return &Literal{Value: p.Cur.Literal == trueWord, Kind: LiteralBool}, nil
}

func parsePrefixNot(p *peruse.Parser[Tree]) (Tree, error) {
	thisTok := p.Cur
	p.Consume()
	target, err := p.ParseAt(NOT)
	if err != nil {
		return nil, err
	}
	switch thisTok.Literal {
	case notWord:
		u := UnaryOp{
			Op:     UnaryNot,
			Target: target,
		}
		return &u, nil
	default:
		return nil, fmt.Errorf("unexpected unary op %q", thisTok)
	}
}
