package tui

import (
	"den/internal/cache"
	"den/internal/config"
	"den/internal/editor"
	"den/internal/project"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// Init initializes the model
func (m Model) Init() tea.Cmd {
	if m.AddingDir {
		// Show initial suggestions
		suggestions := getPathSuggestions("", m.Config)
		m.TabState = &TabCompletionState{
			Suggestions: suggestions,
			Index:       0,
			Page:        0,
			PageSize:    DefaultPageSize,
		}
	}
	return nil
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
				// Show initial suggestions
				suggestions := getPathSuggestions("", m.Config)
				m.TabState = &TabCompletionState{
					Suggestions: suggestions,
					Index:       0,
					Page:        0,
					PageSize:    DefaultPageSize,
				}
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
		case msg.String() == "F":
			m.ShowFavoritesOnly = !m.ShowFavoritesOnly
			if m.ShowFavoritesOnly {
				// Filter to show only favorites
				items := m.List.Items()
				favoriteItems := make([]list.Item, 0)
				for _, item := range items {
					if listItem, ok := item.(ListItem); ok && listItem.Project.Favorite {
						favoriteItems = append(favoriteItems, item)
					}
				}
				m.List.SetItems(favoriteItems)
			} else {
				// Restore all items
				items := make([]list.Item, len(m.Projects))
				for i, p := range m.Projects {
					items[i] = ListItem{Project: p}
				}
				m.List.SetItems(items)
			}
			return m, nil
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
	case tea.KeyEnter:
		return m.handleNewDirectoryConfirmation()

	case tea.KeyEsc:
		m.AddingDir = false
		m.Input = ""
		m.TabState = nil
		m.Err = nil
		return m, nil

	case tea.KeyBackspace:
		if len(m.Input) > 0 {
			m.Input = m.Input[:len(m.Input)-1]
			// Update suggestions after backspace
			suggestions := getPathSuggestions(m.Input, m.Config)
			m.TabState = &TabCompletionState{
				Suggestions: suggestions,
				Index:       0,
				Page:        0,
				PageSize:    DefaultPageSize,
			}
		}
		return m, nil

	case tea.KeyTab:
		if m.TabState != nil && len(m.TabState.Suggestions) > 0 {
			// Complete with current selection
			m.Input = m.TabState.Suggestions[m.TabState.Index]
			// Get new suggestions for the completed path
			suggestions := getPathSuggestions(m.Input, m.Config)
			m.TabState = &TabCompletionState{
				Suggestions: suggestions,
				Index:       0,
				Page:        0,
				PageSize:    DefaultPageSize,
			}
		}
		return m, nil

	case tea.KeyUp, tea.KeyDown:
		if m.TabState != nil && len(m.TabState.Suggestions) > 0 {
			start := m.TabState.Page * m.TabState.PageSize
			end := start + m.TabState.PageSize
			if end > len(m.TabState.Suggestions) {
				end = len(m.TabState.Suggestions)
			}

			if msg.Type == tea.KeyUp {
				m.TabState.Index--
				if m.TabState.Index < start {
					m.TabState.Index = end - 1
				}
			} else {
				m.TabState.Index++
				if m.TabState.Index >= end {
					m.TabState.Index = start
				}
			}
		}
		return m, nil

	case tea.KeyLeft, tea.KeyRight:
		if m.TabState != nil && len(m.TabState.Suggestions) > 0 {
			totalPages := (len(m.TabState.Suggestions) + m.TabState.PageSize - 1) / m.TabState.PageSize
			if msg.Type == tea.KeyLeft {
				m.TabState.Page--
				if m.TabState.Page < 0 {
					m.TabState.Page = totalPages - 1
				}
			} else {
				m.TabState.Page = (m.TabState.Page + 1) % totalPages
			}
			// Adjust index for new page
			start := m.TabState.Page * m.TabState.PageSize
			m.TabState.Index = start
		}
		return m, nil

	default:
		switch msg.String() {
		case "ctrl+j":
			if m.TabState != nil && len(m.TabState.Suggestions) > 0 {
				start := m.TabState.Page * m.TabState.PageSize
				end := start + m.TabState.PageSize
				if end > len(m.TabState.Suggestions) {
					end = len(m.TabState.Suggestions)
				}
				m.TabState.Index++
				if m.TabState.Index >= end {
					m.TabState.Index = start
				}
			}
			return m, nil

		case "ctrl+k":
			if m.TabState != nil && len(m.TabState.Suggestions) > 0 {
				start := m.TabState.Page * m.TabState.PageSize
				end := start + m.TabState.PageSize
				if end > len(m.TabState.Suggestions) {
					end = len(m.TabState.Suggestions)
				}
				m.TabState.Index--
				if m.TabState.Index < start {
					m.TabState.Index = end - 1
				}
			}
			return m, nil

		case "ctrl+h", "ctrl+l":
			if m.TabState != nil && len(m.TabState.Suggestions) > 0 {
				totalPages := (len(m.TabState.Suggestions) + m.TabState.PageSize - 1) / m.TabState.PageSize
				if msg.String() == "ctrl+h" {
					m.TabState.Page--
					if m.TabState.Page < 0 {
						m.TabState.Page = totalPages - 1
					}
				} else {
					m.TabState.Page = (m.TabState.Page + 1) % totalPages
				}
				start := m.TabState.Page * m.TabState.PageSize
				m.TabState.Index = start
			}
			return m, nil

		default:
			if !m.ShowContext && msg.Type == tea.KeyRunes {
				m.Input += string(msg.Runes)
				// Get suggestions immediately when typing
				suggestions := getPathSuggestions(m.Input, m.Config)
				m.TabState = &TabCompletionState{
					Suggestions: suggestions,
					Index:       0,
					Page:        0,
					PageSize:    DefaultPageSize,
				}
			}
			return m, nil
		}
	}
}

