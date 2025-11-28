package explore

import (
	"bufio"
	"fmt"
	"io"
	"iter"
	"regexp"
	"strings"
	"sudonters/libzootr/mido/ast"
	"sudonters/libzootr/mido/objects"
	"sudonters/libzootr/mido/vm"
	"sudonters/libzootr/zecs"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/etc-sudonters/substrate/skelly/bitset32"
)

func newDisassembly(sphere *NamedSphere) disassembly {
	var d disassembly
	d.codeDelegate = &codeLineDelegate{}
	d.constDelegate = &constItemDelegate{}
	if sphere != nil {
		d.constDelegate.sphereInventory = sphere.TokenMap
	}
	d.code = list.New(nil, d.codeDelegate, 0, 0)
	d.constants = list.New(nil, d.constDelegate, 0, 0)
	d.focus = disFocusCode
	d.sphere = sphere
	listDefaults(&d.code)
	listDefaults(&d.constants)
	return d
}

type disFocus int

const (
	disFocusCode   = 1
	disFocusConsts = 2
)

type disassembly struct {
	rule  RuleDisassembled
	size  tea.WindowSizeMsg
	focus disFocus

	codeDelegate  *codeLineDelegate
	constDelegate *constItemDelegate

	sphereInventory map[zecs.Entity]NamedToken
	sphere          *NamedSphere

	code, constants list.Model
}

func (_ disassembly) Init() tea.Cmd {
	return func() tea.Msg {
		return makeHighlighterMsg{}
	}
}

type makeHighlighterMsg struct{}

func (this disassembly) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	add := func(c tea.Cmd) { cmds = append(cmds, c) }
	batch := func() tea.Cmd { return tea.Batch(cmds...) }
	switch msg := msg.(type) {
	case makeHighlighterMsg:
		this.makeHighlighter()
	case tea.WindowSizeMsg:
		this.resize(msg)
		return this, batch()
	case RuleDisassembled:
		this.load(msg)
		this.makeHighlighter()
		return this, batch()
	case tea.KeyMsg:
		if msg.Type == tea.KeyTab {
			switch this.focus {
			case disFocusCode:
				this.focus = disFocusConsts
			default:
				this.focus = disFocusCode
			}
			this.makeHighlighter()
			return this, batch()
		}
	}

	var cmd tea.Cmd
	switch this.focus {
	case disFocusCode:
		this.code, cmd = this.code.Update(msg)
	case disFocusConsts:
		this.constants, cmd = this.constants.Update(msg)
	}
	add(cmd)
	this.makeHighlighter()
	return this, batch()
}

func (this *disassembly) makeHighlighter() {
	currConstant := this.constants.SelectedItem()
	if currConstant != nil {
		constant := vm.Constant(currConstant.(constItem))
		highlight := fmt.Sprintf("0x%04X", constant.Index)
		regex := regexp.MustCompile(highlight)
		highlighter := makeHighlighter(regex, currentStyle)
		this.codeDelegate.name = "constantHighlighter"
		this.codeDelegate.highlighter = highlighter
	} else {
		this.codeDelegate.name = "noConstantsHighlighter"
		this.codeDelegate.highlighter = noConstants
	}
}

func noConstants(s string) string {
	return s
}

func (this disassembly) View() string {
	if this.rule.Id == 0 {
		return "DISASSEMBLY"

	}
	if this.sphere == nil {
		panic("weird state")
	}
	var view strings.Builder
	fmt.Fprintf(&view, "Disassembly of %06X %q\n", this.rule.Id, this.rule.Name)

	if this.rule.Err != nil {
		// no new line
		msg := fmt.Sprintf("Err: %s", this.rule.Err)
		view.WriteString(errStyle.Render(msg))
	}
	// "extra" new line to clear err and leave a space
	view.WriteString("\n")
	fmt.Fprintf(
		&view, "Has Adult Crossed? %t\tHas Child Crossed? %t\n\n",
		bitset32.IsSet(&this.sphere.Adult.Edges.Total, this.rule.Id),
		bitset32.IsSet(&this.sphere.Child.Edges.Total, this.rule.Id),
	)
	if this.rule.Source != "" {
		fmt.Fprintln(&view, this.rule.Source)
	}
	if this.rule.Ast != nil {
		fmt.Fprintln(&view, ast.Render(this.rule.Ast, this.rule.Symbols))
	}
	if this.rule.Opt != nil {
		fmt.Fprintln(&view, ast.Render(this.rule.Opt, this.rule.Symbols))
	}

	listStyle := contentStyle.Border(lipgloss.RoundedBorder())
	activeListStyle := listStyle.BorderForeground(activeColor)
	var code, constants string

	switch this.focus {
	case disFocusCode:
		code = activeListStyle.Render(this.code.View())
		constants = listStyle.Render(this.constants.View())
	case disFocusConsts:
		code = listStyle.Render(this.code.View())
		constants = activeListStyle.Render(this.constants.View())
	}

	lists := lipgloss.JoinHorizontal(lipgloss.Top, code, constants)
	view.WriteString(lists)
	return contentStyle.Render(view.String())
}

