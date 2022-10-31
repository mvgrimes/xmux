package list

import (
	// "log"
	"sort"
	"strings"
)

type items []item
type item struct {
	rank  int
	value string
}

func (l *List) Filter() {
	// log.Printf("Filter: %v", l)
	if l.filter == "" {
		l.filteredItems = l.items
		// log.Printf("Filter: %v", l)
		return
	}

	s := make(items, 0)

	for _, v := range l.items {
		rank := getRank(l.filter, v)
		if rank > 0 {
			s = append(s, item{rank: rank, value: v})
		}
	}

	s.Sort()
	// log.Printf("%v\n", s)
	l.filteredItems = s.strings()
}

// TODO: highlight the found characters
func getRank(pattern string, text string) int {
	rank := 0
	pattern = strings.ToLower(pattern)
	t := strings.Split(strings.ToLower(text), "")

	for _, wanted := range pattern {
		found := false
		for i, char := range t {
			if char == string(wanted) {
				rank++
				t[i] = "" // looking for "xx" should only rank 2 iff there are 2 x's
				found = true
				break
			}
		}
		if !found { // if any char not found, remove
			rank = 0
			break
		}
	}

	return rank
}

func (l *items) Sort() {
	ls := *l
	sort.Slice(ls, func(i, j int) bool {
		// sort by rank
		if ls[j].rank < ls[i].rank {
			return true
		}
		if ls[j].rank > ls[i].rank {
			return false
		}
		// then length
		return len(ls[j].value) > len(ls[i].value)
	})
}

func (i *item) string() string {
	return i.value
}

func (l *items) strings() []string {
	s := make([]string, 0)

	for _, i := range *l {
		if i.rank > 0 {
			s = append(s, i.string())
		}
	}

	return s
}
