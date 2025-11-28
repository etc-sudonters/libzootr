package generation

import (
	"sudonters/libzootr/cmd/knowitall/bubbles/explore"
	"sudonters/libzootr/cmd/knowitall/leaves"
	"sudonters/libzootr/magicbean"
	"sudonters/libzootr/magicbean/tracking"
	"sudonters/libzootr/playthrough"

	tea "github.com/charmbracelet/bubbletea"
)

func New(gen *magicbean.Generation, names tracking.NameTable, searches playthrough.Searches) Model {
	return Model{
		gen:      gen,
		search:   searches,
		names:    names,
		discache: make(discache, 32),
		explore:  explore.New(&gen.CodeGen),
	}
}

type Model struct {
	gen      *magicbean.Generation
	names    tracking.NameTable
	search   playthrough.Searches
	discache discache
	explore  explore.Model
}

func (this Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	add := func(c tea.Cmd) {
		cmds = append(cmds, c)
	}

	switch msg := msg.(type) {
	case explore.ExploreSphere:
		add(runSphere(msg, this.search, this.names, this.gen))
		return this, tea.Batch(cmds...)
	case explore.DisassembleRule:
		add(disassemble(this.gen, msg.Id, this.discache))
		return this, tea.Batch(cmds...)
	case explore.LoadRuleSource:
		return this, loadRuleSource(msg, this.gen)
	case explore.RuleDisassembled:
		if msg.Name == "" && msg.Id != 0 {
			msg.Name = string(this.names[msg.Id])
		}
		if msg.Err != nil {
			add(leaves.WriteStatusMsg("disassembly of %06x failed", msg.Id))
		}
		break
	case explore.SphereExplored:
		break
	}

	var xplrCmd tea.Cmd
	this.explore, xplrCmd = this.explore.Update(msg)
	add(xplrCmd)
	return this, tea.Batch(cmds...)
}

func (this Model) Init() tea.Cmd {
	return this.explore.Init()
}

func (this Model) View() string {
	return this.explore.View()
}
