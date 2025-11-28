package generation

import (
	"sudonters/libzootr/cmd/knowitall/bubbles/explore"
	"sudonters/libzootr/components"
	"sudonters/libzootr/magicbean"
	"sudonters/libzootr/mido/compiler"
	"sudonters/libzootr/mido/vm"
	"sudonters/libzootr/zecs"

	tea "github.com/charmbracelet/bubbletea"
)

func loadRuleSource(msg explore.LoadRuleSource, gen *magicbean.Generation) tea.Cmd {
	return func() tea.Msg {
		values, err := gen.Ocm.GetValues(zecs.Entity(msg), zecs.Get[components.RuleSource])
		src, _ := values[0].(components.RuleSource)
		return explore.EditRule{
			Err:    err,
			Source: src,
		}
	}
}

type discache map[zecs.Entity]explore.RuleDisassembled

func disassemble(gen *magicbean.Generation, edge zecs.Entity, cache discache) tea.Cmd {
	return func() tea.Msg {
		dis, exists := cache[edge]
		if exists {
			dis.Inventory = gen.Inventory
			dis.Symbols = gen.Symbols
			return dis
		}
		ocm := &gen.Ocm
		values, err := ocm.GetValues(edge,
			zecs.Get[components.Name], zecs.Get[components.RuleCompiled],
			zecs.Get[components.RuleSource], zecs.Get[components.RuleParsed],
			zecs.Get[components.RuleOptimized],
		)
		dis.Id = edge
		dis.Err = err
		if err == nil {
			name := values[0].(components.Name)
			code := values[1].(components.RuleCompiled)
			src, hasSrc := values[2].(components.RuleSource)
			disassembled := vm.Disassemble(compiler.Bytecode(code), &gen.Objects)
			dis.Name = string(name)
			dis.Disassembly = disassembled
			if hasSrc {
				dis.Source = string(src)
			}

			if node, isAst := values[3].(components.RuleParsed); isAst {
				dis.Ast = node.Node
			}

			if node, isAst := values[4].(components.RuleOptimized); isAst {
				dis.Opt = node.Node
			}

			cache[edge] = dis
		}
		// don't write inventory into cache
		dis.Inventory = gen.Inventory
		dis.Symbols = gen.Symbols
		return dis
	}
}
