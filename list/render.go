package list

import (
	// "log"

	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Styles
	highlightStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA"))

	// Components
	activeDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).Render("•")
)

func (l *List) Render() string {
	s := ""

	for i, item := range l.filteredItems {
		if i < l.first || i >= l.first+l.height {
			continue
		}

		s += l.renderItem(i, item)
	}

	return s
}

func (l *List) renderItem(i int, item string) string {
	selected := " "

	if i == l.selected {
		selected = activeDot
	}

	return selected + " " + l.highlight(item) + "\n"
}

func (l *List) highlight(item string) string {
	s := ""

	for _, v := range item {
		vString := string(v)
		if strings.Contains(l.filter, vString) {
			s += highlightStyle.Render(vString)
		} else {
			s += vString
		}
	}

	return s
}
