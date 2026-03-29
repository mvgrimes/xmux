package watch

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/mvgrimes/xmux/state"
)

var alertPattern string

// NewCommand returns the cobra command for `xmux watch`.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "watch <name> <icon> -- <command...>",
		Short: "Wrap a background process and monitor its output",
		Long: `watch runs a command, tees its output to a log file, and tracks service
health in ~/.local/state/xmux/<session>/. Use -- to separate flags from the
wrapped command.

Example:
  xmux watch dev 󰎙 --alert 'error|Error|failed' -- npm run dev`,
		Args:               cobra.MinimumNArgs(3),
		DisableFlagParsing: false,
		SilenceUsage:       true,
		RunE:               run,
	}
	cmd.Flags().StringVar(&alertPattern, "alert", "", "regex pattern to trigger alert state")
	return cmd
}

type watcher struct {
	mu      sync.Mutex
	status  state.Status
	alertRe *regexp.Regexp
	session string
	timer   *time.Timer
}

func (w *watcher) onLine(line string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.status.LastLine = line
	w.status.TS = time.Now().Unix()

	// Alert is sticky — stays until service restarts.
	if w.status.State == state.StateAlert {
		_ = state.Write(w.session, w.status)
		return
	}

	if w.alertRe != nil && w.alertRe.MatchString(line) {
		w.status.State = state.StateAlert
		w.status.AlertLine = line
		_ = state.Write(w.session, w.status)
		return
	}

	w.status.State = state.StateActivity

	// Reset quiet timer: activity → running after 3s of silence.
	if w.timer != nil {
		w.timer.Stop()
	}
	w.timer = time.AfterFunc(3*time.Second, w.onQuiet)

	_ = state.Write(w.session, w.status)
}

func (w *watcher) onQuiet() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.status.State == state.StateActivity {
		w.status.State = state.StateRunning
		_ = state.Write(w.session, w.status)
	}
}

func run(cmd *cobra.Command, args []string) error {
	name := args[0]
	icon := args[1]
	command := args[2:]

	// Resolve current tmux session.
	out, err := exec.Command("tmux", "display-message", "-p", "#S").Output()
	if err != nil {
		return fmt.Errorf("must be run inside a tmux session: %w", err)
	}
	session := strings.TrimSpace(string(out))

	// Compile alert regex if provided.
	var alertRe *regexp.Regexp
	if alertPattern != "" {
		alertRe, err = regexp.Compile(alertPattern)
		if err != nil {
			return fmt.Errorf("invalid alert pattern: %w", err)
		}
	}

	// Open log file (append).
	logPath := state.LogFile(session, name)
	if mkErr := os.MkdirAll(state.Dir(session), 0755); mkErr != nil {
		return mkErr
	}
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	defer logFile.Close()

	// Write initial status.
	w := &watcher{
		alertRe: alertRe,
		session: session,
		status: state.Status{
			Name:  name,
			Icon:  icon,
			State: state.StateStarting,
			TS:    time.Now().Unix(),
		},
	}
	_ = state.Write(session, w.status)

	// Start the wrapped command.
	c := exec.Command(command[0], command[1:]...)
	stdoutPipe, err := c.StdoutPipe()
	if err != nil {
		return err
	}
	stderrPipe, err := c.StderrPipe()
	if err != nil {
		return err
	}

	if err := c.Start(); err != nil {
		return fmt.Errorf("start command: %w", err)
	}

	w.mu.Lock()
	w.status.PID = c.Process.Pid
	w.status.State = state.StateRunning
	_ = state.Write(session, w.status)
	w.mu.Unlock()

	// Scan both pipes, tee to stdout/log, and notify the watcher.
	var wg sync.WaitGroup
	scan := func(r io.Reader, out io.Writer) {
		defer wg.Done()
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Fprintln(out, line)
			fmt.Fprintln(logFile, line)
			w.onLine(line)
		}
	}

	wg.Add(2)
	go scan(stdoutPipe, os.Stdout)
	go scan(stderrPipe, os.Stderr)
	wg.Wait()

	exitCode := 0
	if err := c.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	w.mu.Lock()
	if w.timer != nil {
		w.timer.Stop()
	}
	w.status.State = state.StateExited
	w.status.ExitCode = exitCode
	w.status.TS = time.Now().Unix()
	_ = state.Write(session, w.status)
	w.mu.Unlock()

	return nil
}
