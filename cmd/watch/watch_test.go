package watch

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

func TestRotatingLogWriterRotatesAndRetainsFiles(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "svc.log")
	w, err := newRotatingLogWriter(logPath, 16, 3, 0)
	if err != nil {
		t.Fatalf("newRotatingLogWriter: %v", err)
	}
	t.Cleanup(func() {
		_ = w.Close()
	})

	for i := 1; i <= 8; i++ {
		if _, err := fmt.Fprintf(w, "line-%d\n", i); err != nil {
			t.Fatalf("write %d: %v", i, err)
		}
	}

	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	mustContain := func(path, want string) {
		t.Helper()
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if !strings.Contains(string(data), want) {
			t.Fatalf("%s does not contain %q; got %q", path, want, string(data))
		}
	}

	mustContain(logPath, "line-7")
	mustContain(logPath, "line-8")
	mustContain(logPath+".1", "line-5")
	mustContain(logPath+".2", "line-3")
	mustContain(logPath+".3", "line-1")

	if _, err := os.Stat(logPath + ".4"); !os.IsNotExist(err) {
		t.Fatalf("expected no .4 file, stat err: %v", err)
	}
}

func TestRotatingLogWriterSingleLineLargerThanCap(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "svc.log")
	w, err := newRotatingLogWriter(logPath, 4, 2, 0)
	if err != nil {
		t.Fatalf("newRotatingLogWriter: %v", err)
	}

	if _, err := w.Write([]byte("abcdef\n")); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read active log: %v", err)
	}
	if got, want := string(data), "abcdef\n"; got != want {
		t.Fatalf("active log mismatch got %q want %q", got, want)
	}
}

func TestRotatingLogWriterDoesNotRotateAtExactCap(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "svc.log")
	w, err := newRotatingLogWriter(logPath, 7, 2, 0)
	if err != nil {
		t.Fatalf("newRotatingLogWriter: %v", err)
	}

	if _, err := w.Write([]byte("abc\n")); err != nil {
		t.Fatalf("first write: %v", err)
	}
	if _, err := w.Write([]byte("de\n")); err != nil {
		t.Fatalf("second write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	if _, err := os.Stat(logPath + ".1"); !os.IsNotExist(err) {
		t.Fatalf("expected no rotation at exact cap, stat err=%v", err)
	}
}

func TestRotatingLogWriterPrunesAgedRotatedLogs(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "svc.log")
	if err := os.WriteFile(logPath+".2", []byte("old"), 0644); err != nil {
		t.Fatalf("seed rotated log: %v", err)
	}
	old := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(logPath+".2", old, old); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	w, err := newRotatingLogWriter(logPath, 4, 3, time.Hour)
	if err != nil {
		t.Fatalf("newRotatingLogWriter: %v", err)
	}

	if _, err := w.Write([]byte("abcd\n")); err != nil {
		t.Fatalf("write trigger: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	if _, err := os.Stat(logPath + ".2"); !os.IsNotExist(err) {
		t.Fatalf("expected aged rotated log to be pruned, stat err=%v", err)
	}
}

func TestRunRejectsInvalidLogFlags(t *testing.T) {
	t.Setenv("XMUX_SESSION", "watch-invalid-flags")
	oldBytes, oldFiles, oldAge := logMaxBytes, logMaxFiles, logMaxAge
	t.Cleanup(func() {
		logMaxBytes, logMaxFiles, logMaxAge = oldBytes, oldFiles, oldAge
	})

	logMaxBytes = 0
	logMaxFiles = 3
	logMaxAge = 0
	if err := run(nil, []string{"svc", "echo", "ok"}); err == nil || !strings.Contains(err.Error(), "--log-max-bytes") {
		t.Fatalf("expected log-max-bytes error, got %v", err)
	}

	logMaxBytes = 1024
	logMaxFiles = 0
	if err := run(nil, []string{"svc", "echo", "ok"}); err == nil || !strings.Contains(err.Error(), "--log-max-files") {
		t.Fatalf("expected log-max-files error, got %v", err)
	}

	logMaxFiles = 3
	logMaxAge = -time.Second
	if err := run(nil, []string{"svc", "echo", "ok"}); err == nil || !strings.Contains(err.Error(), "--log-max-age") {
		t.Fatalf("expected log-max-age error, got %v", err)
	}
}
