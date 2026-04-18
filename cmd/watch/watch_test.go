package watch

import (
	"bytes"
	"strings"
	"testing"
)

func TestScanOutputWritesToOutAndLog(t *testing.T) {
	in := strings.NewReader("out\nerr\n")
	var out bytes.Buffer
	var log bytes.Buffer
	var lines []string

	err := scanOutput(in, &out, &log, func(line string) {
		lines = append(lines, line)
	})
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

	err := scanOutput(in, &out, &log, func(string) {})
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
