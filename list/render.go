package list

// import "log"

func (l *List) Render() string {
	// log.Printf("%v", l)
	s := ""

	for i, item := range l.filteredItems {
		if i < l.first || i >= l.first+l.height {
			continue
		}

		selected := " "
		if i == l.selected {
			selected = l.activeDot
		}
		s += selected + " " + item + "\n"
	}

	return s
}
