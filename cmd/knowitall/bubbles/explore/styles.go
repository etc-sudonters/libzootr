package explore

import "github.com/charmbracelet/lipgloss"

var (
	activeColor   = lipgloss.Color("33")
	inactiveColor = lipgloss.Color("8")
	latestColor   = lipgloss.Color("202")
	errColor      = lipgloss.Color("9")

	inactiveStyle = lipgloss.NewStyle().Foreground(inactiveColor)
	currentStyle  = lipgloss.NewStyle().Foreground(activeColor)
	latestStyle   = lipgloss.NewStyle().Foreground(latestColor)
	errStyle      = lipgloss.NewStyle().Foreground(errColor)

	tabStyle = lipgloss.NewStyle().PaddingLeft(1).PaddingRight(1)

	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	docStyle          = lipgloss.NewStyle().Padding(1, 2, 1, 2)
	windowStyle       = lipgloss.NewStyle().Padding(2, 0).Border(lipgloss.NormalBorder()).UnsetBorderTop()
	contentStyle      = lipgloss.NewStyle().Padding(0, 2)

	sectionStyle = lipgloss.NewStyle().Margin(1)
)

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}
