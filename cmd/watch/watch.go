package watch

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode"

	"github.com/spf13/cobra"

	"github.com/mvgrimes/xmux/internal/state"
)

var alertPattern string

const maxScanTokenSize = 1024 * 1024

// NewCommand returns the cobra command for `xmux watch`.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "watch <name> [icon] -- <command...>",
		Short: "Wrap a background process and monitor its output",
		Long: `watch runs a command, tees its output to a log file, and tracks service
health in ~/.local/state/xmux/<session>/. Use -- to separate flags from the
wrapped command.

Example:
  xmux watch dev 󰎙 --alert 'error|Error|failed' -- npm run dev
  xmux watch api -- node server.js`,
		Args:               cobra.MinimumNArgs(2),
		DisableFlagParsing: false,
		SilenceUsage:       true,
		RunE:               run,
	}
	cmd.Flags().StringVar(&alertPattern, "alert", "", "regex pattern to trigger alert state")
	return cmd
}

// defaultIconForName returns a nerd font icon based on service name or command name.
func defaultIconForName(name, cmd string) string {
	check := strings.ToLower(name) + " " + strings.ToLower(filepath.Base(cmd))

	contains := func(keys ...string) bool {
		for _, k := range keys {
			if strings.Contains(check, k) {
				return true
			}
		}
		return false
	}

	switch {
	case contains("node", "npm", "npx", "bun", "deno"):
		return "\ue74e"
	case contains("python", "python3"):
		return "\ue73c"
	case contains("go"):
		return "\ue724"
	case contains("cargo", "rust"):
		return "\ue7a8"
	case contains("docker"):
		return "\uf308"
	case contains("redis"):
		return "\ue76d"
	case contains("postgres", "psql", "pg"):
		return "\uf1c0"
	case contains("mysql", "mariadb"):
		return "\uf1c0"
	case contains("nginx", "apache", "caddy"):
		return "\uf233"
	case contains("ruby", "rails"):
		return "\ue739"
	case contains("java", "mvn", "gradle"):
		return "\ue738"
	case contains("php"):
		return "\ue73d"
	case contains("vite", "vue", "nuxt"):
		return "\ue6a0"
	case contains("react", "next", "remix"):
		return "\ue7ba"
	default:
		return "\uf120"
	}
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

	// Detect if args[1] is an icon (non-ASCII first rune) or start of command.
	var icon string
	var command []string
	firstRune := []rune(args[1])[0]
	if firstRune > unicode.MaxASCII {
		icon = args[1]
		command = args[2:]
	} else {
		command = args[1:]
	}

	if icon == "" {
		icon = defaultIconForName(name, command[0])
	}

	// Resolve current tmux session.
	session, err := resolveSession()
	if err != nil {
		return err
	}

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

	// Write initial status with WatcherPID.
	w := &watcher{
		alertRe: alertRe,
		session: session,
		status: state.Status{
			Name:       name,
			Icon:       icon,
			State:      state.StateStarting,
			TS:         time.Now().Unix(),
			WatcherPID: os.Getpid(),
		},
	}
	_ = state.Write(session, w.status)

	// Set up SIGHUP handler for restart support.
	sighupCh := make(chan os.Signal, 1)
	signal.Notify(sighupCh, syscall.SIGHUP)

	// Restart loop: on SIGHUP restart the subprocess.
	for {
		// Start the wrapped command.
		c := exec.Command(command[0], command[1:]...)
		stdoutPipe, err := c.StdoutPipe()
		if err != nil {
			fmt.Fprintf(logFile, "[xmux watch] stdout pipe error: %v\n", err)
			return err
		}
		stderrPipe, err := c.StderrPipe()
		if err != nil {
			fmt.Fprintf(logFile, "[xmux watch] stderr pipe error: %v\n", err)
			return err
		}

		if err := c.Start(); err != nil {
			msg := fmt.Sprintf("[xmux watch] start command error: %v", err)
			fmt.Fprintln(logFile, msg)
			w.mu.Lock()
			w.status.State = state.StateExited
			w.status.LastLine = msg
			w.status.ExitCode = 127
			w.status.TS = time.Now().Unix()
			_ = state.Write(session, w.status)
			w.mu.Unlock()
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
			if err := scanOutput(r, out, logFile, w.onLine); err != nil {
				fmt.Fprintf(logFile, "[xmux watch] output scan error: %v\n", err)
			}
		}

		wg.Add(2)
		go scan(stdoutPipe, os.Stdout)
		go scan(stderrPipe, os.Stderr)

		// Wait for process exit in a goroutine.
		done := make(chan error, 1)
		go func() {
			wg.Wait()
			done <- c.Wait()
		}()

		// Wait for either process exit or SIGHUP.
		select {
		case <-sighupCh:
			// Restart: kill subprocess, reset state, continue loop.
			c.Process.Kill() //nolint:errcheck
			<-done           // drain done channel
			w.mu.Lock()
			if w.timer != nil {
				w.timer.Stop()
			}
			w.status.State = state.StateStarting
			w.status.AlertLine = ""
			_ = state.Write(session, w.status)
			w.mu.Unlock()
			continue

		case err := <-done:
			// Normal exit.
			exitCode := 0
			if err != nil {
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
	}
}

func scanOutput(r io.Reader, out io.Writer, log io.Writer, onLine func(string)) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), maxScanTokenSize)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprintln(out, line)
		fmt.Fprintln(log, line)
		onLine(line)
	}
	return scanner.Err()
}

func resolveSession() (string, error) {
	if session := strings.TrimSpace(os.Getenv("XMUX_SESSION")); session != "" {
		return session, nil
	}

	if pane := strings.TrimSpace(os.Getenv("TMUX_PANE")); pane != "" {
		out, err := exec.Command("tmux", "display-message", "-p", "-t", pane, "#S").Output()
		if err == nil {
			return strings.TrimSpace(string(out)), nil
		}
	}

	out, err := exec.Command("tmux", "display-message", "-p", "#S").Output()
	if err != nil {
		return "", fmt.Errorf("must be run inside a tmux session: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