func (this *disassembly) resize(msg tea.WindowSizeMsg) {
	this.size = msg

	listWidth := (msg.Width / 2) - 4
	this.code.SetSize(listWidth, this.size.Height-5)
	this.constants.SetSize(listWidth, this.size.Height-5)
}

func (this *disassembly) load(msg RuleDisassembled) tea.Cmd {
	var cmds []tea.Cmd
	add := func(c tea.Cmd) { cmds = append(cmds, c) }
	this.rule = msg

	codeItems := make([]list.Item, 0, len(this.rule.Disassembly.Code))
	constItems := make([]list.Item, len(this.rule.Disassembly.Constants))

	var err error
	for item := range lines(this.rule.Disassembly.Dis, &err) {
		codeItems = append(codeItems, codeLine(item))
	}

	for i, constant := range this.rule.Disassembly.Constants {
		constItems[i] = constItem(constant)
	}

	add(this.code.SetItems(codeItems))
	add(this.constants.SetItems(constItems))
	this.code.Select(0)
	this.constants.Select(0)
	return tea.Batch(cmds...)
}

func lines(str string, err *error) iter.Seq[string] {
	scanner := bufio.NewScanner(strings.NewReader(str))
	return func(yield func(string) bool) {
		for scanner.Scan() {
			if !yield(scanner.Text()) {
				break
			}
		}
		if scanErr := scanner.Err(); scanErr != nil {
			*err = scanErr
		}
	}
}

type codeLine string

func (_ codeLine) FilterValue() string { return "" }

type codeLineDelegate struct {
	highlighter func(string) string
	name        string
}

func (_ *codeLineDelegate) Height() int  { return 1 }
func (_ *codeLineDelegate) Spacing() int { return 0 }
func (_ *codeLineDelegate) Update(tea.Msg, *list.Model) tea.Cmd {
	return nil
}

func (this *codeLineDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	codeLine := item.(codeLine)
	cursor := " "
	if index == m.Index() {
		cursor = ">"
	}
	highlighter := this.highlighter
	if highlighter == nil {
		panic("no highlighter set")
	}
	fmt.Fprintf(w, "%s %s", cursor, highlighter(string(codeLine)))
}

func makeHighlighter(highlight *regexp.Regexp, style lipgloss.Style) func(string) string {
	return func(s string) string {
		return highlight.ReplaceAllStringFunc(s, func(s string) string {
			return style.Render(s)
		})
	}
}

type constItem vm.Constant

func (_ constItem) FilterValue() string { return "" }

type constItemDelegate struct {
	sphereInventory map[zecs.Entity]NamedToken
}

func (_ *constItemDelegate) Height() int  { return 5 }
func (_ *constItemDelegate) Spacing() int { return 1 }
func (this *constItemDelegate) Update(_ tea.Msg, parent *list.Model) tea.Cmd {
	return nil
}

func (this *constItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	constant := vm.Constant(item.(constItem))
	cursor := "  "
	if index == m.Index() {
		cursor = "> "
	}
	view := constant.String()
	if this.sphereInventory != nil && constant.Name != "" {
		ptr := objects.UnpackPtr32(constant.Object)
		if ptr.Tag == objects.PtrToken {
			token := this.sphereInventory[zecs.Entity(ptr.Addr)]
			view = fmt.Sprintf("%s\tIn Inventory: %03d\n", view, token.Qty)
		}
	}

	fmt.Fprint(w, lipgloss.JoinHorizontal(lipgloss.Top, cursor, view))
}
