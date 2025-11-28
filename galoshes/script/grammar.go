package script

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/etc-sudonters/substrate/peruse"
	"github.com/etc-sudonters/substrate/slipup"
)

func assert(cond bool, msg string, v ...any) {
	if !cond {
		panic(fmt.Errorf(msg, v...))
	}
}

var grammar Grammar = NewGrammar()

type Parser = peruse.Parser[AstNode]
type Grammar = peruse.Grammar[AstNode]
type Parselet = peruse.Parselet[AstNode]

func NewGrammar() Grammar {
	g := peruse.NewGrammar[AstNode]()
	parslets := map[TokenType]Parselet{
		TOKEN_FIND:   astParseFind,
		TOKEN_INSERT: astParseInsert,
	}
	for tok, fn := range parslets {
		g.Parse(tok, fn)
	}
	return g
}

func traceErr(err error, during string) error {
	return slipup.Describef(err, "while parsing %s", during)
}

func astParseFind(p *Parser) (AstNode, error) {
	find := new(FindNode)

	returning, returningErr := astParseVariables(p)
	if returningErr != nil {
		return nil, traceErr(returningErr, "find.Finding")
	}
	find.Finding = returning
	if err := p.ExpectOrError(TOKEN_WHERE); err != nil {
		return nil, traceErr(err, "find.Clauses")
	}
	clauses, clauseErr := astParseClauses(p)
	if clauseErr != nil {
		return nil, traceErr(clauseErr, "find.Clauses")
	}
	find.Clauses = clauses
	if p.Expect(TOKEN_RULES) {
		decls, declErrs := astParseRuleDecls(p)
		if declErrs != nil {
			return nil, traceErr(declErrs, "find.Rules")
		}
		find.Rules = decls
	}
	return find, nil
}

func astParseInsert(p *Parser) (AstNode, error) {
	insert := new(InsertNode)

	triplets, tripletsErr := astParseInsertTripletNodes(p)
	if tripletsErr != nil {
		return nil, traceErr(tripletsErr, "insert.Inserting")
	}
	insert.Inserting = triplets

	if p.Expect(TOKEN_WHERE) {
		clauses, clauseErr := astParseClauses(p)
		if clauseErr != nil {
			return nil, traceErr(clauseErr, "insert.Clauses")
		}
		insert.Clauses = clauses
		if p.Expect(TOKEN_RULES) {
			decls, declErrs := astParseRuleDecls(p)
			if declErrs != nil {
				return nil, traceErr(declErrs, "insert.Rules")
			}
			insert.Rules = decls
		}
	}

	return insert, nil
}

func astParseRuleDecl(p *Parser) (*RuleDeclNode, error) {
	decl := new(RuleDeclNode)
	if err := p.ExpectOrError(TOKEN_RULE); err != nil {
		return nil, err
	}

	decl.Name = p.Cur.Literal
	args := make([]*VarNode, 0)
	for p.Expect(TOKEN_VARIABLE) {
		arg, err := astParseVariable(p)
		if err != nil {
			return nil, traceErr(err, "ruledecl.Args."+decl.Name)
		}
		args = append(args, arg)
	}
	if err := p.ExpectOrError(TOKEN_CLOSE_BRACKET); err != nil {
		return nil, traceErr(err, "ruledecl.Args."+decl.Name)
	}

	decl.Params = args
	clauses, clausesErr := astParseClauses(p)
	if clausesErr != nil {
		return nil, traceErr(clausesErr, "ruledecl.Clauses."+decl.Name)
	}
	decl.Clauses = clauses
	return decl, nil
}

func astParseVariable(p *Parser) (*VarNode, error) {
	variable := new(VarNode)
	variable.Name = p.Cur.Literal
	return variable, nil
}

func astParseAttribute(p *Parser) (*AttrNode, error) {
	attr := new(AttrNode)
	attr.Name = p.Cur.Literal
	return attr, nil
}

func astParseString(p *Parser) (*StringNode, error) {
	str := new(StringNode)
	str.Value = p.Cur.Literal
	return str, nil
}

func astParseNumber(p *Parser) (*NumberNode, error) {
	n, err := strconv.ParseFloat(p.Cur.Literal, 64)
	if err != nil {
		return nil, traceErr(err, "number")
	}
	node := new(NumberNode)
	node.Value = n
	return node, nil
}

func astParseBool(p *Parser) (*BoolNode, error) {
	b := new(BoolNode)
	b.Value = (p.Cur.Literal == trueWord)
	return b, nil
}

func astParseClause(p *Parser) (ClauseNode, error) {
	var clause ClauseNode
	var err error
	switch p.Next.Type {
	case TOKEN_RULE:
		clause, err = astParserRuleClauseNode(p)
	case TOKEN_NUMBER, TOKEN_VARIABLE:
		clause, err = astParseTripletClauseNode(p)
	default:
		return nil, fmt.Errorf("unexpected token %#v", p.Cur)
	}

	return clause, err
}

