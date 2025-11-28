package explore

import (
	"fmt"
	"iter"
	"strings"
	"sudonters/libzootr/cmd/knowitall/leaves"
	"sudonters/libzootr/mido"
	"sudonters/libzootr/mido/ast"
	"sudonters/libzootr/mido/optimizer"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

func newEditor(codegen *mido.CodeGen) editor {
	var e editor
	e.codegen = codegen
	e.text = textarea.New()
	e.text.SetHeight(10)
	return e
}

func stepOptimize(src string, codegen *mido.CodeGen) *stepper {
	s := new(stepper)
	steps := mido.StepOptimize(src, codegen, &s.err)
	pull, stop := iter.Pull2(steps)
	s.stop = func() {
		stop()
		s.stopped = true
	}
	var last ast.Node
	s.next = func() {
		if s.stopped {
			return
		}
		for {
			optimizer.SetCurrentLocation(codegen.Context, "Rule Editor")
			_, curr, valid := pull()
			if !valid {
				s.stop()
			}
			if s.err != nil {
				s.stop()
			}
			if curr == nil {
				s.stop()
				return
			}
			if len(s.steps) == 0 {
				last = curr
				s.steps = append(s.steps, curr)
				return
			}
			if ast.Hash(last) != ast.Hash(curr) {
				last = curr
				s.steps = append(s.steps, curr)
				return
			}
		}
	}
	return s
}

type stepper struct {
	next    func()
	stop    func()
	steps   []ast.Node
	err     error
	stopped bool
}

type editor struct {
	codegen *mido.CodeGen
	text    textarea.Model
	stepper *stepper
	writing bool
}

func (_ editor) Init() tea.Cmd { return nil }

func (this editor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		this.text.SetHeight(10)
		this.text.SetWidth(msg.Width - 5)
	case EditRule:
		if msg.Err != nil {
			return this, leaves.WriteStatusMsg("edit rule failed: %s", msg.Err.Error())
		}
		this.text.SetValue(string(msg.Source))
		if this.stepper != nil {
			this.stepper.stop()
			this.stepper = nil
		}
		return this, nil
	case tea.KeyMsg:
		if this.writing && msg.Type == tea.KeyEsc {
			cmd := this.normalMode()
			return this, cmd
		} else if this.writing {
			var cmd tea.Cmd
			this.text, cmd = this.text.Update(msg)
			return this, cmd
		} else if msg.String() == "i" {
			this.writing = true
			cmd := this.insertMode()
			return this, cmd
		} else if msg.Type == tea.KeyF5 {
			if this.stepper == nil {
				this.stepper = stepOptimize(this.text.Value(), this.codegen)
				this.stepper.next()
			} else {
				this.stepper.next()
			}
		}
	}

	return this, nil
}

func (this editor) View() string {
	var view strings.Builder
	fmt.Fprintln(&view, "EDITOR")
	view.WriteString("\n\n")
	fmt.Fprint(&view, this.text.View())
	fmt.Fprintln(&view)
	fmt.Fprintln(&view)

	if this.stepper != nil {

		for _, step := range this.stepper.steps {
			fmt.Fprint(&view, ast.Render(step, this.codegen.SymbolTable()))
			fmt.Fprintln(&view)
			fmt.Fprintln(&view)
		}
		if this.stepper.err != nil {
			fmt.Fprintln(&view, errStyle.Render(this.stepper.err.Error()))
		}
	}

	return view.String()
}

func (this *editor) insertMode() tea.Cmd {
	if this.stepper != nil {
		this.stepper.stop()
		this.stepper = nil
	}

	this.writing = true
	cmd := this.text.Focus()
	return tea.Batch(cmd, leaves.WriteStatusMsg("insert mode"))
}

func (this *editor) normalMode() tea.Cmd {
	this.writing = false
	this.text.Blur()
	return leaves.WriteStatusMsg("normal mode")
}
