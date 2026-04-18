package bar

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/mvgrimes/xmux/internal/state"
)

func TestNewCommandPassesAllSpawnFlags(t *testing.T) {
	spawnServices = nil

	oldRunBar := runBar
	defer func() {
		runBar = oldRunBar
		spawnServices = nil
	}()

	var got []string
	runBar = func(spawns []string) error {
		got = append([]string(nil), spawns...)
		return nil
	}

	want := []string{
		"dev -- npm run dev",
		"gen --alert 'error|Error' -- npm run codegen --watch",
	}

	cmd := NewCommand()
	cmd.SetArgs([]string{
		"--spawn", want[0],
		"--spawn", want[1],
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("spawn flags mismatch\n got: %#v\nwant: %#v", got, want)
	}
}

func TestEnvWithSessionOverridesExistingValue(t *testing.T) {
	t.Setenv("XMUX_SESSION", "stale")

	env := envWithSession("fresh")
	count := 0
	for _, kv := range env {
		if strings.HasPrefix(kv, "XMUX_SESSION=") {
			count++
			if kv != "XMUX_SESSION=fresh" {
				t.Fatalf("unexpected session env: %q", kv)
			}
		}
	}

	if count != 1 {
		t.Fatalf("expected exactly 1 XMUX_SESSION entry, got %d", count)
	}
}

func TestStopAndWaitTerminatesSpawnedProcessGroup(t *testing.T) {
	cmd := exec.Command("sh", "-c", "sleep 60")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start sleep: %v", err)
	}

	if err := stopAndWait([]*exec.Cmd{cmd}); err != nil {
		t.Fatalf("stopAndWait: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	err := syscall.Kill(cmd.Process.Pid, 0)
	if err != syscall.ESRCH {
		t.Fatalf("expected process to exit, kill(0) err=%v", err)
	}
}

func TestCleanupSpawnedLogsRemovesOnlyMatchingWatcherLogs(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_STATE_HOME", tmp)

	session := "cleanup-spawned-logs"
	if err := state.Write(session, state.Status{Name: "dev", WatcherPID: 111}); err != nil {
		t.Fatalf("write dev state: %v", err)
	}
	if err := state.Write(session, state.Status{Name: "api", WatcherPID: 222}); err != nil {
		t.Fatalf("write api state: %v", err)
	}

	devLog := state.LogFile(session, "dev")
	apiLog := state.LogFile(session, "api")
	if err := os.WriteFile(devLog, []byte("dev logs"), 0644); err != nil {
		t.Fatalf("write dev log: %v", err)
	}
	if err := os.WriteFile(apiLog, []byte("api logs"), 0644); err != nil {
		t.Fatalf("write api log: %v", err)
	}

	cmds := []*exec.Cmd{{Process: &os.Process{Pid: 111}}}
	if err := cleanupSpawnedLogs(session, cmds); err != nil {
		t.Fatalf("cleanupSpawnedLogs: %v", err)
	}

	if _, err := os.Stat(devLog); !os.IsNotExist(err) {
		t.Fatalf("expected dev log removed, stat err=%v", err)
	}
	if _, err := os.Stat(apiLog); err != nil {
		t.Fatalf("expected api log to remain, stat err=%v", err)
	}

	if _, err := os.Stat(filepath.Join(state.Dir(session), "dev.json")); err != nil {
		t.Fatalf("expected dev state to remain, stat err=%v", err)
	}
}
