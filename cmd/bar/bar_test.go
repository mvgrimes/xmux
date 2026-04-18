package bar

import (
	"reflect"
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
