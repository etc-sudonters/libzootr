package mido

import (
	"errors"
	"fmt"
	"iter"
	"sudonters/libzootr/internal/ruleparser"
	"sudonters/libzootr/mido/ast"
	"sudonters/libzootr/mido/compiler"
	"sudonters/libzootr/mido/objects"
	"sudonters/libzootr/mido/optimizer"
	"sudonters/libzootr/mido/symbols"

	"github.com/etc-sudonters/substrate/peruse"
)

func GlobalNames() []string {
	return globalNames[:]
}

var globalNames = []string{
	"Fire",
	"Forest",
	"Light",
	"Shadow",
	"Spirit",
	"Water",
	"adult",
	"age",
	"both",
	"either",
	"child",
}

type SourceKind string
type SourceString string

func (this SourceKind) AsSymbolKind() symbols.Kind {
	switch this {
	case SourceCheck, SourceEvent:
		return symbols.TOKEN
	case SourceTransit:
		return symbols.TRANSIT
	default:
		panic("unreachable")
	}
}

const (
	_             SourceKind = ""
	SourceCheck              = "SourceCheck"
	SourceEvent              = "SourceEvent"
	SourceTransit            = "SourceTransit"
)

type CompiledSource struct {
	CompilationSource
	ByteCode compiler.Bytecode
}

func ptr[T any](what T) *T {
	return &what
}

type CompilationSource struct {
	Kind              SourceKind
	String            SourceString
	Ast               ast.Node
	Optimized         ast.Node
	OriginatingRegion string
	Destination       string
}

type ConfigureCompiler func(*CompileEnv)

func WithCompilerFunctions(create func(*CompileEnv) optimizer.CompilerFunctionTable) ConfigureCompiler {
	return func(env *CompileEnv) {
		funcs := create(env)

		for name := range funcs {
			env.Symbols.Declare(name, symbols.COMPILER_FUNCTION)
		}

		compiler := optimizer.NewCompilerFuncs(env.Symbols, funcs)
		env.Optimize.AddOptimizer(func(*CompileEnv) ast.Rewriter {
			return compiler
		})
	}
}

func WithBuiltInFunctionDefs(create func(*CompileEnv) []objects.BuiltInFunctionDef) ConfigureCompiler {
	return func(env *CompileEnv) {
		builtins := create(env)
		for i, builtin := range builtins {
			symbol := env.Symbols.Declare(builtin.Name, symbols.BUILT_IN_FUNCTION)
			ptr := objects.PackPtr32(objects.Ptr32{
				Tag:  objects.PtrFunc,
				Addr: objects.Addr32(i),
			})
			env.Objects.DefineFunction(symbol, ptr, builtin)
		}
	}
}

func CompilerWithTokens(names []string) ConfigureCompiler {
	return func(env *CompileEnv) {
		env.Symbols.DeclareMany(symbols.TOKEN, names)
	}
}

func CompilerWithGenerationSettings(names []string) ConfigureCompiler {
	return func(env *CompileEnv) {
		for i, name := range names {
			symbol := env.Symbols.Declare(name, symbols.SETTING)
			env.Objects.AssociateSymbol(
				symbol,
				objects.PackPtr32(objects.Ptr32{Tag: objects.PtrSetting, Addr: objects.Addr32(i)}),
			)
		}
	}
}

func CompilerDefaults() ConfigureCompiler {
	return func(env *CompileEnv) {
		env.Optimize.Passes = 10

		env.Symbols.DeclareMany(symbols.GLOBAL, GlobalNames())

		env.Optimize.AddOptimizer(func(env *CompileEnv) ast.Rewriter {
			return optimizer.InlineCalls(env.Optimize.Context, env.Symbols, env.ScriptedFuncs)
		})
		env.Optimize.AddOptimizer(func(env *CompileEnv) ast.Rewriter {
			return optimizer.FoldConstants(env.Symbols)
		})
		env.Optimize.AddOptimizer(func(env *CompileEnv) ast.Rewriter {
			return optimizer.InvokeBareFuncs(env.Symbols, env.ScriptedFuncs)
		})
		env.Optimize.AddOptimizer(func(env *CompileEnv) ast.Rewriter {
			return optimizer.CollapseHas(env.Symbols)
		})
		env.Optimize.AddOptimizer(func(env *CompileEnv) ast.Rewriter {
			return optimizer.PromoteTokens(env.Symbols)
		})
	}
}

func NewCompileEnv(configure ...ConfigureCompiler) CompileEnv {
	var env CompileEnv
	env.Grammar = ruleparser.NewRulesGrammar()
	env.Symbols = ptr(symbols.NewTable())
	env.Objects = ptr(objects.NewTableBuilder())
	env.Optimize.Context = ptr(optimizer.NewCtx())

	for i := range configure {
		configure[i](&env)
	}

	return env
}

type SourceLoaded func(*CompileEnv, *CompilationSource)
type Optimizer func(*CompileEnv) ast.Rewriter
type Analyzer func(*CompileEnv) ast.Visitor

type CompileEnv struct {
	Grammar       peruse.Grammar[ruleparser.Tree]
	Symbols       *symbols.Table
	ScriptedFuncs *optimizer.ScriptedFunctions
	Objects       *objects.Builder

	Optimize Optimize
	Analysis Analysis
}

