package tui

import (
	"den/internal/config"
	"den/internal/editor"
	"den/internal/project"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.SetWindowTitle("Den - Project Manager")
}

// Update handles all state updates
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

		// Update list dimensions
		m.List.SetSize(msg.Width, msg.Height)

		// Update instruction width based on window size
		m.Styles.Instruction = m.Styles.Instruction.Copy().
			Width(msg.Width - 4) // Subtract some padding

		// Update title width based on window size
		m.Styles.Title = m.Styles.Title.Copy().
			Width(msg.Width - 4)

		return m, cmd

	case ProjectsLoadedMsg:
		items := make([]list.Item, len(msg))
		for i, p := range msg {
			items[i] = ListItem{Project: p}
		}
		m.List.SetItems(items)
		return m, cmd

	case tea.KeyMsg:
		// First check if the list wants to handle this key message
		if !m.ShowContext && !m.AddingDir && !m.InputMode {
			// Always let the list handle filtering keys
			if m.List.FilterState() == list.Filtering {
				var cmd tea.Cmd
				m.List, cmd = m.List.Update(msg)
				return m, cmd
			}

			// Check for filter trigger
			if key.Matches(msg, m.List.KeyMap.Filter) {
				var cmd tea.Cmd
				m.List, cmd = m.List.Update(msg)
				return m, cmd
			}
		}

		// Handle directory addition mode
		if m.AddingDir {
			return m.handleAddingDirUpdate(msg)
		}

		// Handle context menu if it's shown
		if m.ShowContext {
			return m.handleContextMenuUpdate(msg)
		}

		// Handle input mode
		if m.InputMode {
			return m.handleInputModeUpdate(msg)
		}

		// Handle normal mode key presses
		switch {
		case key.Matches(msg, m.KeyMap.AddDirectory):
			if !m.ShowContext && !m.InputMode {
				m.AddingDir = true
				m.Input = ""
				m.Err = nil
				return m, nil
			}
		case key.Matches(msg, m.KeyMap.ShowContext):
			if !m.ShowContext && !m.InputMode && m.List.FilterState() != list.Filtering {
				m.ShowContext = true
				m.ContextCursor = 0
				return m, nil
			}
		case key.Matches(msg, m.KeyMap.OpenConfig):
			if !m.ShowContext && !m.InputMode && !m.AddingDir {
				configPath, err := config.GetConfigPath()
				if err != nil {
					m.Status = fmt.Sprintf("Error getting config path: %v", err)
					return m, nil
				}
				if err := editor.OpenInEditor(configPath, m.Config); err != nil {
					m.Status = fmt.Sprintf("Error opening config: %v", err)
					return m, nil
				}
				return m, tea.Quit
			}
		}

		// Let the list handle all other keys
		var cmd tea.Cmd
		m.List, cmd = m.List.Update(msg)
		return m, cmd

	default:
		// Make sure to pass all other messages to the list
		var cmd tea.Cmd
		m.List, cmd = m.List.Update(msg)
		return m, cmd
	}
}

func (m Model) handleAddingDirUpdate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyTab:
		if m.TabState == nil {
			suggestions := getPathSuggestions(m.Input)
			if len(suggestions) > 0 {
				m.TabState = &TabCompletionState{
					Suggestions: suggestions,
					Index:       0,
				}
				m.Input = m.TabState.Suggestions[0]
			}
		} else {
			m.TabState.Index = (m.TabState.Index + 1) % len(m.TabState.Suggestions)
			m.Input = m.TabState.Suggestions[m.TabState.Index]
		}
		return m, nil

	case tea.KeyEnter:
		return m.handleNewDirectoryConfirmation()

	case tea.KeyEsc:
		m.AddingDir = false
		m.Input = ""
		m.Err = nil
		return m, nil

	case tea.KeyBackspace, tea.KeyDelete:
		m.TabState = nil
		if len(m.Input) > 0 {
			m.Input = m.Input[:len(m.Input)-1]
		}
		m.Err = nil
		return m, nil

	default:
		if !m.ShowContext && msg.Type == tea.KeyRunes {
			m.Input += string(msg.Runes)
			m.TabState = nil
		}
		return m, nil
	}
}

