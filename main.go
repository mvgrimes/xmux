package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"mccwk.com/xmux/list"
	"mccwk.com/xmux/sessions"
	"mccwk.com/xmux/utils"
)

type stage int

const (
	liveSession stage = iota
	inactiveSession
	remoteSession
)

/* STYLING */
var (
	focusedStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))
	helpStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			PaddingLeft(2).
			PaddingRight(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	docStyle          = lipgloss.NewStyle().Margin(1, 2)
)

var (
	activeDot   = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).Render("•")
	inactiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"}).Render("•")
)

/* MAIN MODEL */

type Model struct {
	focused  stage
	lists    []list.List
	filter   string
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
		list.New("Active Session", activeDot),
		list.New("Inactive Session", activeDot),
		list.New("Remote Session", activeDot),
	}
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
		for i, _ := range m.lists {
			m.lists[i].SetHeight(utils.Max(3, msg.Height-4))
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
			m.lists[m.focused].SetFilter(m.filter)
			m.lists[m.focused].Filter()
		case "right", "tab":
			m.NextList()
			m.lists[m.focused].SetFilter(m.filter)
			m.lists[m.focused].Filter()
		case "down":
			m.lists[m.focused].Next()
		case "up":
			m.lists[m.focused].Prev()
		case "pgdown", "pgup":
			m.filter += msg.String()
		case "enter":
			m.quitting = true
			m.choice = m.lists[m.focused].Selected()
			log.Printf("focused: %d", m.focused)
			executeTmux(m.focused, m.choice)
			return m, tea.Quit
		case "backspace", "delete":
			end := utils.Max(0, len(m.filter)-1)
			m.filter = m.filter[0:end]
			m.lists[m.focused].SetFilter(m.filter)
			m.lists[m.focused].Filter()

		default:
			m.filter += msg.String()
			m.lists[m.focused].SetSelected(0)
			m.lists[m.focused].SetFilter(m.filter)
			m.lists[m.focused].Filter()
		}
	}

	// var cmd tea.Cmd
	// m.textInput, cmd = m.textInput.Update(msg)
	// m.lists[m.focused], cmd = m.lists[m.focused].Update(msg)

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

	s := titleStyle.Render(m.lists[m.focused].Title())
	s += "\n"

	s += "> " + m.filter + "\n"
	s += "\n"

	s += m.lists[m.focused].Render()
	return s
}

func main() {
	model := New()
	model.listInit()

	if len(os.Getenv("DEBUG")) > 0 {
		fileName := fmt.Sprintf("%s/%s", utils.GetHomeDir(), "xmux.log")
		f, err := tea.LogToFile(fileName, "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
