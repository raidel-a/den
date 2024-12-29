package tui

import (
	"den/internal/config"
	"den/internal/editor"
	"den/internal/project"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.SetWindowTitle("Den - Project Manager")
}

// Update handles all state updates
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

		// Update list dimensions
		m.List.SetSize(msg.Width, msg.Height)

		// Recenter title with new width
		m.List.Styles.Title = m.List.Styles.Title.Copy().
			Width(msg.Width).
			Align(lipgloss.Center)

		return m, cmd

	case ProjectsLoadedMsg:
		items := make([]list.Item, len(msg))
		for i, p := range msg {
			items[i] = ListItem{Project: p}
		}
		m.List.SetItems(items)
		return m, cmd

	case tea.KeyMsg:
		// If we're filtering, let the list handle everything
		if m.List.FilterState() == list.Filtering {
			return m, cmd
		}

		// Handle directory addition mode
		if m.AddingDir {
			return m.handleAddingDirUpdate(msg)
		}

		// Handle normal mode
		switch msg.String() {
		case "a":
			if !m.ShowContext && !m.InputMode {
				m.AddingDir = true
				m.Input = ""
				m.Err = nil
				return m, nil
			}
		case "enter", " ":
			if !m.ShowContext && !m.InputMode {
				m.ShowContext = true
				m.ContextCursor = 0
				return m, nil
			}
		}

		if m.InputMode {
			return m.handleInputModeUpdate(msg)
		}

		if m.ShowContext {
			return m.handleContextMenuUpdate(msg)
		}

		// Return the list update command
		return m, cmd
	}

	return m, cmd
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

	default:
		if !m.ShowContext {
			m.Input += msg.String()
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
	switch msg.String() {
	case "up", "k":
		m.ContextCursor--
		if m.ContextCursor < 0 {
			m.ContextCursor = len(ContextOptions) - 1
		}
		return m, nil
	case "down", "j":
		m.ContextCursor = (m.ContextCursor + 1) % len(ContextOptions)
		return m, nil
	case "enter", " ":
		return m.handleContextMenuSelection()
	case "esc":
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

	// Reset state and rescan projects
	m.AddingDir = false
	m.Input = ""
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
