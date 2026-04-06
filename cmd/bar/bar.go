package bar

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/mvgrimes/xmux/internal/state"
)

var spawnServices []string

// NewCommand returns the cobra command for `xmux bar`.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bar",
		Short: "Run the sidebar service monitor",
		Long: `bar runs the sidebar service monitor in the current tmux pane.

Use --spawn to launch watched services directly from the bar, eliminating the
need for a separate background window. Each --spawn value is the args you would
pass to "xmux watch":

  xmux bar \
    --spawn "dev -- npm run dev" \
    --spawn "gen --alert 'error|Error' -- npm run codegen --watch"`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(spawnServices)
		},
	}
	cmd.Flags().StringArrayVar(&spawnServices, "spawn", nil,
		`spawn a watched service; value is args for "xmux watch" (repeatable)`)
	return cmd
}

func run(spawns []string) error {
	out, err := exec.Command("tmux", "display-message", "-p", "#S").Output()
	if err != nil {
		return fmt.Errorf("must be run inside a tmux session: %w", err)
	}
	session := strings.TrimSpace(string(out))

	if err := spawnWatchers(spawns); err != nil {
		return err
	}

	p := tea.NewProgram(newModel(session))
	return p.Start()
}

// spawnWatchers launches an `xmux watch` subprocess for each spawn spec.
// Subprocess stdout/stderr are discarded — output is captured in the log file.
// Cleanup happens via ctrl+d → killAllAndQuit, which signals each WatcherPID.
func spawnWatchers(spawns []string) error {
	if len(spawns) == 0 {
		return nil
	}
	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable: %w", err)
	}
	for _, s := range spawns {
		c := exec.Command("sh", "-c", shellQuote(self)+" watch "+s)
		c.Stdout = io.Discard
		c.Stderr = io.Discard
		if err := c.Start(); err != nil {
			return fmt.Errorf("spawn %q: %w", s, err)
		}
		// Do not Wait — watcher runs for the lifetime of the session.
	}
	return nil
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

/* ── model ── */

type tickMsg time.Time
type svcsMsg []state.Status

type model struct {
	session  string
	services []state.Status
	selected int
	width    int
	height   int
}

func newModel(session string) *model {
	return &model{session: session}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(fetchServices(m.session), scheduleTick())
}

func scheduleTick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func fetchServices(session string) tea.Cmd {
	return func() tea.Msg {
		svcs, _ := state.ReadAll(session)
		return svcsMsg(svcs)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tickMsg:
		return m, tea.Batch(fetchServices(m.session), scheduleTick())

	case svcsMsg:
		m.services = []state.Status(msg)
		if m.selected >= len(m.services) && len(m.services) > 0 {
			m.selected = len(m.services) - 1
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "j", "down":
			if m.selected < len(m.services)-1 {
				m.selected++
			}
		case "k", "up":
			if m.selected > 0 {
				m.selected--
			}
		case "enter", " ":
			if len(m.services) > 0 {
				return m, openPopup(m.session, m.services[m.selected])
			}
		case "r":
			if len(m.services) > 0 {
				return m, restartService(m.session, m.services[m.selected])
			}
		case "ctrl+c":
			if len(m.services) > 0 {
				return m, killService(m.services[m.selected])
			}
		case "a":
			return m, addWatcher()
		case "ctrl+d":
			return m, killAllAndQuit(m.services)
		}
	}
	return m, nil
}

func openPopup(session string, svc state.Status) tea.Cmd {
	return func() tea.Msg {
		exec.Command("tmux", "popup", "-E",
			"-w", "80%", "-h", "80%",
			"-T", " "+svc.Name,
			"xmux", "popup", session, svc.Name,
		).Run() //nolint:errcheck
		return nil
	}
}

/* ── process helpers ── */

func killProcess(pid int, sig syscall.Signal) {
	if pid <= 0 {
		return
	}
	proc, err := os.FindProcess(pid)
	if err == nil {
		proc.Signal(sig) //nolint:errcheck
	}
}

func restartService(_ string, svc state.Status) tea.Cmd {
	return func() tea.Msg {
		killProcess(svc.WatcherPID, syscall.SIGHUP)
		return nil
	}
}

func killService(svc state.Status) tea.Cmd {
	return func() tea.Msg {
		killProcess(svc.PID, syscall.SIGTERM)
		return nil
	}
}

func killAllAndQuit(services []state.Status) tea.Cmd {
	return tea.Sequentially(
		func() tea.Msg {
			for _, svc := range services {
				killProcess(svc.WatcherPID, syscall.SIGTERM)
			}
			return nil
		},
		tea.Quit,
	)
}

func addWatcher() tea.Cmd {
	return func() tea.Msg {
		exec.Command("tmux", "popup", "-E",
			"-w", "80%", "-h", "40%",
			"-T", " add watcher",
			"bash", "-c",
			`read -ep "xmux watch " args && eval "xmux watch $args"`,
		).Run() //nolint:errcheck
		return nil
	}
}

/* ── view ── */

var stateColor = map[string]lipgloss.Color{
	state.StateStarting: lipgloss.Color("240"),
	state.StateRunning:  lipgloss.Color("#00d700"),
	state.StateActivity: lipgloss.Color("#ffaf00"),
	state.StateAlert:    lipgloss.Color("#ff0000"),
	state.StateExited:   lipgloss.Color("240"),
}

func (m model) View() string {
	if len(m.services) == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("·")
	}

	var sb strings.Builder
	for i, svc := range m.services {
		if i > 0 {
			sb.WriteString("\n\n")
		}

		color := stateColor[svc.State]
		if color == "" {
			color = lipgloss.Color("240")
		}

		style := lipgloss.NewStyle().Foreground(color)
		if svc.State == state.StateAlert {
			style = style.Bold(true)
		}

		prefix := " "
		if i == m.selected {
			prefix = "▶"
		}

		sb.WriteString(prefix + style.Render(svc.Icon) + " ")
	}

	return sb.String()
}