func (m Model) handleInputModeUpdate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyTab:
		if m.TabState == nil {
			suggestions := getPathSuggestions(m.Input, m.Config)
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
			m.Input += string(msg.Runes)
			// Get suggestions immediately when typing
			suggestions := getPathSuggestions(m.Input, m.Config)
			if len(suggestions) > 0 {
				m.TabState = &TabCompletionState{
					Suggestions: suggestions,
					Index:       0,
					Page:        0,
					PageSize:    DefaultPageSize,
				}
			} else {
				m.TabState = nil
			}
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

		case 2: // Copy Path
			if err := CopyToClipboard(i.Project.Path); err != nil {
				m.Status = fmt.Sprintf("Error copying to clipboard: %v", err)
			} else {
				m.Status = "Path copied to clipboard"
			}
			m.ShowContext = false

		case 3: // Toggle Favorite
			// Toggle favorite status
			i.Project.Favorite = !i.Project.Favorite

			// Update favorites in config
			if i.Project.Favorite {
				m.Config.Favorites = append(m.Config.Favorites, i.Project.Path)
			} else {
				// Remove from favorites
				favorites := make([]string, 0)
				for _, fav := range m.Config.Favorites {
					if fav != i.Project.Path {
						favorites = append(favorites, fav)
					}
				}
				m.Config.Favorites = favorites
			}

			// Save config
			if err := config.SaveConfig(m.Config); err != nil {
				m.Status = fmt.Sprintf("Error saving favorites: %v", err)
				return m, nil
			}

			// Update project in the main list
			for idx, proj := range m.Projects {
				if proj.Path == i.Project.Path {
					m.Projects[idx].Favorite = i.Project.Favorite
					break
				}
			}

			// Update cache
			cache := &cache.ProjectCache{
				Projects:    project.ConvertProjectsToCache(m.Projects),
				LastUpdated: time.Now(),
			}
			if err := cache.SaveCache(); err != nil {
				m.Status = fmt.Sprintf("Error saving cache: %v", err)
				return m, nil
			}

			// Update list items
			items := make([]list.Item, len(m.Projects))
			for i, p := range m.Projects {
				items[i] = ListItem{Project: p}
			}
			m.List.SetItems(items)

			m.Status = "Favorite status updated"
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

	// Check if directory already exists in config
	for _, dir := range m.Config.ProjectDirs {
		if dir == path {
			m.Err = fmt.Errorf("directory already exists in config: %s", path)
			return m, nil
		}
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

func getPathSuggestions(partial string, cfg *config.Config) []string {
	// Determine the directory to search
	var dir string
	if partial == "" {
		dir = "/"
	} else {
		// If path ends with slash, use it as directory, otherwise use parent
		if strings.HasSuffix(partial, "/") {
			dir = partial
		} else {
			dir = filepath.Dir(partial)
		}
	}

	// Get the filter text (only if path doesn't end with slash)
	var filter string
	if partial != "" && !strings.HasSuffix(partial, "/") {
		filter = filepath.Base(partial)
		if filter == "/" {
			filter = ""
		}
	}

	// Read the directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var suggestions []string
	for _, entry := range entries {
		if entry.IsDir() {
			name := entry.Name()
			// Skip hidden directories if showHiddenFiles is false
			if !cfg.Preferences.ShowHiddenFiles && strings.HasPrefix(name, ".") {
				continue
			}
			// Only filter if there's input to filter by
			if filter == "" || strings.HasPrefix(strings.ToLower(name), strings.ToLower(filter)) {
				fullPath := filepath.Join(dir, name)
				fullPath += "/"
				suggestions = append(suggestions, fullPath)
			}
		}
	}

	return suggestions
}
