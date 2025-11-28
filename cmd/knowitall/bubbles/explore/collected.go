package explore

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func newCollected(sphere *NamedSphere) collected {
	var c collected
	var sphereItems []list.Item
	var allItems []list.Item
	if sphere != nil {
		sphereItems = make([]list.Item, len(sphere.Tokens))
		for i := range sphere.Tokens {
			sphereItems[i] = listToken(sphere.Tokens[i])
			c.total += sphere.Tokens[i].Qty
		}

		allItems = make([]list.Item, len(sphere.AllTokens))
		for i := range sphere.AllTokens {
			allItems[i] = listToken(sphere.AllTokens[i])
		}
	}
	c.tokens.sphere = sphereItems
	c.tokens.total = allItems
	c.list = list.New(sphereItems, listTokenDelegate{}, 0, 0)
	listDefaults(&c.list)
	c.list.SetFilteringEnabled(true)
	c.list.SetShowFilter(true)
	return c
}

type collectedFocus int

const (
	focusSphere = 1
	focusAll    = 2
)

type collected struct {
	total  int
	tokens struct {
		sphere []list.Item
		total  []list.Item
	}

	list  list.Model
	focus collectedFocus
}

func (_ collected) Init() tea.Cmd { return nil }

func (this collected) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	add := func(c tea.Cmd) { cmds = append(cmds, c) }
	batch := func() tea.Cmd { return tea.Batch(cmds...) }
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		this.list.SetSize(msg.Width, msg.Height)
		return this, batch()
	case tea.KeyMsg:
		if this.list.FilterState() == list.Filtering {
			if msg.Type == tea.KeyEnter {
				var cmd tea.Cmd
				this.list, cmd = this.list.Update(msg)
				return this, cmd
			}

			break
		}
		switch msg.String() {

		case "F":
			this.focus = focusAll
			cmd := this.list.SetItems(this.tokens.total)
			this.list.ResetFilter()
			this.list.ResetSelected()
			return this, cmd
		case "P":
			this.focus = focusSphere
			cmd := this.list.SetItems(this.tokens.sphere)
			this.list.ResetFilter()
			this.list.ResetSelected()
			return this, cmd
		}
	}

	var cmd tea.Cmd
	this.list, cmd = this.list.Update(msg)
	add(cmd)

	return this, batch()
}

func (this collected) View() string {
	var view strings.Builder
	view.WriteString("INVENTORY\n")
	fmt.Fprintf(&view, "Collected: %3d\n", this.total)
	fmt.Fprint(&view, this.list.View())
	return view.String()
}

type listToken NamedToken

func (this listToken) FilterValue() string {
	return string(this.Name)
}

type listTokenDelegate struct{}

func (_ listTokenDelegate) Height() int  { return 1 }
func (_ listTokenDelegate) Spacing() int { return 0 }
func (_ listTokenDelegate) Update(tea.Msg, *list.Model) tea.Cmd {
	return nil
}
func (_ listTokenDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	token := item.(listToken)
	cursor := "  "
	if index == m.Index() {
		cursor = "> "
	}
	row := fmt.Sprintf("%3d %s", token.Qty, token.Name)
	fmt.Fprint(w, lipgloss.JoinHorizontal(lipgloss.Top, cursor, row))
}
