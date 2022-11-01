package list

import (
	// "log"

	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	highlightStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA"))
)

func (l *List) Render() string {
	s := ""

	for i, item := range l.filteredItems {
		if i < l.first || i >= l.first+l.height {
			continue
		}

		selected := " "
		if i == l.selected {
			selected = l.activeDot
		}
		s += selected + " " + highlight(item, l.filter) + "\n"
	}

	return s
}

func highlight(item, filter string) string {
	s := ""
	for _, v := range item {
		vString := string(v)
		if strings.Contains(filter, vString) {
			s += highlightStyle.Render(vString)
		} else {
			s += vString
		}
	}

	return s
}
