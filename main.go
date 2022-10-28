package main

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type stage int

const (
	liveSession stage = iota
	inactiveSession
	remoteSession
)

/* MODEL MANAGEMENT */
var model tea.Model

/* STYLING */
var (
	focusedStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	// docStyle          = lipgloss.NewStyle().Margin(1, 2)
)

// Session is our custom Item

type Session struct {
	stage stage
	title string
}

func NewSession(stage stage, title string) Session {
	return Session{stage: stage, title: title}
}

// implement the list.Item interface
func (s Session) FilterValue() string {
	return s.title
}

func (s Session) Title() string {
	return s.title
}

func (s Session) Description() string {
	return ""
}

// Implement the itemDelegate  -------------

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Session)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.Title())

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s string) string {
			return selectedItemStyle.Render("> " + s)
		}
	}

	fmt.Fprint(w, fn(str))
}

// Pager

// p.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).Render("•")
// p.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"}).Render("•")

/* MAIN MODEL */

type Model struct {
	loaded    bool
	focused   stage
	lists     []list.Model
	textInput textinput.Model
	err       error
	quitting  bool
}

func New() *Model {
	return &Model{}
}

func (m *Model) Next() {
	if m.focused == remoteSession {
		m.focused = liveSession
	} else {
		m.focused++
	}
}

func (m *Model) Prev() {
	if m.focused == liveSession {
		m.focused = remoteSession
	} else {
		m.focused--
	}
}

func (m *Model) initLists(width, height int) {
	defaultList := list.New([]list.Item{}, itemDelegate{}, width, height)
	// defaultList.SetShowHelp(false)
	defaultList.SetShowStatusBar(false)
	defaultList.SetFilteringEnabled(false)
	m.lists = []list.Model{defaultList, defaultList, defaultList}

	// Init To Do
	m.lists[liveSession].Title = "Active Sessions"
	m.lists[liveSession].SetItems([]list.Item{
		Session{stage: liveSession, title: "buy milk"},
		Session{stage: liveSession, title: "eat sushi"},
		Session{stage: liveSession, title: "fold laundry"},
	})
	// Init in progress
	m.lists[inactiveSession].Title = "Inactive Sessions"
	m.lists[inactiveSession].SetItems([]list.Item{
		Session{stage: inactiveSession, title: "write code"},
	})
	// Init done
	m.lists[remoteSession].Title = "Remote Sessions"
	m.lists[remoteSession].SetItems([]list.Item{
		Session{stage: remoteSession, title: "stay cool"},
	})
}

func (m *Model) initTextInput() {
	ti := textinput.New()
	ti.Placeholder = "..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 30

	m.textInput = ti
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.loaded {
			m.initTextInput()
			m.initLists(msg.Width, msg.Height)
			m.loaded = true
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "left", "shift+tab":
			m.Prev()
		case "right", "tab":
			m.Next()
		case "enter":
			m.quitting = true
			return m, tea.Quit
		}
		// case Session:
		// 	task := msg
		// 	return m, m.lists[task.status].InsertItem(len(m.lists[task.status].Items()), task)
	}

	var cmd tea.Cmd
	// m.textInput, cmd = m.textInput.Update(msg)
	m.lists[m.focused], cmd = m.lists[m.focused].Update(msg)

	return m, cmd
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	if m.loaded {
		switch m.focused {
		case inactiveSession:
			return lipgloss.JoinVertical(
				lipgloss.Top,
				m.textInput.View(),
				m.lists[inactiveSession].View(),
				// return docStyle.Render(m.lists[inactiveSession].View())
			)
		case remoteSession:
			return lipgloss.JoinVertical(
				lipgloss.Top,
				m.textInput.View(),
				m.lists[remoteSession].View(),
				// return docStyle.Render(m.lists[remoteSession].View())
			)
		default:
			return lipgloss.JoinVertical(
				lipgloss.Top,
				m.textInput.View(),
				m.lists[liveSession].View(),
				// return docStyle.Render(m.lists[liveSession].View())
			)
		}
	} else {
		return "loading..."
	}
}

func main() {
	model = New()
	p := tea.NewProgram(model, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
