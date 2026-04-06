package popup

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/mvgrimes/xmux/internal/state"
)

// NewCommand returns the cobra command for `xmux popup`.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:          "popup <session> <name>",
		Short:        "Open a log viewer popup for a watched service",
		Args:         cobra.ExactArgs(2),
		SilenceUsage: true,
		RunE:         run,
	}
}

func run(cmd *cobra.Command, args []string) error {
	session, name := args[0], args[1]
	p := tea.NewProgram(newModel(session, name), tea.WithAltScreen())
	return p.Start()
}

/* ── messages ── */

type logLinesMsg []string
type tickMsg time.Time

/* ── model ── */

type model struct {
	session string
	name    string
	lines   []string
	width   int
	height  int
	offset  int // scroll offset from bottom
}

func newModel(session, name string) *model {
	return &model{session: session, name: name}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(loadLog(state.LogFile(m.session, m.name)), scheduleTick())
}

func scheduleTick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func loadLog(logPath string) tea.Cmd {
	return func() tea.Msg {
		f, err := os.Open(logPath)
		if err != nil {
			return logLinesMsg(nil)
		}
		defer f.Close()

		var lines []string
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		return logLinesMsg(lines)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tickMsg:
		return m, tea.Batch(loadLog(state.LogFile(m.session, m.name)), scheduleTick())

	case logLinesMsg:
		m.lines = []string(msg)

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, tea.Quit

		case "ctrl+c":
			// Send SIGTERM to the subprocess PID.
			s, err := state.Read(m.session, m.name)
			if err == nil && s.PID > 0 {
				sendSignal(s.PID, syscall.SIGTERM)
			}
			return m, tea.Quit

		case "r":
			// Send SIGHUP to the watcher PID to trigger restart.
			s, err := state.Read(m.session, m.name)
			if err == nil && s.WatcherPID > 0 {
				sendSignal(s.WatcherPID, syscall.SIGHUP)
			}
			return m, tea.Quit

		case "j", "down":
			if m.offset > 0 {
				m.offset--
			}

		case "k", "up":
			m.offset++
		}
	}
	return m, nil
}

func sendSignal(pid int, sig syscall.Signal) {
	proc, err := os.FindProcess(pid)
	if err == nil {
		proc.Signal(sig) //nolint:errcheck
	}
}

var statusBarStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("240"))

func (m model) View() string {
	if m.height == 0 {
		return ""
	}

	viewLines := m.height - 2
	if viewLines < 1 {
		viewLines = 1
	}

	// Determine the visible window from the end of lines, accounting for offset.
	total := len(m.lines)
	end := total - m.offset
	if end < 0 {
		end = 0
	}
	start := end - viewLines
	if start < 0 {
		start = 0
	}

	visible := m.lines[start:end]

	var sb strings.Builder
	for _, line := range visible {
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	// Pad remaining lines if not enough content.
	for i := len(visible); i < viewLines; i++ {
		sb.WriteString("\n")
	}

	statusBar := fmt.Sprintf(" [q]uit [r]estart [ctrl-c]kill  %d/%d", end, total)
	sb.WriteString(statusBarStyle.Render(statusBar))

	return sb.String()
}
