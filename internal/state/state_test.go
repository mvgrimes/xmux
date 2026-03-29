package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDir(t *testing.T) {
	t.Run("uses XDG_STATE_HOME when set", func(t *testing.T) {
		t.Setenv("XDG_STATE_HOME", "/tmp/xdg")
		got := Dir("mysession")
		want := "/tmp/xdg/xmux/mysession"
		if got != want {
			t.Errorf("Dir() = %q, want %q", got, want)
		}
	})

	t.Run("defaults to ~/.local/state", func(t *testing.T) {
		t.Setenv("XDG_STATE_HOME", "")
		home, _ := os.UserHomeDir()
		got := Dir("mysession")
		want := filepath.Join(home, ".local", "state", "xmux", "mysession")
		if got != want {
			t.Errorf("Dir() = %q, want %q", got, want)
		}
	})
}

func TestLogFile(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", "/tmp/xdg")
	got := LogFile("sess", "dev")
	want := "/tmp/xdg/xmux/sess/dev.log"
	if got != want {
		t.Errorf("LogFile() = %q, want %q", got, want)
	}
}

func TestWriteAndRead(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", dir)

	s := Status{
		Name:  "dev",
		Icon:  "󰎙",
		State: StateRunning,
		PID:   42,
		TS:    1234567890,
	}

	if err := Write("testsess", s); err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	got, err := Read("testsess", "dev")
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}

	if got.Name != s.Name || got.State != s.State || got.PID != s.PID {
		t.Errorf("Read() = %+v, want %+v", got, s)
	}
}

func TestReadAll(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", dir)

	services := []Status{
		{Name: "alpha", Icon: "A", State: StateRunning},
		{Name: "beta", Icon: "B", State: StateAlert},
	}
	for _, s := range services {
		if err := Write("mysess", s); err != nil {
			t.Fatalf("Write(%q) error: %v", s.Name, err)
		}
	}

	got, err := ReadAll("mysess")
	if err != nil {
		t.Fatalf("ReadAll() error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("ReadAll() returned %d items, want 2", len(got))
	}
}

func TestReadAllMissingDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", dir)

	got, err := ReadAll("nosuchsession")
	if err != nil {
		t.Fatalf("ReadAll() on missing dir should not error, got: %v", err)
	}
	if got != nil {
		t.Errorf("ReadAll() = %v, want nil", got)
	}
}
