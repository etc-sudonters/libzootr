package explore

import (
	"sudonters/libzootr/mido"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func New(codegen *mido.CodeGen) Model {
	var m Model
	m.codegen = codegen
	m.sidebar = newSidebar()
	m.tabs = newTabs(nil, codegen)
	m.focusTabs()

	if len(m.tabs.tabs) == 0 {
		panic("no tabs instantiated")
	}
	return m
}

type mainFocus int

const (
	focusSidebar = 1
	focusTabs    = 2
)

type Model struct {
	spheres []NamedSphere
	codegen *mido.CodeGen

	focused mainFocus
	sidebar sidebar
	tabs    tabs

	size        tea.WindowSizeMsg
	sidebarSize tea.WindowSizeMsg
	tabsSize    tea.WindowSizeMsg
}

func (this Model) Init() tea.Cmd {
	return tea.Batch(this.sidebar.Init(), this.tabs.Init())
}

func (this Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		this.setSize(msg)
		cmd := this.resizeChildren()
		return this, tea.Batch(cmd)
	case SphereExplored:
		this.spheres = append(this.spheres, msg.Sphere)
		cmd := this.sidebar.pushSphere(msg.Sphere)
		return this, cmd
	case sphereSelected:
		this.tabs = newTabs(&this.spheres[int(msg)], this.codegen)
		this.focusTab(TAB_INVENTORY)
		cmd := this.resizeChildren()
		return this, cmd
	case RuleDisassembled:
		var cmd tea.Cmd
		this.tabs, cmd = this.tabs.Update(msg)
		this.focusTab(TAB_DISASSEMBLY)
		return this, cmd
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyShiftTab:
			this.swapFocus()
			return this, nil
		}
	}

	var cmd tea.Cmd
	switch this.focused {
	case focusSidebar:
		this.sidebar, cmd = this.sidebar.Update(msg)
	case focusTabs:
		this.tabs, cmd = this.tabs.Update(msg)
	}
	return this, cmd
}

func (this Model) View() string {
	sidebarStyle := sectionStyle.Width(this.sidebarSize.Width).Border(lipgloss.NormalBorder())
	tabsStyle := sectionStyle

	if this.focused == focusSidebar {
		sidebarStyle = sidebarStyle.BorderForeground(activeColor)
		tabsStyle = tabsStyle.BorderForeground(inactiveColor)
	} else {
		sidebarStyle = sidebarStyle.BorderForeground(inactiveColor)
		tabsStyle = tabsStyle.BorderForeground(activeColor)
	}

	sidebar := sidebarStyle.Render(this.sidebar.View())
	tabs := tabsStyle.Render(this.tabs.View())
	return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, tabs)
}

func (this *Model) focusSidebar() {
	this.focused = focusSidebar
}

func (this *Model) focusTab(tab int) {
	this.focusTabs()
	this.tabs.curr = tab
}

func (this *Model) focusTabs() {
	this.focused = focusTabs
}

func (this *Model) swapFocus() {
	if this.focused == focusSidebar {
		this.focusTabs()
	} else {
		this.focusSidebar()
	}
}

func (this *Model) setSize(size tea.WindowSizeMsg) {
	this.size.Width = min(800, size.Width)
	this.size.Height = min(600, size.Height)
}

func (this *Model) resizeChildren() tea.Cmd {
	cmds := []tea.Cmd{nil, nil}
	this.sidebarSize = tea.WindowSizeMsg{
		Width:  30,
		Height: this.size.Height - 5,
	}
	this.tabsSize = tea.WindowSizeMsg{
		Width:  this.size.Width - 40,
		Height: this.size.Height - 5,
	}
	this.sidebar, cmds[0] = this.sidebar.Update(this.sidebarSize)
	this.tabs, cmds[1] = this.tabs.Update(this.tabsSize)
	return tea.Batch(cmds...)
}

func listDefaults(l *list.Model) {
	l.DisableQuitKeybindings()
	l.SetFilteringEnabled(false)
	l.SetShowFilter(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.SetShowStatusBar(false)
	l.SetShowTitle(false)
	l.Title = ""
}
