package query

import (
	"fmt"
	"slices"
	"strings"
	"sudonters/libzootr/galoshes/script"
	"sudonters/libzootr/internal/table"
)

type Engine struct {
	tbl *table.Table

	attrs map[string]table.ColumnId
}

type tripletvalue struct {
	name  string
	value any
	typ   script.Type
}

type tripletId struct {
	name string
	id   table.RowId
}

type triplet struct {
	id    tripletId
	attr  string
	value tripletvalue
}

func (this triplet) String() string {
	view := &strings.Builder{}
	view.WriteString("[ ")
	if this.id.name != "" {
		fmt.Fprintf(view, "$%s ", this.id.name)
	} else {
		fmt.Fprintf(view, "%d ", this.id.id)
	}
	fmt.Fprintf(view, "%s ", this.attr)
	if this.value.name != "" {
		fmt.Fprintf(view, "$%s", this.value.name)
	} else {
		fmt.Fprintf(view, "%v", this.value.value)
	}
	view.WriteString(" ]")
	return view.String()
}

type QueryPlan struct {
	returning []string
	triplets  []triplet
}

func (this QueryPlan) String() string {
	view := &strings.Builder{}
	fmt.Fprintf(view, "returning: %v\n", this.returning)
	fmt.Fprintln(view, "triplets")
	for _, triplet := range this.triplets {
		fmt.Fprintln(view, triplet)
	}
	return view.String()
}

type planner struct {
	qp    *QueryPlan
	decls map[string]*script.RuleDeclNode
	calls []string
}

func BuildQueryPlan(ast script.QueryNode) QueryPlan {
	var qp QueryPlan
	planner := planner{&qp, nil, nil}
	planner.planQuery(ast)
	return qp
}

func (this *planner) planQuery(node script.QueryNode) {
	switch node := node.(type) {
	case *script.FindNode:
		this.planFind(node)
	default:
		panic(fmt.Errorf("unknown query type: %#v", node))
	}
	this.finalizePlan()
}

func (this *planner) planFind(node *script.FindNode) {
	this.qp.returning = make([]string, len(node.Finding))
	for i, finding := range node.Finding {
		this.qp.returning[i] = finding.Name
	}
	if len(node.Rules) > 0 {
		this.decls = make(map[string]*script.RuleDeclNode, len(node.Rules))
		for _, decl := range node.Rules {
			this.decls[decl.Name] = decl
		}
	}

	for _, clause := range node.Clauses {
		this.applyClause(clause)
	}

}

func (this *planner) applyClause(clause script.ClauseNode) {
	switch clause := clause.(type) {
	case *script.RuleClauseNode:
		this.applyRuleClause(clause)
	case *script.TripletClauseNode:
		this.applyTripletClause(clause)
	default:
		panic(fmt.Errorf("unknown clause type: %#v", clause))
	}
}

func (this *planner) applyRuleClause(clause *script.RuleClauseNode) {
	triplets := this.produceTriplets(clause, make(map[string]script.ValueNode))
	this.qp.triplets = append(this.qp.triplets, triplets...)
}

func (this *planner) produceTriplets(clause *script.RuleClauseNode, env map[string]script.ValueNode) []triplet {
	if slices.Contains(this.calls, clause.Name) {
		panic("recursive & mutually recursive calls are not supported and should be caught in ast, build query requires a correct query tree")
	}
	this.calls = append(this.calls, clause.Name)
	var triplets []triplet
	decl := this.decls[clause.Name]
	args := make(map[string]script.ValueNode, len(decl.Args))
	for i, arg := range decl.Args {
		name := arg.Name
		if value, exists := env[name]; exists {
			args[arg.Name] = value
		} else {
			args[arg.Name] = clause.Args[i]
		}
	}
	for _, clause := range decl.Clauses {
		var produced []triplet
		switch clause := clause.(type) {
		case *script.RuleClauseNode:
			produced = this.produceTriplets(clause, args)
		case *script.TripletClauseNode:
			produced = append(produced, this.produceBoundTriplet((*clause).TripletNode, args))
		default:
			panic(fmt.Errorf("unknown clause type %#v", clause))
		}
		triplets = append(triplets, produced...)
	}
	last := len(this.calls) - 1
	if this.calls[last] != clause.Name {
		panic(fmt.Errorf("last call changed some how %q -> %q", clause.Name, this.calls[last]))
	}
	this.calls = this.calls[:last]
	return triplets
}

func (this *planner) produceBoundTriplet(clause script.TripletNode, args map[string]script.ValueNode) triplet {
	if clause.Id.Var != nil {
		arg := args[clause.Id.Var.Name]
		if arg == nil {
			name := clause.Id.Var.Name
			view := script.AstRender{Sink: &strings.Builder{}}
			clause := script.TripletClauseNode{TripletNode: clause}
			view.VisitTripletClauseNode(&clause)
			panic(fmt.Errorf("%s isn't bound in %v\n%v", name, view.Sink.String(), args))
		}
		switch arg := arg.(type) {
		case *script.NumberNode:
			clause.Id = &script.EntityNode{Value: uint32(arg.Value), Type: script.TypeNumber{}}
		case *script.VarNode:
			clause.Id = &script.EntityNode{Var: arg, Type: script.TypeNumber{}}
		default:
			panic(fmt.Errorf("unsupported id value %#v", arg))
		}
	}

	if variable, isVar := clause.Value.(*script.VarNode); isVar {
		arg := args[variable.Name]
		if arg == nil {
			panic(fmt.Errorf("%s isn't bound", clause.Id.Var.Name))
		}
		clause.Value = arg
	}

	return this.produceTriplet(clause)
}

func (this *planner) produceTriplet(clause script.TripletNode) triplet {
	var trip triplet
	if clause.Id.Var != nil {
		trip.id.name = clause.Id.Var.Name
	} else {
		trip.id.id = table.RowId(clause.Id.Value)
	}
	trip.attr = clause.Attribute.Name
	bindValue(clause.Value, &trip.value)
	return trip
}

func bindValue(value script.ValueNode, trip *tripletvalue) {
	switch value := value.(type) {
	case *script.StringNode:
		trip.value = value.Value
		trip.typ = value.GetType()
	case *script.NumberNode:
		trip.value = value.Value
		trip.typ = value.GetType()
	case *script.BoolNode:
		trip.value = value.Value
		trip.typ = value.GetType()
	case *script.VarNode:
		trip.name = value.Name
		trip.value = value.GetType()
	default:
		panic(fmt.Errorf("unsupported triplet value type %#v", value))
	}
}

func (this *planner) applyTripletClause(clause *script.TripletClauseNode) {
	this.qp.triplets = append(this.qp.triplets, this.produceTriplet(clause.TripletNode))
}

func (_ planner) finalizePlan() {}
