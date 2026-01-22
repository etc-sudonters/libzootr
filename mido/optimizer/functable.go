package optimizer

import (
	"fmt"
	"strings"
	"sudonters/libzootr/ruleparser"
	"sudonters/libzootr/mido/ast"
	"sudonters/libzootr/mido/symbols"

	"github.com/etc-sudonters/substrate/peruse"
)

func FastScriptNameFromDecl(decl string) string {
	parts := strings.Split(decl, "(")
	return parts[0]
}

func BuildScriptedFuncTable(symbolTable *symbols.Table, grammar peruse.Grammar[ruleparser.Tree], decls map[string]string) (ScriptedFunctions, error) {
	funcTable := ScriptedFunctions{make(map[string]ScriptedFunction, len(decls))}
	bodies := make(map[string]string, len(decls))

	for header, body := range decls {
		var decl ScriptedFunction
		head, declErr := ast.Parse(header, symbolTable, grammar)
		if declErr != nil {
			panic(declErr)
		}

		switch head := head.(type) {
		case ast.Invoke:
			decl.Symbol = ast.LookUpNodeInTable(symbolTable, head.Target)
			if decl.Symbol.Kind == symbols.BUILT_IN_FUNCTION {
				continue
			}
			if decl.Symbol == nil {
				panic(fmt.Errorf("did not find entry for %#v", head))
			}
			decl.Params = make([]ast.Identifier, len(head.Args))
			for i := range decl.Params {
				param := head.Args[i].(ast.Identifier)
				symbol := symbolTable.LookUpByIndex(param.AsIndex())
				symbol.SetKind(symbols.LOCAL)
				decl.Params[i] = param
			}
		case ast.Identifier:
			decl.Symbol = symbolTable.LookUpByIndex(head.AsIndex())
			if decl.Symbol.Kind == symbols.BUILT_IN_FUNCTION {
				continue
			}
			decl.Params = nil
		default:
			panic(fmt.Errorf("expected identifier or invoke style declaration, got %#v", head))
		}

		decl.Symbol.SetKind(symbols.SCRIPTED_FUNC)
		bodies[decl.Symbol.Name] = body
		funcTable.add(decl.Symbol, decl)
	}

	for name, compiling := range funcTable.tbl {
		body := bodies[name]
		nodes, bodyErr := ast.Parse(body, symbolTable, grammar)
		if bodyErr != nil {
			panic(bodyErr)
		}
		compiling.Body = nodes
		funcTable.tbl[name] = compiling
	}

	return funcTable, nil
}

type ScriptedFunction struct {
	Symbol *symbols.Sym
	Params []ast.Identifier
	Body   ast.Node
}

type ScriptedFunctions struct {
	tbl map[string]ScriptedFunction
}

func (this *ScriptedFunctions) All(yield func(string, ScriptedFunction) bool) {
	for name, sf := range this.tbl {
		if !yield(name, sf) {
			return
		}
	}
}

func (this *ScriptedFunctions) Get(name string) (ScriptedFunction, bool) {
	fn, exists := this.tbl[name]
	return fn, exists
}

func (this *ScriptedFunctions) add(sym *symbols.Sym, body ScriptedFunction) {
	if _, exists := this.tbl[sym.Name]; exists {
		return
	}

	this.tbl[sym.Name] = body
}
