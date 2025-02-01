package install

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type InstallStep int

const (
	StepFeatures InstallStep = iota
	StepShells
)

type Option struct {
	title       string
	description string
	selected    bool
}

type Model struct {
	viewport viewport.Model
	options  []Option
	cursor   int
	done     bool
	step     InstallStep
	features map[string]bool
	shells   map[string]bool
	width    int
	height   int
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170")).
			MarginLeft(2)

	itemStyle = lipgloss.NewStyle().
			MarginLeft(4)

	selectedItemStyle = lipgloss.NewStyle().
				MarginLeft(4).
				Foreground(lipgloss.Color("170")).
				Bold(true)

	checkboxStyle = lipgloss.NewStyle().
			MarginRight(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginLeft(4)
)

func NewModel() Model {
	return Model{
		viewport: viewport.New(0, 0),
		options: []Option{
			{title: "Shell Completions", description: "Install command completions for shells", selected: true},
			{title: "Man Pages", description: "Install manual pages", selected: true},
			{title: "Shell Integration", description: "Enable directory changing functionality", selected: true},
		},
		features: make(map[string]bool),
		shells:   make(map[string]bool),
		step:     StepFeatures,
	}
}

func (m Model) View() string {
	var s strings.Builder

	title := "Select Features to Install"
	if m.step == StepShells {
		title = "Select Shells"
	}
	s.WriteString(titleStyle.Render(title) + "\n\n")

	for i, option := range m.options {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		checkbox := "[ ]"
		if option.selected {
			checkbox = "[x]"
		}

		item := fmt.Sprintf("%s %s %s\n    %s",
			cursor,
			checkboxStyle.Render(checkbox),
			option.title,
			option.description,
		)

		if i == m.cursor {
			s.WriteString(selectedItemStyle.Render(item))
		} else {
			s.WriteString(itemStyle.Render(item))
		}
		s.WriteString("\n")
	}

	s.WriteString("\n" + helpStyle.Render("↑/↓: move • space: toggle • enter: confirm • q: quit"))
	return s.String()
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
			return m, nil

		case " ":
			m.options[m.cursor].selected = !m.options[m.cursor].selected
			return m, nil

		case "enter":
			if m.step == StepFeatures {
				// Store feature selections
				needsShells := false
				for _, opt := range m.options {
					m.features[opt.title] = opt.selected
					if opt.selected && (opt.title == "Shell Completions" || opt.title == "Shell Integration") {
						needsShells = true
					}
				}

				// Move to shell selection if needed
				if needsShells {
					m.step = StepShells
					m.cursor = 0
					m.options = []Option{
						{title: "Bash", description: "Bourne Again Shell", selected: true},
						{title: "Zsh", description: "Z Shell", selected: true},
						{title: "Fish", description: "Friendly Interactive Shell", selected: true},
					}
					return m, nil
				}
			} else {
				// Store shell selections
				for _, opt := range m.options {
					m.shells[opt.title] = opt.selected
				}
			}
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) Done() bool {
	return m.done
}

func (m Model) GetFeatures() map[string]bool {
	return m.features
}

func (m Model) GetShells() map[string]bool {
	return m.shells
}