func astParserRuleClauseNode(p *Parser) (*RuleClauseNode, error) {
	clause := new(RuleClauseNode)
	if err := p.ExpectOrError(TOKEN_RULE); err != nil {
		return nil, traceErr(err, "rule-name")
	}

	clause.Name = p.Cur.Literal
	if p.Expect(TOKEN_CLOSE_BRACKET) {
		return nil, errors.New("at least one argument is needed")
	}

	values, valErr := astParseRemainingAsValues(p)
	if valErr != nil {
		return nil, traceErr(valErr, "ruleclause.Args."+clause.Name)
	}
	clause.Args = values
	return clause, nil
}

func astParseTriplet(p *Parser) (TripletNode, error) {
	triplet := TripletNode{}
	entity, entityErr := astParseEntity(p)
	if entityErr != nil {
		return triplet, traceErr(entityErr, "triplet.Entity")
	}
	if err := p.ExpectOrError(TOKEN_ATTRIBUTE); err != nil {
		return triplet, traceErr(err, "triplet.Attribute")
	}
	attr, attrErr := astParseAttribute(p)
	if attrErr != nil {
		return triplet, traceErr(attrErr, "triplet.Attribute")
	}
	value, valueErr := astParseValue(p)
	if valueErr != nil {
		return triplet, traceErr(valueErr, "triplet.Value")
	}
	p.Consume()

	triplet.Id = entity
	triplet.Attribute = attr
	triplet.Predicate = value
	assert(p.Cur.Is(TOKEN_CLOSE_BRACKET), "expected to end on close bracket")
	return triplet, nil
}

func astParseEntity(p *Parser) (*EntityNode, error) {
	entity := new(EntityNode)
	if p.Expect(TOKEN_NUMBER) {
		num, err := astParseNumber(p)
		if err != nil {
			return nil, traceErr(err, "entity.Id")
		}
		entity.Value = uint32(num.Value)
		return entity, nil
	} else if p.Expect(TOKEN_VARIABLE) {
		variable, err := astParseVariable(p)
		if err != nil {
			return nil, traceErr(err, "entity.Var")
		}
		entity.Var = variable
		return entity, nil
	} else {
		return nil, fmt.Errorf("unexpected token %#v", p.Cur)
	}
}

func astParseTripletClauseNode(p *Parser) (*TripletClauseNode, error) {
	clause := new(TripletClauseNode)
	triplet, err := astParseTriplet(p)
	clause.TripletNode = triplet
	return clause, err
}

func astParseInsertTripletNode(p *Parser) (*InsertTripletNode, error) {
	insert := new(InsertTripletNode)
	triplet, err := astParseTriplet(p)
	insert.TripletNode = triplet
	return insert, err
}

func astParseValue(p *Parser) (ValueNode, error) {
	var value ValueNode
	var err error
	if p.Expect(TOKEN_VARIABLE) {
		value, err = astParseVariable(p)
	} else if p.Expect(TOKEN_NUMBER) {
		value, err = astParseNumber(p)
	} else if p.Expect(TOKEN_STRING) {
		value, err = astParseString(p)
	} else if p.Expect(TOKEN_TRUE) || p.Expect(TOKEN_FALSE) {
		value, err = astParseBool(p)
	} else {
		return nil, fmt.Errorf("unexpected token %#v", p.Cur)
	}
	return value, err
}

func astParseVariables(p *Parser) ([]*VarNode, error) {
	return astParseMany(p, TOKEN_VARIABLE, astParseVariable)
}

func astParseClauses(p *Parser) ([]ClauseNode, error) {
	return astParseMany(p, TOKEN_OPEN_BRACKET, astParseClause)
}

func astParseRuleDecls(p *Parser) ([]*RuleDeclNode, error) {
	return astParseMany(p, TOKEN_OPEN_BRACKET, astParseRuleDecl)
}

func astParseInsertTripletNodes(p *Parser) ([]*InsertTripletNode, error) {
	return astParseMany(p, TOKEN_OPEN_BRACKET, astParseInsertTripletNode)
}

func astParseRemainingAsValues(p *Parser) ([]ValueNode, error) {
	return astParseManyUntil(p, TOKEN_CLOSE_BRACKET, astParseValue)
}

func astParseManyUntil[T AstNode](p *Parser, until TokenType, fn func(*Parser) (T, error)) ([]T, error) {
	var elms []T

	var i int64
	for !p.Expect(until) {
		if i > 99 {
			panic("stuck")
		}
		elm, err := fn(p)
		if err != nil {
			return elms, traceErr(err, strconv.FormatInt(i, 10))
		}
		elms = append(elms, elm)
		i++
	}

	return elms, nil

}

func astParseMany[T AstNode](p *Parser, expect TokenType, fn func(*Parser) (T, error)) ([]T, error) {
	var elms []T

	if err := p.ExpectOrError(TOKEN_OPEN_BRACKET); err != nil {
		return elms, err
	}

	for p.Expect(expect) {
		elm, err := fn(p)
		if err != nil {
			return elms, err
		}
		elms = append(elms, elm)
	}

	if err := p.ExpectOrError(TOKEN_CLOSE_BRACKET); err != nil {
		return nil, err
	}

	return elms, nil
}
