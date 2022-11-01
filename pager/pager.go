package pager

import (
	// "log"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Styles
	pagerStyle = lipgloss.NewStyle().Margin(0, 2)

	// Components
	activeDot   = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).Render("•")
	inactiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"}).Render("•")
)

func Render(focused int, topPadding int) string {
	p := "\n"

	for i := 0; i < topPadding; i++ {
		p += "\n"
	}

	for i := 0; i < 3; i++ {
		if i == focused {
			p += activeDot
		} else {
			p += inactiveDot
		}
	}

	return pagerStyle.Render(p)
}
