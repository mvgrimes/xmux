package list

import (
	// "log"

	"github.com/charmbracelet/lipgloss"

	"github.com/mvgrimes/xmux/internal/utils"
)

type List struct {
	title         string
	filter        string
	items         []string
	filteredItems []string
	selected      int
	height        int
	first         int
}

var (
	// Styles
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		PaddingLeft(2).
		PaddingRight(2)
)

func New(title string) List {
	return List{title: title, items: make([]string, 0), height: 20}
}

func (l *List) Title() string {
	return titleStyle.Render(l.title)
}

func (l *List) SetTitle(title string) *List {
	l.title = title
	return l
}

func (l *List) SetItems(items []string) *List {
	l.items = items
	l.Filter()
	return l
}

func (l *List) SetFilter(filter string) *List {
	l.filter = filter
	l.Filter()
	return l
}

func (l *List) FilteredItemsCount() int {
	return len(l.filteredItems)
}

func (l *List) Selected() string {
	// log.Printf("selected: %d -> %s", l.selected, l.filteredItems[l.selected])
	// log.Printf("selected: %v", l)
	return l.filteredItems[l.selected]
}

func (l *List) SetSelected(i int) *List {
	if i >= 0 && i < len(l.filteredItems) {
		l.selected = i
	}
	return l
}

func (l *List) SetHeight(h int) *List {
	l.height = h
	return l
}

func (l *List) Next() *List {
	if l.selected == len(l.filteredItems)-1 {
		l.selected = 0
		l.first = 0
	} else {
		l.selected++
		if l.selected >= l.first+l.height {
			l.first++
		}
	}
	return l
}

func (l *List) Prev() *List {
	if l.selected == 0 {
		l.selected = len(l.filteredItems) - 1
		l.first = utils.Max(0, l.selected-l.height+1)
	} else {
		l.selected--
		if l.selected < l.first {
			l.first = l.selected
		}
	}
	return l
}

func (l *List) PageDown() *List {
	nextPage := l.selected + l.height
	lastPage := l.FilteredItemsCount() - l.height

	l.first = utils.Min(nextPage, lastPage)
	l.selected = l.first

	return l
}

func (l *List) PageUp() *List {
	priorPage := l.selected - l.height

	l.first = utils.Max(priorPage, 0)
	l.selected = l.first

	return l
}
