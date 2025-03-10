package optimizer

import (
	"fmt"
	"sudonters/libzootr/mido/ast"
	"sudonters/libzootr/mido/symbols"
)

func InvokeBareFuncs(symbols *symbols.Table, funcs *ScriptedFunctions) ast.Rewriter {
	promote := invokebarefuncs{symbols, funcs}
	return ast.Rewriter{
		Invoke:     ast.DontRewrite[ast.Invoke](),
		Identifier: promote.Identifier,
	}
}

type invokebarefuncs struct {
	symbols *symbols.Table
	funcs   *ScriptedFunctions
}

func (this invokebarefuncs) Identifier(node ast.Identifier, _ ast.Rewriting) (ast.Node, error) {
	symbol := ast.LookUpNodeInTable(this.symbols, node)
	switch symbol.Kind {
	case symbols.BUILT_IN_FUNCTION, symbols.COMPILER_FUNCTION:
		return ast.Invoke{Target: node, Args: nil}, nil
	case symbols.FUNCTION, symbols.SCRIPTED_FUNC:
		fn, exists := this.funcs.Get(symbol.Name)
		if !exists {
			return nil, fmt.Errorf("fn %q was declared but not available in table", symbol.Name)
		}

		if len(fn.Params) == 0 {
			return ast.Invoke{Target: node, Args: nil}, nil
		}
		return nil, fmt.Errorf("expected 0-arg function, but %q has %d args", symbol.Name, len(fn.Params))
	default:
		return node, nil
	}
}
