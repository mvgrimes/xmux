package switchcmd

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTmuxSessionTarget(t *testing.T) {
	tests := []struct {
		name    string
		session string
		want    string
	}{
		{
			name:    "plain session",
			session: "xen-bal01",
			want:    "xen-bal01:",
		},
		{
			name:    "dotted session",
			session: "xen-bal01.neb.mccwk.com",
			want:    "xen-bal01.neb.mccwk.com:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tmuxSessionTarget(tt.session); got != tt.want {
				t.Fatalf("tmuxSessionTarget(%q) = %q, want %q", tt.session, got, tt.want)
			}
		})
	}
}

func TestUpdateDeletesFilterInput(t *testing.T) {
	tests := []struct {
		name   string
		filter string
		key    tea.KeyMsg
		want   string
	}{
		{
			name:   "backspace deletes one rune",
			filter: "café",
			key:    tea.KeyMsg{Type: tea.KeyBackspace},
			want:   "caf",
		},
		{
			name:   "alt backspace deletes a word",
			filter: "active session",
			key:    tea.KeyMsg{Type: tea.KeyBackspace, Alt: true},
			want:   "active ",
		},
		{
			name:   "ctrl h deletes a word",
			filter: "active session",
			key:    tea.KeyMsg{Type: tea.KeyCtrlH},
			want:   "active ",
		},
		{
			name:   "ctrl w deletes a word and trailing separators",
			filter: "active  session-",
			key:    tea.KeyMsg{Type: tea.KeyCtrlW},
			want:   "active  ",
		},
		{
			name:   "modified backspace handles unicode words",
			filter: "session café",
			key:    tea.KeyMsg{Type: tea.KeyBackspace, Alt: true},
			want:   "session ",
		},
		{
			name:   "modified backspace on empty input is safe",
			filter: "",
			key:    tea.KeyMsg{Type: tea.KeyBackspace, Alt: true},
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := New()
			model.listInit()
			model.filter = tt.filter

			updated, _ := model.Update(tt.key)
			got := updated.(Model).filter
			if got != tt.want {
				t.Fatalf("filter = %q, want %q", got, tt.want)
			}
		})
	}
}