type Analysis struct {
	pre  []Analyzer
	post []Analyzer
}

func (this *Analysis) PreOptimize(v Analyzer) {
	this.pre = append(this.pre, v)
}

func (this *Analysis) PostOptimize(v Analyzer) {
	this.post = append(this.post, v)
}

type Optimize struct {
	Context     *optimizer.Context
	Optimiziers []Optimizer
	Passes      int
}

func (this *Optimize) AddOptimizer(o Optimizer) {
	this.Optimiziers = append(this.Optimiziers, o)
}

func (this *CompileEnv) BuildScriptedFuncs(declarations map[string]string) error {
	var err error
	var funcs optimizer.ScriptedFunctions
	funcs, err = optimizer.BuildScriptedFuncTable(this.Symbols, this.Grammar, declarations)
	if err != nil {
		return err
	}
	this.ScriptedFuncs = &funcs
	return nil
}

func Compiler(env *CompileEnv) CodeGen {
	optimizers := env.Optimize.Optimiziers
	analysis := env.Analysis
	codegen := CodeGen{
		Context:       env.Optimize.Context,
		env:           env,
		rewriters:     make([]ast.Rewriter, len(optimizers)),
		preanalyzers:  make([]ast.Visitor, len(analysis.pre)),
		postanalyzers: make([]ast.Visitor, len(analysis.post)),
	}
	for i := range codegen.rewriters {
		codegen.rewriters[i] = optimizers[i](env)
	}
	for i := range codegen.preanalyzers {
		codegen.preanalyzers[i] = analysis.pre[i](env)
	}
	for i := range codegen.postanalyzers {
		codegen.postanalyzers[i] = analysis.post[i](env)
	}
	return codegen
}

type CodeGen struct {
	Context       *optimizer.Context
	env           *CompileEnv
	rewriters     []ast.Rewriter
	preanalyzers  []ast.Visitor
	postanalyzers []ast.Visitor
}

func (this CodeGen) SymbolTable() *symbols.Table {
	return this.env.Symbols
}

func (this CodeGen) Parse(source string) (ast.Node, error) {
	ast, err := ast.Parse(source, this.env.Symbols, this.env.Grammar)
	return ast, err
}

func (this CodeGen) StepOptimize(node ast.Node, err *error) iter.Seq2[int, ast.Node] {
	return func(yield func(int, ast.Node) bool) {
		if !yield(0, node) {
			return
		}
		for i := range this.env.Optimize.Passes {
			for _, rw := range this.rewriters {
				node, *err = rw.Rewrite(node)
				if !yield(i+1, node) {
					return
				}
				if node == nil {
					return
				}
				if *err != nil {
					return
				}
			}
		}
	}
}

func (this CodeGen) Optimize(node ast.Node) (ast.Node, error) {
	var rewriteErr error
	for range this.env.Optimize.Passes {
		node, rewriteErr = ast.RewriteWithEvery(node, this.rewriters)
		if rewriteErr != nil {
			rewriteErr = fmt.Errorf("%w: %w", ErrOptimization, rewriteErr)
			break
		}
	}

	return node, rewriteErr
}

func (this CodeGen) Compile(node ast.Node) (compiler.Bytecode, error) {
	bytecode, compileErr := compiler.Compile(node, this.env.Symbols, this.env.Objects)
	if compileErr != nil {
		compileErr = fmt.Errorf("%w: %w", ErrCompile, compileErr)
	}
	return bytecode, compileErr

}

func (this CodeGen) CompileSource(src *CompilationSource) (compiler.Bytecode, error) {
	var bytecode compiler.Bytecode

	if src.Ast == nil {
		var astErr error
		src.Ast, astErr = ast.Parse(string(src.String), this.env.Symbols, this.env.Grammar)
		if astErr != nil {
			return bytecode, fmt.Errorf("%w: %w", ErrParse, astErr)
		}
	}

	for i := range this.env.Analysis.pre {
		this.preanalyzers[i].Visit(src.Ast)
	}

	src.Optimized = src.Ast
	for range this.env.Optimize.Passes {
		var rewriteErr error
		src.Optimized, rewriteErr = ast.RewriteWithEvery(src.Optimized, this.rewriters)
		if rewriteErr != nil {
			return bytecode, fmt.Errorf("%w: %w", ErrOptimization, rewriteErr)
		}
	}

	for i := range this.env.Analysis.post {
		this.postanalyzers[i].Visit(src.Ast)
	}

	var compileErr error
	bytecode, compileErr = compiler.Compile(src.Ast, this.env.Symbols, this.env.Objects)
	if compileErr != nil {
		compileErr = fmt.Errorf("%w: %w", ErrCompile, compileErr)
	}
	return bytecode, compileErr
}

var (
	ErrSourceLoad   = errors.New("source load")
	ErrParse        = errors.New("parsing")
	ErrOptimization = errors.New("optimization")
	ErrCompile      = errors.New("compile")
)

func StepOptimize(src string, codegen *CodeGen, err *error) iter.Seq2[int, ast.Node] {
	var node ast.Node
	node, *err = codegen.Parse(src)
	if *err != nil {
		return func(yield func(int, ast.Node) bool) {}
	}

	return codegen.StepOptimize(node, err)
}
