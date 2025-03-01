package script

import (
	"fmt"
	"strings"
)

type TypeDisplay struct {
	Sink   *strings.Builder
	indent int
}

func (this *TypeDisplay) fprint(msg string) {
	this.Sink.WriteString(msg)
}

func (this *TypeDisplay) fprintf(msg string, v ...any) {
	fmt.Fprintf(this.Sink, msg, v...)
}

func (this *TypeDisplay) fprintln(msg ...any) {
	fmt.Fprintln(this.Sink, msg...)
}

func (this *TypeDisplay) writeindent() {
	this.Sink.WriteString(strings.Repeat("  ", this.indent))
}

func (this *TypeDisplay) VisitFindNode(node *FindNode) {
	this.fprintln("find [ ")
	this.indent++
	for i := range node.Finding {
		this.writeindent()
		this.VisitVarNode(node.Finding[i])
		this.fprintln()
	}
	this.indent--
	this.fprintln("]\nwhere [")
	this.indent++
	for i := range node.Clauses {
		this.writeindent()
		this.VisitClauseNode(node.Clauses[i])
		this.fprintln()
	}
	this.indent--
	this.fprintln("]\n rules [")
	this.indent++
	for i := range node.Rules {
		this.writeindent()
		this.VisitRuleDeclNode(node.Rules[i])
		this.fprintln()
	}
	this.indent--
	this.fprintf("] %s\n", node.GetType())
}

func (this *TypeDisplay) VisitInsertNode(node *InsertNode) {
	this.fprintln("insert [ ")
	this.indent++
	for i := range node.Inserting {
		this.writeindent()
		this.VisitInsertTripletNode(node.Inserting[i])
		this.fprintln()
	}
	this.indent--
	this.fprintln("]\nwhere [")
	this.indent++
	for i := range node.Clauses {
		this.writeindent()
		this.VisitClauseNode(node.Clauses[i])
		this.fprintln()
	}
	this.indent--
	this.fprintln("]\n rules [")
	this.indent++
	for i := range node.Rules {
		this.writeindent()
		this.VisitRuleDeclNode(node.Rules[i])
		this.fprintln()
	}
	this.indent--
	this.fprintf("] %s\n", node.GetType())
}

func (this *TypeDisplay) visitTriplet(node TripletNode) {
	this.fprint("[ ")
	this.VisitValueNode(node.Id)
	this.fprint(" ")
	this.VisitValueNode(node.Attribute)
	this.fprint(" ")
	this.VisitValueNode(node.Predicate)
	this.fprintf(" ] %s", node.GetType())
}

func (this *TypeDisplay) VisitInsertTripletNode(node *InsertTripletNode) {
	this.visitTriplet(node.TripletNode)
}

func (this *TypeDisplay) VisitRuleDeclNode(node *RuleDeclNode) {
	this.fprintf("[ %s ", node.Name)
	for i := range node.Params {
		this.VisitVarNode(node.Params[i])
		this.fprint(" ")
	}
	this.fprintf("] %s [\n", node.GetType())
	this.indent++
	for i := range node.Clauses {
		this.writeindent()
		this.VisitClauseNode(node.Clauses[i])
		this.fprintln("")
	}
	this.indent--
	this.fprintln("]")
}

func (this *TypeDisplay) VisitClauseNode(node ClauseNode) {
	switch node := node.(type) {
	case *RuleClauseNode:
		this.VisitRuleClauseNode(node)
	case *TripletClauseNode:
		this.VisitTripletClauseNode(node)
	}
}

func (this *TypeDisplay) VisitTripletClauseNode(node *TripletClauseNode) {
	this.visitTriplet(node.TripletNode)
}

func (this *TypeDisplay) VisitRuleClauseNode(node *RuleClauseNode) {
	this.fprintf("[ %s ", node.Name)
	for i := range node.Args {
		this.VisitValueNode(node.Args[i])
		this.fprint(" ")
	}
	this.fprintf("] %s", node.GetType())
}

func (this *TypeDisplay) VisitAttrNode(node *AttrNode) {
	this.fprintf("%s", node.GetType())
}

func (this *TypeDisplay) VisitValueNode(node ValueNode) {
	switch node := node.(type) {
	case *StringNode:
		this.VisitStringNode(node)
	case *NumberNode:
		this.VisitNumber(node)
	case *AttrNode:
		this.VisitAttrNode(node)
	case *BoolNode:
		this.VisitBoolNode(node)
	case *EntityNode:
		this.VisitEntityNode(node)
	case *VarNode:
		this.VisitVarNode(node)
	}
}

func (this *TypeDisplay) VisitVarNode(node *VarNode) {
	this.fprintf("(%s %s)", node.Name, node.GetType())
}

func (this *TypeDisplay) VisitEntityNode(node *EntityNode) {
	if node.Var != nil {
		this.fprintf("%s ", node.Var.Name)
	}
	this.fprintf("%s", node.GetType())
}

func (this *TypeDisplay) VisitNumber(node *NumberNode) {
	this.fprintf("%s", node.GetType())
}

func (this *TypeDisplay) VisitBoolNode(node *BoolNode) {
	this.fprintf("%s", node.GetType())
}

func (this *TypeDisplay) VisitStringNode(node *StringNode) {
	this.fprintf("%s", node.GetType())
}
