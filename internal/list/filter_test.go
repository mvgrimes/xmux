package list

import "testing"

func TestGetRank(t *testing.T) {
	tests := []struct {
		pattern string
		text    string
		wantPos bool // want rank > 0
	}{
		{"dev", "development", true},
		{"dev", "devops", true},
		{"abc", "xaxbxcx", true},
		{"xyz", "abc", false},
		{"", "anything", false}, // empty pattern handled by Filter(), but getRank returns 0 for ""
		{"aa", "ba", false},     // only one 'a' in text
		{"aa", "baa", true},     // two 'a's available
	}
	for _, tt := range tests {
		rank := getRank(tt.pattern, tt.text)
		got := rank > 0
		if got != tt.wantPos {
			t.Errorf("getRank(%q, %q) = %d (positive=%v), want positive=%v",
				tt.pattern, tt.text, rank, got, tt.wantPos)
		}
	}
}

func TestFilter_EmptyPattern(t *testing.T) {
	l := New("test")
	l.SetItems([]string{"alpha", "beta", "gamma"})
	l.SetFilter("")
	if l.FilteredItemsCount() != 3 {
		t.Errorf("FilteredItemsCount() = %d, want 3", l.FilteredItemsCount())
	}
}

func TestFilter_WithPattern(t *testing.T) {
	l := New("test")
	l.SetItems([]string{"development", "staging", "production"})
	l.SetFilter("dev")
	if l.FilteredItemsCount() != 1 {
		t.Errorf("FilteredItemsCount() = %d, want 1", l.FilteredItemsCount())
	}
	if l.Selected() != "development" {
		t.Errorf("Selected() = %q, want %q", l.Selected(), "development")
	}
}

func TestFilter_NoMatch(t *testing.T) {
	l := New("test")
	l.SetItems([]string{"alpha", "beta"})
	l.SetFilter("xyz")
	if l.FilteredItemsCount() != 0 {
		t.Errorf("FilteredItemsCount() = %d, want 0", l.FilteredItemsCount())
	}
}

func TestNextPrev(t *testing.T) {
	l := New("test")
	l.SetItems([]string{"a", "b", "c"})
	l.SetHeight(10)

	l.Next()
	if l.selected != 1 {
		t.Errorf("after Next(), selected = %d, want 1", l.selected)
	}

	l.Prev()
	if l.selected != 0 {
		t.Errorf("after Prev(), selected = %d, want 0", l.selected)
	}

	// Prev at beginning wraps to end
	l.Prev()
	if l.selected != 2 {
		t.Errorf("Prev() at start should wrap, selected = %d, want 2", l.selected)
	}

	// Next at end wraps to beginning
	l.Next()
	if l.selected != 0 {
		t.Errorf("Next() at end should wrap, selected = %d, want 0", l.selected)
	}
}
