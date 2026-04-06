package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// State values for a watched service.
const (
	StateStarting = "starting"
	StateRunning  = "running"
	StateActivity = "activity"
	StateAlert    = "alert"
	StateExited   = "exited"
)

// Status represents the current health of a watched service.
type Status struct {
	Name       string `json:"name"`
	Icon       string `json:"icon"`
	State      string `json:"state"`
	LastLine   string `json:"last_line"`
	AlertLine  string `json:"alert_line"`
	PID        int    `json:"pid"`
	WatcherPID int    `json:"watcher_pid"`
	TS         int64  `json:"ts"`
	ExitCode   int    `json:"exit_code"`
}

// Dir returns the state directory for a session.
// Respects XDG_STATE_HOME; defaults to ~/.local/state/xmux/<session>.
func Dir(session string) string {
	base := os.Getenv("XDG_STATE_HOME")
	if base == "" {
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".local", "state")
	}
	return filepath.Join(base, "xmux", session)
}

// LogFile returns the path to the output log for a service.
func LogFile(session, name string) string {
	return filepath.Join(Dir(session), name+".log")
}

// Write atomically writes a status file.
func Write(session string, s Status) error {
	dir := Dir(session)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	tmp := filepath.Join(dir, s.Name+".json.tmp")
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(dir, s.Name+".json"))
}

// Read reads the status for a single service.
func Read(session, name string) (Status, error) {
	data, err := os.ReadFile(filepath.Join(Dir(session), name+".json"))
	if err != nil {
		return Status{}, err
	}
	var s Status
	return s, json.Unmarshal(data, &s)
}

// ReadAll reads all service statuses for a session, sorted by name.
func ReadAll(session string) ([]Status, error) {
	entries, err := os.ReadDir(Dir(session))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var statuses []Status
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".json")
		s, err := Read(session, name)
		if err != nil {
			continue
		}
		statuses = append(statuses, s)
	}
	return statuses, nil
}
