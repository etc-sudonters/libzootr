package explore

import (
	"fmt"
	"io"
	"strings"
	"sudonters/libzootr/cmd/knowitall/leaves"
	"sudonters/libzootr/magicbean"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/etc-sudonters/substrate/skelly/bitset32"
)

func newEdges(sphere *NamedSphere, age magicbean.Age) edges {
	var new edges
	var items []list.Item
	new.age = age

	if sphere != nil {
		var crossed, pended, edgeSet bitset32.Bitset
		switch age {
		case magicbean.AgeAdult:
			crossed = sphere.Adult.Edges.Crossed
			pended = sphere.Adult.Edges.Pended
			edgeSet = crossed.Union(pended)
		case magicbean.AgeChild:
			crossed = sphere.Child.Edges.Crossed
			pended = sphere.Child.Edges.Pended
			edgeSet = crossed.Union(pended)
		}
		new.crossed = crossed.Len()
		new.pended = pended.Len()
		items = make([]list.Item, 0, edgeSet.Len())
		for index := range bitset32.Iter(&edgeSet).UntilEmpty {
			items = append(items, edgeItem{
				crossed: crossed.IsSet(index),
				edge:    sphere.Edges[int(index)],
			})
		}
	}
	new.list = list.New(items, edgeItemDelegate{}, 0, 0)
	listDefaults(&new.list)
	new.list.SetShowFilter(true)
	new.list.SetFilteringEnabled(true)
	new.list.SetShowPagination(true)
	return new
}

type edges struct {
	age  magicbean.Age
	list list.Model

	crossed, pended int
}

func (_ edges) Init() tea.Cmd { return nil }

func (this edges) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	add := func(c tea.Cmd) { cmds = append(cmds, c) }
	batch := func() tea.Cmd { return tea.Batch(cmds...) }
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		this.list.SetSize(msg.Width, msg.Height)
		return this, batch()
	case tea.KeyMsg:
		if this.list.FilterState() == list.Filtering {
			break
		}
		if len(this.list.Items()) > 0 {
			if msg.Type == tea.KeyEnter {
				item := this.list.SelectedItem()
				edge := item.(edgeItem)
				return this, RequestDisassembly(edge.edge.Id)
			}
			if msg.String() == "e" {
				item := this.list.SelectedItem()
				edge := item.(edgeItem)
				cmd := tea.Batch(startEditingRule(edge.edge.Id), leaves.WriteStatusMsg("editing rule %04x", edge.edge.Id))
				return this, cmd
			}
		}
	}

	var cmd tea.Cmd
	this.list, cmd = this.list.Update(msg)
	add(cmd)

	return this, batch()
}

func (this edges) View() string {
	var view strings.Builder
	fmt.Fprintf(&view, "%s Edges\n", this.age)
	fmt.Fprintf(&view, "Crossed: %3d\tPended: %3d\n", this.crossed, this.pended)
	fmt.Fprint(&view, this.list.View())
	return view.String()
}

type edgeItem struct {
	edge    NamedEdge
	crossed bool
}

func (this edgeItem) FilterValue() string {
	return string(this.edge.Name)
}

type edgeItemDelegate struct{}

func (_ edgeItemDelegate) Height() int  { return 1 }
func (_ edgeItemDelegate) Spacing() int { return 0 }
func (_ edgeItemDelegate) Update(tea.Msg, *list.Model) tea.Cmd {
	return nil
}

func (_ edgeItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	this := item.(edgeItem)
	cursor := " "
	if index == m.Index() {
		cursor = ">"
	}
	mark := " "
	if !this.crossed {
		mark = "P"
	}
	fmt.Fprintf(w, "%s %s %06X %s", cursor, mark, this.edge.Id, this.edge.Name)
}
