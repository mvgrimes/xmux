package list

// import "log"

func (l *List) Render(filter string) string {
	filteredItems := l.filter(filter)
	s := ""

	for i, item := range filteredItems {
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

// str := fmt.Sprintf("%d. %s", index+1, i.Title())
//
// fn := itemStyle.Render
// if index == m.Index() {
// 	fn = func(s string) string {
// 		return selectedItemStyle.Render("> " + s)
// 	}
// }
//
// fmt.Fprint(w, fn(str))
