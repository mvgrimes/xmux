package list

import (
	// "log"
	"mccwk.com/xmux/utils"
)

type List struct {
	title         string
	filter        string
	items         []string
	filteredItems []string
	selected      int
	activeDot     string
	height        int
	first         int
}

func New(title string, activeDot string) List {
	if activeDot == "" {
		activeDot = ">"
	}
	return List{title: title, items: make([]string, 0), activeDot: activeDot, height: 20}
}

func (l *List) Title() string {
	return l.title
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

func (l *List) SetActiveDot(a string) *List {
	l.activeDot = a
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
