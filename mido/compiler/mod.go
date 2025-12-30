package compiler

import (
	"fmt"
	"maps"
	"slices"
	"sudonters/libzootr/mido/ast"
	"sudonters/libzootr/mido/code"
	"sudonters/libzootr/mido/objects"
	"sudonters/libzootr/mido/symbols"
)

type Bytecode struct {
	Tape   code.Instructions
	Consts []objects.Index
	Names  map[objects.Index]string
}

func (this *Bytecode) concat(tape code.Instructions) int {
	written := len(tape)
	this.Tape = slices.Concat(this.Tape, tape)
	return written
}

func Compile(nodes ast.Node, symbols *symbols.Table, objs *objects.Builder) (Bytecode, error) {
	var compiler compiler
	var bytecode Bytecode
	compiler.symbols = symbols
	compiler.objects = objs
	compiler.code = &bytecode
	compiler.consts = map[objects.Index]struct{}{}
	compiler.names = map[objects.Index]string{}

	visiting := &compiler
	visitor := ast.Visitor{
		AnyOf:      visiting.AnyOf,
		Boolean:    visiting.Boolean,
		Compare:    visiting.Compare,
		Every:      visiting.Every,
		Identifier: visiting.Identifier,
		Invert:     visiting.Invert,
		Invoke:     visiting.Invoke,
		Number:     visiting.Number,
		String:     visiting.String,
	}
	err := visitor.Visit(nodes)
	bytecode.Consts = slices.Collect(maps.Keys(compiler.consts))
	bytecode.Names = compiler.names
	return bytecode, err
}

type compiler struct {
	tapePtr int
	symbols *symbols.Table
	objects *objects.Builder
	code    *Bytecode
	consts  map[objects.Index]struct{}
	names   map[objects.Index]string
}

func (this *compiler) emit(op code.Op, operands ...int) int {
	return this.join(code.Make(op, operands...))
}

func (this *compiler) join(emitted code.Instructions) int {
	startOfInstruction := this.tapePtr
	this.tapePtr += this.code.concat(emitted)
	return startOfInstruction
}

func (this *compiler) AnyOf(node ast.AnyOf, visit ast.Visiting) error {
	err := visit.All(node)
	if err != nil {
		return err
	}
	this.emit(code.NEED_ANY, len(node))
	return nil
}

func (this *compiler) Boolean(node ast.Boolean, visit ast.Visiting) error {
	if node {
		this.emit(code.PUSH_T)
	} else {
		this.emit(code.PUSH_F)
	}
	return nil
}

func (this *compiler) Compare(node ast.Compare, visit ast.Visiting) error {
	if err := visit.All([]ast.Node{node.RHS, node.LHS}); err != nil {
		return err
	}
	switch node.Op {
	case ast.CompareEq:
		this.emit(code.CMP_EQ)
	case ast.CompareNq:
		this.emit(code.CMP_NQ)
	case ast.CompareLt:
		this.emit(code.CMP_LT)
	default:
		return fmt.Errorf("uncompilable comparison op: %v", node.Op)
	}

	return nil
}

func (this *compiler) Every(node ast.Every, visit ast.Visiting) error {
	err := visit.All(node)
	if err != nil {
		return err
	}
	this.emit(code.NEED_ALL, len(node))
	return nil
}

func (this *compiler) Identifier(node ast.Identifier, visit ast.Visiting) error {
	symbol := this.symbols.LookUpByIndex(node.AsIndex())
	ptr := this.objects.PtrFor(symbol)
	switch symbol.Kind {
	case symbols.BUILT_IN_FUNCTION:
		this.pushPtr(code.PUSH_FUNC, ptr, symbol.Name)
	case symbols.TOKEN, symbols.SETTING:
		this.pushPtr(code.PUSH_PTR, ptr, symbol.Name)
	default:
		return fmt.Errorf("uncompilable identifier: %s", symbol)
	}
	return nil
}

func (this *compiler) Invert(node ast.Invert, visit ast.Visiting) error {
	if err := visit(node.Inner); err != nil {
		return err
	}
	this.emit(code.INVERT)
	return nil
}

func (this *compiler) Invoke(node ast.Invoke, visit ast.Visiting) error {
	callee := ast.LookUpNodeInTable(this.symbols, node.Target)
	if callee == nil {
		return fmt.Errorf("can only invoke functions, not %s", node.Target.Kind())
	}

	def := this.objects.FunctionDefinition(callee)
	if argCount := len(node.Args); def.ParamCount > -1 && def.ParamCount != argCount {
		return fmt.Errorf("%q expects %d arguments but received %d", def.Name, def.ParamCount, argCount)
	}

	if this.trySpecializeInvoke(node, callee, def) {
		return nil
	}

	if argsErr := visit.All(node.Args); argsErr != nil {
		return argsErr
	}

	if targetErr := visit(node.Target); targetErr != nil {
		return targetErr
	}

	this.emit(code.INVOKE, len(node.Args))
	return nil
}

func (this *compiler) Number(node ast.Number, visit ast.Visiting) error {
	idx := this.objects.InternNumber(float64(node))
	this.consts[idx] = struct{}{}
	this.emit(code.PUSH_CONST, int(idx))
	return nil
}

func (this *compiler) String(node ast.String, visit ast.Visiting) error {
	idx := this.objects.InternStr(string(node))
	this.consts[idx] = struct{}{}
	this.emit(code.PUSH_STR, int(idx))
	return nil
}

func (this *compiler) pushPtr(op code.Op, idx objects.Index, name string) {
	this.consts[idx] = struct{}{}
	this.names[idx] = name
	this.emit(op, int(idx))
}

func (this *compiler) trySpecializeInvoke(node ast.Invoke, callee *symbols.Sym, def objects.BuiltInFunctionDef) bool {
	switch callee.Name {
	case "has":
		what := ast.LookUpNodeInTable(this.symbols, node.Args[0])
		qty := node.Args[1]
		if what != nil && qty.Kind() == ast.KindNumber {
			ptr := this.objects.PtrFor(what)
			this.consts[ptr] = struct{}{}
			this.names[ptr] = what.Name
			this.emit(code.CHK_QTY, int(ptr), int(uint8(qty.(ast.Number))))
			return true
		}
	}

	if def.ParamCount == 0 {
		ptr := this.objects.PtrFor(callee)
		this.consts[ptr] = struct{}{}
		this.names[ptr] = callee.Name
		this.emit(code.INVOKE_0, int(ptr))
		return true
	}

	return false
}
