package script

import (
	"fmt"
	"strings"
)

type AstRender struct {
	Sink   *strings.Builder
	indent int
}

func (this *AstRender) fprint(msg string) {
	this.Sink.WriteString(msg)
}

func (this *AstRender) fprintf(msg string, v ...any) {
	fmt.Fprintf(this.Sink, msg, v...)
}

func (this *AstRender) fprintln(msg ...any) {
	fmt.Fprintln(this.Sink, msg...)
}

func (this *AstRender) writeindent() {
	this.Sink.WriteString(strings.Repeat("  ", this.indent))
}

func (this *AstRender) VisitFindNode(node *FindNode) {
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
	this.fprintln("]")
}

func (this *AstRender) VisitInsertNode(node *InsertNode) {
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
	this.fprintln("]")
}

func (this *AstRender) visitTriplet(node TripletNode) {
	this.fprint("[ ")
	this.VisitValueNode(node.Id)
	this.fprint(" ")
	this.VisitValueNode(node.Attribute)
	this.fprint(" ")
	this.VisitValueNode(node.Predicate)
	this.fprint(" ]")
}

func (this *AstRender) VisitInsertTripletNode(node *InsertTripletNode) {
	this.visitTriplet(node.TripletNode)
}

func (this *AstRender) VisitRuleDeclNode(node *RuleDeclNode) {
	this.fprintf("[ %s ", node.Name)
	for i := range node.Params {
		this.VisitVarNode(node.Params[i])
		this.fprint(" ")
	}
	this.fprintf("] [\n")
	this.indent++
	for i := range node.Clauses {
		this.writeindent()
		this.VisitClauseNode(node.Clauses[i])
		this.fprintln("")
	}
	this.indent--
	this.fprintln("]")
}

func (this *AstRender) VisitClauseNode(node ClauseNode) {
	switch node := node.(type) {
	case *RuleClauseNode:
		this.VisitRuleClauseNode(node)
	case *TripletClauseNode:
		this.VisitTripletClauseNode(node)
	}
}

func (this *AstRender) VisitTripletClauseNode(node *TripletClauseNode) {
	this.visitTriplet(node.TripletNode)
}

func (this *AstRender) VisitRuleClauseNode(node *RuleClauseNode) {
	this.fprintf("[ %s ", node.Name)
	for i := range node.Args {
		this.VisitValueNode(node.Args[i])
		this.fprint(" ")
	}
	this.fprintf("]")
}

func (this *AstRender) VisitAttrNode(node *AttrNode) {
	this.fprintf("%s", node.Name)
}

func (this *AstRender) VisitValueNode(node ValueNode) {
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

func (this *AstRender) VisitVarNode(node *VarNode) {
	this.fprintf("$%s", node.Name)
}

func (this *AstRender) VisitEntityNode(node *EntityNode) {
	if node.Var != nil {
		this.VisitVarNode(node.Var)
	} else {
		this.fprintf("%d", node.Value)
	}
}

func (this *AstRender) VisitNumber(node *NumberNode) {
	this.fprintf("%f", node.Value)
}

func (this *AstRender) VisitBoolNode(node *BoolNode) {
	this.fprintf("%t", node.Value)
}

func (this *AstRender) VisitStringNode(node *StringNode) {
	this.fprintf("%s", node.Value)
}
