package bar

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/mvgrimes/xmux/internal/state"
)

// NewCommand returns the cobra command for `xmux bar`.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:          "bar",
		Short:        "Run the sidebar service monitor",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run()
		},
	}
}

func run() error {
	out, err := exec.Command("tmux", "display-message", "-p", "#S").Output()
	if err != nil {
		return fmt.Errorf("must be run inside a tmux session: %w", err)
	}
	session := strings.TrimSpace(string(out))

	p := tea.NewProgram(newModel(session))
	return p.Start()
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
		case "ctrl+c", "q":
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
		}
	}
	return m, nil
}

func openPopup(session string, svc state.Status) tea.Cmd {
	return func() tea.Msg {
		logFile := state.LogFile(session, svc.Name)
		exec.Command("tmux", "popup", "-E",
			"-w", "80%", "-h", "80%",
			"-T", " "+svc.Name,
			fmt.Sprintf("less +G '%s'", logFile),
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