func (m Model) handleInputModeUpdate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyTab:
		if m.TabState == nil {
			suggestions := getPathSuggestions(m.Input)
			if len(suggestions) > 0 {
				m.TabState = &TabCompletionState{
					Suggestions: suggestions,
					Index:       0,
				}
				m.Input = m.TabState.Suggestions[0]
			}
		} else {
			m.TabState.Index = (m.TabState.Index + 1) % len(m.TabState.Suggestions)
			m.Input = m.TabState.Suggestions[m.TabState.Index]
		}
		return m, nil

	case tea.KeyEnter:
		return m.handleNewDirectoryConfirmation()

	case tea.KeyEsc:
		if len(m.Config.ProjectDirs) > 0 {
			m.InputMode = false
			return m, nil
		}
		return m, tea.Quit

	case tea.KeyBackspace, tea.KeyDelete:
		m.TabState = nil
		if len(m.Input) > 0 {
			m.Input = m.Input[:len(m.Input)-1]
		}
		m.Err = nil
		return m, nil

	default:
		if msg.Type == tea.KeyRunes {
			m.TabState = nil
			m.Input += string(msg.Runes)
			m.Err = nil
		}
		return m, nil
	}
}

func (m Model) handleContextMenuUpdate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.KeyMap.Up):
		m.ContextCursor--
		if m.ContextCursor < 0 {
			m.ContextCursor = len(ContextOptions) - 1
		}
		return m, nil
	case key.Matches(msg, m.KeyMap.Down):
		m.ContextCursor = (m.ContextCursor + 1) % len(ContextOptions)
		return m, nil
	case key.Matches(msg, m.KeyMap.Enter):
		return m.handleContextMenuSelection()
	case key.Matches(msg, m.KeyMap.Escape):
		m.ShowContext = false
		return m, nil
	}
	return m, nil
}

func (m Model) handleContextMenuSelection() (tea.Model, tea.Cmd) {
	if i, ok := m.List.SelectedItem().(ListItem); ok {
		switch m.ContextCursor {
		case 0: // Open in Editor
			if err := editor.OpenInEditor(i.Project.Path, m.Config); err != nil {
				m.Status = fmt.Sprintf("Error opening editor: %v", err)
			}
			return m, tea.Quit

		case 1: // Open in File Explorer
			if err := editor.OpenInFileExplorer(i.Project.Path, m.Config); err != nil {
				m.Status = fmt.Sprintf("Error opening file explorer: %v", err)
			}
			return m, tea.Quit

		case 2: // Change Working Directory
			if err := os.Chdir(i.Project.Path); err != nil {
				m.Status = fmt.Sprintf("Error changing directory: %v", err)
			} else {
				fmt.Printf("\nChanged working directory to: %s\n", i.Project.Path)
			}
			return m, tea.Quit

		case 3: // Copy Path
			if err := CopyToClipboard(i.Project.Path); err != nil {
				m.Status = fmt.Sprintf("Error copying to clipboard: %v", err)
			} else {
				m.Status = "Path copied to clipboard"
			}
			m.ShowContext = false
		}
	}
	m.ShowContext = false
	return m, nil
}

func (m Model) handleNewDirectoryConfirmation() (tea.Model, tea.Cmd) {
	m.TabState = nil
	path := m.Input
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}

	// Validate directory
	if info, err := os.Stat(path); err != nil {
		m.Err = fmt.Errorf("cannot access directory: %v", err)
		return m, nil
	} else if !info.IsDir() {
		m.Err = fmt.Errorf("not a directory: %s", path)
		return m, nil
	}

	// Add directory and save config
	m.Config.ProjectDirs = append(m.Config.ProjectDirs, path)
	if err := config.SaveConfig(m.Config); err != nil {
		m.Err = fmt.Errorf("failed to save config: %v", err)
		return m, nil
	}

	// Reset state and prepare for project scanning
	m.AddingDir = false
	m.InputMode = false
	m.Input = ""

	// Reinitialize list key bindings
	m.List.KeyMap.Filter.SetEnabled(true)
	m.List.KeyMap.ShowFullHelp.SetEnabled(true)
	m.List.KeyMap.CancelWhileFiltering.SetEnabled(true)
	m.List.KeyMap.AcceptWhileFiltering.SetEnabled(true)
	m.List.SetFilteringEnabled(true)
	m.List.SetShowFilter(true)
	m.List.SetShowHelp(true)

	// Return command to scan for projects
	return m, func() tea.Msg {
		return ProjectsLoadedMsg(project.ScanForProjects(m.Config.ProjectDirs, m.Config))
	}
}

func getPathSuggestions(partial string) []string {
	if partial == "" {
		partial = "."
	}

	// Get the directory to search in
	dir := filepath.Dir(partial)
	if dir == "." {
		dir = "/"
	}

	// Get the partial filename to match against
	prefix := filepath.Base(partial)

	// Read the directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var suggestions []string
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, prefix) {
			fullPath := filepath.Join(dir, name)
			if entry.IsDir() {
				fullPath += "/"
			}
			suggestions = append(suggestions, fullPath)
		}
	}

	return suggestions
}
