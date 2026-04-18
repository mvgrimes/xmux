package bar

import (
	"reflect"
	"strings"
	"testing"
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
