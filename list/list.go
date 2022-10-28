package list

import "log"

type List struct {
	title     string
	items     []string
	selected  int
	activeDot string
	height    int
	first     int
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

func (l *List) SetTitle(title string) {
	l.title = title
}

func (l *List) Items() []string {
	return l.items
}

func (l *List) SetItems(items []string) {
	l.items = items
}

func (l *List) AddItem(title string) {
	l.items = append(l.items, title)
}

func (l *List) Selected() string {
	log.Printf("selected: %d -> %s", l.selected, l.items[l.selected])
	return l.items[l.selected]
}

func (l *List) SetSelected(i int) {
	if i >= 0 && i < len(l.items) {
		l.selected = i
	}
}

func (l *List) SetActiveDot(a string) {
	l.activeDot = a
}

func (l *List) SetHeight(h int) {
	l.height = h
}

func (l *List) Next() {
	if l.selected == len(l.items)-1 {
		l.selected = 0
		l.first = 0
	} else {
		l.selected++
		if l.selected >= l.first+l.height {
			l.first++
		}
	}
}

func (l *List) Prev() {
	if l.selected == 0 {
		l.selected = len(l.items) - 1
		l.first = max(0, l.selected-l.height+1)
	} else {
		l.selected--
		if l.selected < l.first {
			l.first = l.selected
		}
	}

}

func max(a, b int) int {
	if a >= b {
		return a
	}
	return b
}
