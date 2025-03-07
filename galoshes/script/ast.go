package script

import (
	"fmt"
	"sudonters/libzootr/internal/table"
	"sync"

	"github.com/etc-sudonters/substrate/peruse"
	"github.com/etc-sudonters/substrate/slipup"
)

func ParseScript(script string, env *AstEnv) (AstNode, error) {
	lexer := NewLexer(script)
	parser := peruse.NewParser(grammar, lexer)
	result, parseErr := parser.ParseAt(peruse.LOWEST)
	if parseErr != nil {
		return nil, slipup.Describe(parseErr, "while parsing")
	}

	subs := make(Substitutions, 16)
	if err := Annotate(result, env, subs); err != nil {
		return nil, slipup.Describe(err, "while annotating types")
	}

	if err := SubstituteTypes(result, subs); err != nil {
		return nil, slipup.Describe(err, "while infering types")
	}

	return result, nil
}

var _ QueryNode = (*FindNode)(nil)
var _ QueryNode = (*InsertNode)(nil)
var _ AstNode = (*RuleDeclNode)(nil)
var _ AstNode = (*InsertTripletNode)(nil)
var _ ClauseNode = (*TripletClauseNode)(nil)
var _ ClauseNode = (*RuleClauseNode)(nil)
var _ ValueNode = (*VarNode)(nil)
var _ ValueNode = (*EntityNode)(nil)
var _ ValueNode = (*BoolNode)(nil)
var _ ValueNode = (*NumberNode)(nil)
var _ ValueNode = (*StringNode)(nil)
var _ ValueNode = (*AttrNode)(nil)

func GetTypes[T AstNode](nodes []T) []Type {
	tys := make([]Type, len(nodes))
	for i := range tys {
		tys[i] = nodes[i].GetType()
	}

	return tys
}

type AstKind uint8
type AstNode interface {
	NodeKind() AstKind
	GetType() Type
}
type ValueNode interface {
	AstNode
	isValue()
}
type ClauseNode interface {
	AstNode
	isClause()
}
type QueryNode interface {
	AstNode
	isQuery()
}

const (
	_ AstKind = iota
	AST_FIND
	AST_INSERT
	AST_RULE_DECL
	AST_INSERT_TRIPLET
	AST_TRIPLET_CLAUSE
	AST_RULE_CLAUSE
	AST_VAR
	AST_ENTITY
	AST_ATTR
	AST_NUMBER
	AST_BOOL
	AST_STRING
)

func NewAstEnv() *AstEnv {
	env := new(AstEnv)
	env.names = make(map[string]Type)
	env.attrs = make(map[string]AttrId)
	return env
}

type AttrId = table.ColumnId

type AstEnv struct {
	names  map[string]Type
	attrs  map[string]AttrId
	parent *AstEnv
}

var nextTypeVar TypeVar = 1
var typeVarLock = &sync.Mutex{}

func NextTypeVar() TypeVar {
	typeVarLock.Lock()
	defer typeVarLock.Unlock()
	curr := nextTypeVar
	nextTypeVar++
	return curr
}

func (this *AstEnv) GetNamed(name string) Type {
	ty, exists := this.names[name]
	if !exists && this.parent != nil {
		return this.parent.GetNamed(name)
	}
	return ty
}

func (this *AstEnv) GetAttr(name string) (Type, AttrId) {
	ty := this.GetNamed(name)
	if ty == nil {
		return nil, 0
	}
	id := this.GetAttrId(name)
	return ty, id
}

func (this *AstEnv) GetAttrId(name string) AttrId {
	id := this.attrs[name]
	if id == 0 && this.parent != nil {
		id = this.parent.GetAttrId(name)
	}
	return id
}

func (this *AstEnv) AddNamed(name string, ty Type) {
	if ty == nil {
		panic(fmt.Errorf("%q passed with nil type", name))
	}
	this.names[name] = ty
}

func (this *AstEnv) AddAttr(name string, id AttrId, ty Type) {
	this.AddNamed(name, ty)
	this.attrs[name] = id

}

type FindNode struct {
	Type    Type
	Env     AstEnv
	Finding []*VarNode
	Clauses []ClauseNode
	Rules   []*RuleDeclNode
}

func (this *FindNode) isQuery() {}

func (this *FindNode) GetType() Type {
	return this.Type
}

func (this *FindNode) NodeKind() AstKind {
	return AST_FIND
}

type InsertNode struct {
	Type      Type
	Env       AstEnv
	Inserting []*InsertTripletNode
	Clauses   []ClauseNode
	Rules     []*RuleDeclNode
}

func (this *InsertNode) isQuery() {}

func (this *InsertNode) NodeKind() AstKind {
	return AST_INSERT
}

func (this *InsertNode) GetType() Type {
	return this.Type
}

type RuleDeclNode struct {
	Type    Type
	Env     AstEnv
	Name    string
	Params  []*VarNode
	Clauses []ClauseNode
}

func (this *RuleDeclNode) NodeKind() AstKind {
	return AST_RULE_DECL
}

func (this *RuleDeclNode) GetType() Type {
	return this.Type
}

type TripletNode struct {
	Type      Type
	Id        *EntityNode
	Attribute *AttrNode
	Predicate ValueNode
}

func (this *TripletNode) GetType() Type {
	return this.Type
}

type TripletClauseNode struct {
	TripletNode
}

func (this *TripletClauseNode) NodeKind() AstKind {
	return AST_TRIPLET_CLAUSE
}

func (this *TripletClauseNode) isClause() {}

type InsertTripletNode struct {
	TripletNode
}

func (this *InsertTripletNode) NodeKind() AstKind {
	return AST_INSERT_TRIPLET
}

type RuleClauseNode struct {
	Type Type
	Name string
	Args []ValueNode
}

func (this *RuleClauseNode) NodeKind() AstKind {
	return AST_RULE_CLAUSE
}

func (this *RuleClauseNode) GetType() Type {
	return this.Type
}

func (this *RuleClauseNode) isClause() {}

type VarNode struct {
	Name string
	Type Type
}

func (this *VarNode) NodeKind() AstKind {
	return AST_VAR
}

func (this *VarNode) GetType() Type {
	return this.Type
}

func (this *VarNode) isValue() {}

type EntityNode struct {
	Value uint32
	Var   *VarNode
	Type  Type
}

func (this *EntityNode) NodeKind() AstKind {
	return AST_ENTITY
}
func (this *EntityNode) isValue() {}

func (this *EntityNode) GetType() Type {
	return this.Type
}

type AttrNode struct {
	Id   AttrId
	Type Type
	Name string
}

func (this *AttrNode) NodeKind() AstKind {
	return AST_ATTR
}
func (this *AttrNode) isValue() {}
func (this *AttrNode) GetType() Type {
	return this.Type
}

type NumberNode struct {
	Value float64
}

func (this *NumberNode) NodeKind() AstKind {
	return AST_NUMBER
}

func (this *NumberNode) isValue() {}

func (this *NumberNode) GetType() Type {
	return TypeNumber{}
}

type BoolNode struct {
	Value bool
}

func (this *BoolNode) NodeKind() AstKind {
	return AST_BOOL
}

func (this *BoolNode) isValue() {}

func (this *BoolNode) GetType() Type {
	return TypeBool{}
}

type StringNode struct {
	Value string
}

func (this *StringNode) NodeKind() AstKind {
	return AST_STRING
}

func (this *StringNode) GetType() Type {
	return TypeString{}
}

func (this *StringNode) isValue() {}
