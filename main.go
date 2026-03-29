package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/mvgrimes/xmux/cmd/bar"
	"github.com/mvgrimes/xmux/cmd/watch"
	"github.com/mvgrimes/xmux/list"
	"github.com/mvgrimes/xmux/pager"
	"github.com/mvgrimes/xmux/sessions"
	"github.com/mvgrimes/xmux/utils"
)

type stage int

const (
	liveSession stage = iota
	inactiveSession
	remoteSession
)

const (
	headerLines = 4
	pagerLines  = 2
)

/* STYLING */

var (
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

/* MAIN MODEL */

type Model struct {
	focused  stage
	lists    []list.List
	filter   string
	height   int
	err      error
	choice   string
	quitting bool
}

func New() *Model {
	return &Model{}
}

func (m *Model) NextList() {
	if m.focused == remoteSession {
		m.focused = liveSession
	} else {
		m.focused++
	}
}

func (m *Model) PrevList() {
	if m.focused == liveSession {
		m.focused = remoteSession
	} else {
		m.focused--
	}
}

func (m *Model) listInit() {
	m.lists = []list.List{
		list.New("Active Session"),
		list.New("Inactive Session"),
		list.New("Remote Session"),
	}
}

func (m *Model) CurrentList() *list.List {
	return &m.lists[m.focused]
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		getActiveSessions,
		getInactiveSessions,
		getRemoteSessions,
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		for i := range m.lists {
			m.lists[i].SetHeight(utils.Max(3, msg.Height-headerLines-pagerLines))
		}

	case activeSessionsMsg:
		m.lists[liveSession].SetItems([]string(msg))
	case inactiveSessionsMsg:
		m.lists[inactiveSession].SetItems([]string(msg))
	case remoteSessonsMsg:
		m.lists[remoteSession].SetItems([]string(msg))

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "left", "shift+tab":
			m.PrevList()
			m.CurrentList().SetFilter(m.filter).Filter()
		case "right", "tab":
			m.NextList()
			m.CurrentList().SetFilter(m.filter).Filter()
		case "down":
			m.CurrentList().Next()
		case "up":
			m.CurrentList().Prev()
		case "pgdown":
			m.CurrentList().PageDown()
		case "pgup":
			m.CurrentList().PageUp()
		case "enter":
			m.quitting = true
			m.choice = m.CurrentList().Selected()
			executeTmux(m.focused, m.choice)
			return m, tea.Quit
		case "backspace", "delete":
			end := utils.Max(0, len(m.filter)-1)
			m.filter = m.filter[0:end]
			m.CurrentList().SetFilter(m.filter).Filter()

		default:
			m.filter += msg.String()
			m.CurrentList().
				SetSelected(0).
				SetFilter(m.filter).
				Filter()
		}
	}

	return m, nil
}

type activeSessionsMsg []string

func getActiveSessions() tea.Msg {
	s := sessions.GetActiveSessions()
	return activeSessionsMsg(s)
}

type inactiveSessionsMsg []string

func getInactiveSessions() tea.Msg {
	s := sessions.GetInactiveSessions()
	return inactiveSessionsMsg(s)
}

type remoteSessonsMsg []string

func getRemoteSessions() tea.Msg {
	s := sessions.GetRemoteSessions()
	return remoteSessonsMsg(s)
}

func (m Model) View() string {
	if m.quitting {
		log.Printf("quitting.... choice is %s", m.choice)
		return ""
	}

	s := m.CurrentList().Title()
	s += "\n"

	s += "> " + m.filter + "\n"
	s += "\n"

	s += m.CurrentList().Render()

	padding := utils.Max(0, m.height-m.CurrentList().FilteredItemsCount()-headerLines-pagerLines)
	log.Printf("padding: %v", padding)
	s += pager.Render(int(m.focused), padding)

	return s
}

func runTUI(cmd *cobra.Command, args []string) error {
	model := New()
	model.listInit()

	if len(os.Getenv("DEBUG")) > 0 {
		fileName := fmt.Sprintf("%s/%s", utils.GetHomeDir(), "xmux.log")
		f, err := tea.LogToFile(fileName, "debug")
		if err != nil {
			return fmt.Errorf("fatal: %w", err)
		}
		defer f.Close()
	} else {
		log.SetOutput(ioutil.Discard)
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	return p.Start()
}

func main() {
	root := &cobra.Command{
		Use:          "xmux",
		Short:        "tmux session launcher and service monitor",
		SilenceUsage: true,
		RunE:         runTUI,
	}
	root.AddCommand(watch.NewCommand())
	root.AddCommand(bar.NewCommand())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
