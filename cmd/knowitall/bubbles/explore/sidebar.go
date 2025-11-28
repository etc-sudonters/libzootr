package explore

import (
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type sidebar struct {
	list list.Model
}

func (this *sidebar) pushSphere(sphere NamedSphere) tea.Cmd {
	index := len(this.list.Items())
	insert := this.list.InsertItem(math.MaxInt, intoSidebarSphere(sphere))
	this.list.Select(index)
	return tea.Batch(insert, selectSphere(index))
}

func (this sidebar) Init() tea.Cmd {
	return nil
}

func (this sidebar) Update(msg tea.Msg) (sidebar, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		this.list.SetSize(msg.Width, msg.Height)
		return this, nil
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter {
			return this, selectSphere(this.list.Index())
		}
	}

	var cmd tea.Cmd
	this.list, cmd = this.list.Update(msg)
	return this, cmd
}

func (this sidebar) View() string {
	return this.list.View()
}

type sidebarSphere struct {
	number int
	errMsg string

	crossed, pended, new, collected int
}

func (_ sidebarSphere) FilterValue() string {
	return ""
}

type sidebarDelegate struct{}

func (_ sidebarDelegate) Height() int  { return 3 }
func (_ sidebarDelegate) Spacing() int { return 1 }

func (_ sidebarDelegate) Update(tea.Msg, *list.Model) tea.Cmd {
	return nil
}

func (_ sidebarDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	sphere := item.(sidebarSphere)
	cursor := "  "
	style := inactiveStyle
	if index == m.Index() {
		cursor = "> "
		style = currentStyle
	} else if sphere.errMsg != "" {
		style = errStyle
	} else if index == len(m.Items())-1 {
		style = latestStyle
	}

	style = style.Padding(0, 1, 0, 1)

	repr := strings.Builder{}
	fmt.Fprintf(&repr, "Sphere %3d\n", sphere.number)
	fmt.Fprintf(&repr, "C: %3d P: %3d N: %3d I: %3d\n", sphere.crossed, sphere.pended, sphere.new, sphere.collected)
	if sphere.errMsg != "" {
		fmt.Fprint(&repr, "Err:", sphere.errMsg)
	}

	view := lipgloss.JoinHorizontal(lipgloss.Top, cursor, repr.String())
	fmt.Fprint(w, style.Render(view))
}

func newSidebar() sidebar {
	l := list.New(nil, sidebarDelegate{}, 0, 0)
	listDefaults(&l)
	return sidebar{l}
}

func intoSidebarSphere(sphere NamedSphere) sidebarSphere {
	var sidebar sidebarSphere
	sidebar.number = sphere.I
	sidebar.crossed = sphere.Adult.Edges.Crossed.Len() + sphere.Child.Edges.Crossed.Len()
	sidebar.pended = sphere.Adult.Edges.Pended.Len() + sphere.Child.Edges.Pended.Len()
	sidebar.new = sphere.Adult.Nodes.Reached.Len() + sphere.Child.Nodes.Reached.Len()

	for _, tokens := range sphere.Tokens {
		sidebar.collected += tokens.Qty
	}

	if sphere.Error != nil {
		msg := sphere.Error.Error()
		if len(msg) > 15 {
			msg = msg[:15] + "..."
		}
		sidebar.errMsg = msg
	}

	return sidebar
}
