package mido

import (
	"errors"
	"fmt"
	"sudonters/zootler/internal/ruleparser"
	"sudonters/zootler/internal/settings"
	"sudonters/zootler/mido/ast"
	"sudonters/zootler/mido/compiler"
	"sudonters/zootler/mido/objects"
	"sudonters/zootler/mido/optimizer"
	"sudonters/zootler/mido/symbols"

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
	case SourceCheck:
		return symbols.TOKEN
	case SourceEvent:
		return symbols.EVENT
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
	Source
	ByteCode compiler.Bytecode
}

func ptr[T any](what T) *T {
	return &what
}

type Source struct {
	Kind              SourceKind
	String            SourceString
	Ast               ast.Node
	OriginatingRegion string
	Destination       string
}

type ConfigureCompiler func(*CompileEnv)

func CompilerWithFastOps(ops compiler.FastOps) ConfigureCompiler {
	return func(env *CompileEnv) {
		for name, fast := range ops {
			env.Optimize.FastOps[name] = fast
		}
	}
}

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
			env.Objects.DefineFunction(symbol, objects.PackTaggedPtr32(objects.PtrFunc, uint32(i)), builtin)
		}
	}
}

func CompilerWithTokens(names []string) ConfigureCompiler {
	return func(env *CompileEnv) {
		env.Symbols.DeclareMany(symbols.TOKEN, names)
	}
}

func CompilerDefaults() ConfigureCompiler {
	return func(env *CompileEnv) {
		env.Optimize.Passes = 10

		env.Symbols.DeclareMany(symbols.GLOBAL, GlobalNames())
		env.Symbols.DeclareMany(symbols.SETTING, settings.Names())

		env.OnSourceLoad(func(env *CompileEnv, src *Source) {
			SetCurrentLocation(env.Optimize.Context, src.OriginatingRegion)
		})
		env.Optimize.AddOptimizer(func(env *CompileEnv) ast.Rewriter {
			return optimizer.InlineCalls(env.Optimize.Context, env.Symbols, env.Functions)
		})
		env.Optimize.AddOptimizer(func(env *CompileEnv) ast.Rewriter {
			return optimizer.FoldConstants(env.Symbols)
		})
		env.Optimize.AddOptimizer(func(env *CompileEnv) ast.Rewriter {
			return optimizer.InvokeBareFuncs(env.Symbols, env.Functions)
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
	env.Optimize.FastOps = make(compiler.FastOps)

	for i := range configure {
		configure[i](&env)
	}

	return env
}

type SourceLoaded func(*CompileEnv, *Source)
type Optimizer func(*CompileEnv) ast.Rewriter
type Analyzer func(*CompileEnv) ast.Visitor

type CompileEnv struct {
	Grammar   peruse.Grammar[ruleparser.Tree]
	Symbols   *symbols.Table
	Functions *ast.PartialFunctionTable
	Objects   *objects.Builder

	Optimize     Optimize
	Analysis     Analysis
	onSourceLoad []SourceLoaded
}

func (this *CompileEnv) OnSourceLoad(f SourceLoaded) {
	this.onSourceLoad = append(this.onSourceLoad, f)
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
	FastOps     compiler.FastOps
}

func (this *Optimize) AddOptimizer(o Optimizer) {
	this.Optimiziers = append(this.Optimiziers, o)
}

func (this *CompileEnv) BuildFunctionTable(declarations map[string]string) error {
	var err error
	var funcs ast.PartialFunctionTable
	funcs, err = ast.BuildCompilingFunctionTable(this.Symbols, this.Grammar, declarations)
	if err != nil {
		return err
	}
	this.Functions = &funcs
	return nil
}

func Compiler(env *CompileEnv) codegen {
	optimizers := env.Optimize.Optimiziers
	analysis := env.Analysis
	codegen := codegen{
		CompileEnv:    env,
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

type codegen struct {
	*CompileEnv
	rewriters     []ast.Rewriter
	preanalyzers  []ast.Visitor
	postanalyzers []ast.Visitor
}

func (this codegen) CompileSource(src *Source) (compiler.Bytecode, error) {
	var bytecode compiler.Bytecode
	for i := range this.onSourceLoad {
		this.onSourceLoad[i](this.CompileEnv, src)
	}

	if src.Ast == nil {
		var astErr error
		src.Ast, astErr = ast.Parse(string(src.String), this.Symbols, this.Grammar)
		if astErr != nil {
			return bytecode, fmt.Errorf("%w: %w", ErrParse, astErr)
		}
	}

	for i := range this.Analysis.pre {
		this.preanalyzers[i].Visit(src.Ast)
	}

	for range this.Optimize.Passes {
		var rewriteErr error
		src.Ast, rewriteErr = ast.RewriteWithEvery(src.Ast, this.rewriters)
		if rewriteErr != nil {
			return bytecode, fmt.Errorf("%w: %w", ErrOptimization, rewriteErr)
		}
	}

	for i := range this.Analysis.post {
		this.postanalyzers[i].Visit(src.Ast)
	}

	var compileErr error
	bytecode, compileErr = compiler.Compile(src.Ast, this.Symbols, this.Objects, this.Optimize.FastOps)
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
