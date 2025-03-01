package query

import (
	"fmt"
	"slices"
	"sudonters/libzootr/galoshes/script"

	"github.com/etc-sudonters/substrate/skelly/hashset"
)

type FindQuery struct {
	Selecting []Expr
	Clauses   []Clause
}

type Clause struct {
	Entity, Attribute, Predicate Expr
}

type Expr interface{}
type Var struct{}
type Number struct{}
type String struct{}
type Boolean struct{}
type Attribute struct{}

func ProduceClauses(
	clauses []script.ClauseNode,
	decls map[string]*script.RuleDeclNode,
) ([]Clause, error) {
	var produced []Clause
	for i := range clauses {
		switch clause := clauses[i].(type) {
		case *script.RuleClauseNode:
			args := BuildArgs(clause, decls[clause.Name], make(map[string]script.ValueNode))
			applied, err := ApplyRuleClause(clause, decls, args, make(hashset.Hash[string], 32))
			if err != nil {
				return nil, fmt.Errorf("while applying rule %s: %w", clause.Name, err)
			}
			produced = slices.Concat(produced, applied)
		case *script.TripletClauseNode:
		default:
			return nil, fmt.Errorf("unknown clause type: %#v", clause)
		}
	}
}

func ApplyRuleClause(
	clause *script.RuleClauseNode,
	decls map[string]*script.RuleDeclNode,
	env map[string]script.ValueNode,
	called hashset.Hash[string],
) ([]Clause, error) {
	if called.Exists(clause.Name) {
		return nil, RecursionNotSupported{}
	}
	called.Add(clause.Name)
	var produced []Clause

	delete(called, clause.Name)
	return produced, nil
}

func BuildArgs(clause *script.RuleClauseNode, decl *script.RuleDeclNode, env map[string]script.ValueNode) map[string]script.ValueNode {
	args := make(map[string]script.ValueNode, len(decl.Params))
	for i, param := range decl.Params {
		name := param.Name
		if value, exists := env[name]; exists {
			args[name] = value
		} else {
			args[name] = clause.Args[i]
		}
	}

}

type RecursionNotSupported struct{}

func (this RecursionNotSupported) Error() string {
	return "recursive & mutually recursive calls are not supported and should be caught before planning"
}
