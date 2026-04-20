package watch

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mvgrimes/xmux/internal/state"
)

func TestScanOutputWritesToOutAndLog(t *testing.T) {
	in := strings.NewReader("out\nerr\n")
	var out bytes.Buffer
	var log bytes.Buffer
	var lines []string

	err := scanOutput(in, &out, &log, func(line string, _ bool) {
		lines = append(lines, line)
	}, false)
	if err != nil {
		t.Fatalf("scanOutput returned error: %v", err)
	}

	if got, want := out.String(), "out\nerr\n"; got != want {
		t.Fatalf("stdout mismatch\n got: %q\nwant: %q", got, want)
	}
	if got, want := log.String(), "out\nerr\n"; got != want {
		t.Fatalf("log mismatch\n got: %q\nwant: %q", got, want)
	}
	if got, want := strings.Join(lines, ","), "out,err"; got != want {
		t.Fatalf("lines mismatch\n got: %q\nwant: %q", got, want)
	}
}

func TestScanOutputAcceptsLargeLines(t *testing.T) {
	large := strings.Repeat("x", 70*1024)
	in := strings.NewReader(large + "\n")
	var out bytes.Buffer
	var log bytes.Buffer

	err := scanOutput(in, &out, &log, func(string, bool) {}, false)
	if err != nil {
		t.Fatalf("scanOutput returned error for large line: %v", err)
	}
	if got, want := out.Len(), len(large)+1; got != want {
		t.Fatalf("stdout length mismatch\n got: %d\nwant: %d", got, want)
	}
	if got, want := log.Len(), len(large)+1; got != want {
		t.Fatalf("log length mismatch\n got: %d\nwant: %d", got, want)
	}
}

func TestWatcherOnLineTracksRecentBars(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_STATE_HOME", tmp)

	w := &watcher{
		session: "recent-bars",
		status:  state.Status{Name: "svc", State: state.StateRunning},
	}

	w.onLine("ok", false)
	w.onLine("warn", true)

	got, err := state.Read("recent-bars", "svc")
	if err != nil {
		t.Fatalf("read status: %v", err)
	}
	if got.RecentBars != "01" {
		t.Fatalf("unexpected recent bars: %q", got.RecentBars)
	}
}

func TestRunWritesStartFailureToLog(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_STATE_HOME", tmp)
	t.Setenv("XMUX_SESSION", "watch-start-fail")
	alertPattern = ""

	err := run(nil, []string{"svc", "command-that-does-not-exist"})
	if err == nil {
		t.Fatal("expected run to fail for missing command")
	}

	logPath := state.LogFile("watch-start-fail", "svc")
	data, readErr := os.ReadFile(logPath)
	if readErr != nil {
		t.Fatalf("read log file: %v", readErr)
	}
	if !strings.Contains(string(data), "start command error") {
		t.Fatalf("expected start failure in log, got: %q", string(data))
	}

	statusPath := filepath.Join(state.Dir("watch-start-fail"), "svc.json")
	statusData, readStatusErr := os.ReadFile(statusPath)
	if readStatusErr != nil {
		t.Fatalf("read status file: %v", readStatusErr)
	}
	if !strings.Contains(string(statusData), `"state":"exited"`) {
		t.Fatalf("expected exited state after start failure, got: %q", string(statusData))
	}
}
