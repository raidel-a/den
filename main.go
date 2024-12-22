package main

import (
	"den/internal/cache"
	"den/internal/config"
	"den/internal/theme"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// Application state model
type Model struct {
	projects      []Project
	list          list.Model
	err           error
	config        *config.Config
	inputMode     bool
	input         string
	tabState      *tabCompletionState
	status        string
	showContext   bool
	contextCursor int
	addingDir     bool
	width         int
	height        int
}

type Project struct {
	Name     string
	Path     string
	LastMod  string
	GitState string
}

type tabCompletionState struct {
	suggestions []string
	index       int
}

type item struct {
	project Project
}

func (i item) Title() string { return i.project.Name }
func (i item) Description() string {
	desc := i.project.Path
	if i.project.GitState != "" {
		desc += " (" + i.project.GitState + ")"
	}
	return desc
}
func (i item) FilterValue() string {
	return i.project.Name + " " + i.project.Path
}

var debugMode bool

var defaultEditors = map[string][]string{
	"darwin":  {"code", "vim", "nano"},     // macOS
	"linux":   {"code", "vim", "nano"},     // Linux
	"windows": {"code.exe", "notepad.exe"}, // Windows
}

var contextOptions = []string{
	"Editor",
	"File Explorer",
	"Go to Dir",
	"Copy Path",
	"Cancel",
}

// Styling constants
var (
	activeTheme theme.Theme

	inputStyle = func() lipgloss.Style {
		return lipgloss.NewStyle().
			Foreground(activeTheme.Primary).
			MarginTop(1)
	}

	instructionStyle = func() lipgloss.Style {
		return lipgloss.NewStyle().
			Foreground(activeTheme.Primary).
			Margin(1, 0).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(activeTheme.Border)
	}

	titleStyle = func() lipgloss.Style {
		return lipgloss.NewStyle().
			Bold(true).
			Foreground(activeTheme.Primary).
			Background(activeTheme.Secondary).
			MarginBottom(1)
	}

	selectedItemStyle = func() lipgloss.Style {
		return lipgloss.NewStyle().
			Foreground(activeTheme.Primary).
			Margin(0, 2).
			Bold(true)
	}

	regularItemStyle = func() lipgloss.Style {
		return lipgloss.NewStyle().
			Foreground(activeTheme.Secondary)
	}

	contextStyle = func() lipgloss.Style {
		return lipgloss.NewStyle().
			Align(lipgloss.Center)
	}

	menuItemStyle = lipgloss.NewStyle().
			Faint(true)

	selectedMenuItemStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("87"))
)

func initialModel() Model {
	cfg, err := config.LoadConfig()
	if err != nil {
		return Model{
			err:           fmt.Errorf("failed to load config: %v", err),
			list:          list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
			tabState:      nil,
			showContext:   false,
			contextCursor: 0,
		}
	}

	// Set active theme
	activeTheme = theme.GetTheme(cfg.Preferences.Theme)

	// Try to load from cache first
	cache, err := cache.LoadCache()
	if err != nil {
		if debugMode {
			log.Printf("Error loading cache: %v", err)
		}
	}

	var projects []Project
	if cache != nil && cache.IsCacheValid(cfg) {
		// Use cached projects
		projects = convertCacheToProjects(cache.Projects)
		if debugMode {
			log.Printf("Using cached projects (%d items)", len(projects))
		}
	} else {
		// Scan directories and update cache
		projects = scanForProjects(cfg.ProjectDirs, cfg)
		if err := updateCache(projects); err != nil && debugMode {
			log.Printf("Error updating cache: %v", err)
		}
	}

	// Create custom delegate with themed colors
	delegate := list.NewDefaultDelegate()

	// Style the delegate to match our theme
	delegate.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(activeTheme.Text)

	delegate.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(activeTheme.Secondary)

	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(activeTheme.Primary).
		Foreground(activeTheme.Primary).
		Bold(true)

	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(activeTheme.Text)

	// Create list with themed delegate
	projectList := list.New([]list.Item{}, delegate, 0, 0)

	// Get terminal width for centering
	w, _, _ := term.GetSize(int(os.Stdout.Fd()))

	// Center and style the title
	projectList.Title = cfg.Preferences.ProjectListTitle
	projectList.Styles.Title = lipgloss.NewStyle().
		Foreground(activeTheme.Primary).
		Bold(true).
		Padding(0, 1).
		Width(w). // Use full terminal width
		Align(lipgloss.Center)

	projectList.SetShowHelp(true)

	// Add custom help
	projectList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("a"),
				key.WithHelp("a", "add directory"),
			),
		}
	}

	// Style the filter prompt
	projectList.Styles.FilterPrompt = lipgloss.NewStyle().
		Foreground(activeTheme.Secondary)

	// Style the filtered text
	projectList.Styles.FilterCursor = lipgloss.NewStyle().
		Foreground(activeTheme.Primary)

	items := make([]list.Item, len(projects))
	for i, p := range projects {
		items[i] = item{project: p}
	}
	projectList.SetItems(items)

	return Model{
		config:        cfg,
		list:          projectList,
		tabState:      nil,
		showContext:   false,
		contextCursor: 0,
		inputMode:     len(cfg.ProjectDirs) == 0,
	}
}

func updateCache(projects []Project) error {
	cache := &cache.ProjectCache{
		Projects:     convertProjectsToCache(projects),
		LastUpdated:  time.Now(),
		DirectoryMap: make(map[string]int),
	}

	// Update directory map
	for _, p := range projects {
		dir := filepath.Dir(p.Path)
		cache.DirectoryMap[dir]++
	}

	// Call SaveCache as a method on the cache instance
	return cache.SaveCache()
}

func (m Model) Init() tea.Cmd {
	// Set the window title when the program starts
	return tea.SetWindowTitle("Den - Project Manager")
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update list dimensions
		m.list.SetSize(msg.Width, msg.Height)

		// Recenter title with new width
		m.list.Styles.Title = m.list.Styles.Title.Copy().
			Width(msg.Width).
			Align(lipgloss.Center)

		return m, nil

	case projectsLoadedMsg:
		items := make([]list.Item, len(msg))
		for i, p := range msg {
			items[i] = item{project: p}
		}
		m.list.SetItems(items)
		return m, nil

	case tea.KeyMsg:
		// Handle directory addition mode
		if m.addingDir {
			switch msg.Type {
			case tea.KeyTab:
				if m.tabState == nil {
					suggestions := getPathSuggestions(m.input)
					if len(suggestions) > 0 {
						m.tabState = &tabCompletionState{
							suggestions: suggestions,
							index:       0,
						}
						m.input = m.tabState.suggestions[0]
					}
				} else {
					m.tabState.index = (m.tabState.index + 1) % len(m.tabState.suggestions)
					m.input = m.tabState.suggestions[m.tabState.index]
				}
				return m, nil

			case tea.KeyEnter:
				m.tabState = nil
				path := m.input
				if abs, err := filepath.Abs(path); err == nil {
					path = abs
				}

				// Validate directory
				if info, err := os.Stat(path); err != nil {
					m.err = fmt.Errorf("cannot access directory: %v", err)
					return m, nil
				} else if !info.IsDir() {
					m.err = fmt.Errorf("not a directory: %s", path)
					return m, nil
				}

				// Add directory and save config
				m.config.ProjectDirs = append(m.config.ProjectDirs, path)
				if err := config.SaveConfig(m.config); err != nil {
					m.err = fmt.Errorf("failed to save config: %v", err)
					return m, nil
				}

				// Reset state and rescan projects
				m.addingDir = false
				m.input = ""
				return m, func() tea.Msg {
					return projectsLoadedMsg(scanForProjects(m.config.ProjectDirs, m.config))
				}

			case tea.KeyEsc:
				m.addingDir = false
				m.input = ""
				m.err = nil
				return m, nil

			default:
				if !m.showContext {
					m.input += msg.String()
				}
				m.tabState = nil
				return m, nil
			}
		}

		// Handle normal mode
		switch msg.String() {
		case "a":
			if !m.showContext && !m.inputMode {
				m.addingDir = true
				m.input = ""
				m.err = nil
				return m, nil
			}
		}

		if m.inputMode {
			switch msg.Type {
			case tea.KeyTab:
				if m.tabState == nil {
					suggestions := getPathSuggestions(m.input)
					if len(suggestions) > 0 {
						m.tabState = &tabCompletionState{
							suggestions: suggestions,
							index:       0,
						}
						m.input = m.tabState.suggestions[0]
					}
				} else {
					m.tabState.index = (m.tabState.index + 1) % len(m.tabState.suggestions)
					m.input = m.tabState.suggestions[m.tabState.index]
				}
				return m, nil

			case tea.KeyEnter:
				m.tabState = nil
				path := m.input
				if abs, err := filepath.Abs(path); err == nil {
					path = abs
				}

				if info, err := os.Stat(path); err != nil {
					m.err = fmt.Errorf("cannot access directory: %v", err)
					if debugMode {
						log.Printf("Error validating directory %s: %v", path, err)
					}
					return m, nil
				} else if !info.IsDir() {
					m.err = fmt.Errorf("not a directory: %s", path)
					if debugMode {
						log.Printf("Path is not a directory: %s", path)
					}
					return m, nil
				}

				m.config.ProjectDirs = append(m.config.ProjectDirs, path)
				if err := config.SaveConfig(m.config); err != nil {
					m.err = fmt.Errorf("failed to save config: %v", err)
					if debugMode {
						log.Printf("Error saving config: %v", err)
					}
					return m, nil
				}

				if debugMode {
					log.Printf("Added new project directory: %s", path)
				}

				m.inputMode = false
				return m, func() tea.Msg {
					return projectsLoadedMsg(scanForProjects(m.config.ProjectDirs, m.config))
				}

			case tea.KeyEsc:
				if len(m.config.ProjectDirs) > 0 {
					m.inputMode = false
					return m, nil
				}
				return m, tea.Quit

			case tea.KeyBackspace, tea.KeyDelete:
				m.tabState = nil
				if len(m.input) > 0 {
					m.input = m.input[:len(m.input)-1]
				}
				m.err = nil
				return m, nil

			default:
				if msg.Type == tea.KeyRunes {
					m.tabState = nil
					m.input += string(msg.Runes)
					m.err = nil
				}
				return m, nil
			}
		}

		if m.showContext {
			switch msg.String() {
			case "up", "k":
				m.contextCursor--
				if m.contextCursor < 0 {
					m.contextCursor = len(contextOptions) - 1
				}
				return m, nil
			case "down", "j":
				m.contextCursor = (m.contextCursor + 1) % len(contextOptions)
				return m, nil
			case "enter", " ":
				if i, ok := m.list.SelectedItem().(item); ok {
					switch m.contextCursor {
					case 0: // Open in Editor
						if err := openInEditor(i.project.Path, m.config); err != nil {
							m.status = fmt.Sprintf("Error opening editor: %v", err)
						}
						return m, tea.Quit

					case 1: // Open in File Explorer
						if err := openInFileExplorer(i.project.Path, m.config); err != nil {
							m.status = fmt.Sprintf("Error opening file explorer: %v", err)
						}
						return m, tea.Quit

					case 2: // Change Working Directory
						if err := os.Chdir(i.project.Path); err != nil {
							m.status = fmt.Sprintf("Error changing directory: %v", err)
						} else {
							fmt.Printf("\nChanged working directory to: %s\n", i.project.Path)
						}
						return m, tea.Quit

					case 3: // Copy Path
						if err := copyToClipboard(i.project.Path); err != nil {
							m.status = fmt.Sprintf("Error copying to clipboard: %v", err)
						} else {
							m.status = "Path copied to clipboard"
						}
						m.showContext = false
					}
				}
				m.showContext = false
				return m, nil
			case "esc":
				m.showContext = false
				return m, nil
			}
			return m, nil
		}

		if !m.inputMode {
			switch msg.String() {
			case "enter", " ":
				m.showContext = true
				m.contextCursor = 0
				return m, nil
			}
		}
	}

	// Handle list updates when not in input mode and context menu is not shown
	if !m.inputMode && !m.showContext {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {
	if m.addingDir {
		var s strings.Builder
		s.WriteString(titleStyle().Render("Add Project Directory"))
		s.WriteString("\n\n")
		s.WriteString(instructionStyle().Render(
			"Enter the path to your projects directory.\n" +
				"This is typically something like ~/Developer or ~/Projects\n",
		))
		s.WriteString("\n" + inputStyle().Render(m.input))

		if m.tabState != nil && len(m.tabState.suggestions) > 0 {
			s.WriteString("\n\nSuggestions:\n")
			for i, sugg := range m.tabState.suggestions {
				if i == m.tabState.index {
					s.WriteString(selectedItemStyle().Render("> " + sugg + "\n"))
				} else {
					s.WriteString(regularItemStyle().Render("  " + sugg + "\n"))
				}
			}
		}

		if m.err != nil {
			s.WriteString("\n\n" + lipgloss.NewStyle().
				Foreground(activeTheme.Error).
				Render(fmt.Sprintf("Error: %v", m.err)))
		}

		s.WriteString("\n\n" + regularItemStyle().Render(
			"Tab: autocomplete • Enter: confirm • Esc: cancel",
		))

		return s.String()
	}

	if m.showContext {
		var s strings.Builder
		s.WriteString(m.list.View())

		// Create horizontal menu
		var menuItems []string
		width := 0
		maxWidth := m.list.Width() - 4 // Account for margins

		for i, opt := range contextOptions {
			item := opt
			if i == m.contextCursor {
				item = selectedMenuItemStyle.Render(opt)
			} else {
				item = menuItemStyle.Render(opt)
			}

			// Check if adding this item would exceed available width
			itemWidth := lipgloss.Width(item) + 1 // +1 for separator
			if width+itemWidth > maxWidth {
				break
			}

			menuItems = append(menuItems, item)
			width += itemWidth
		}

		menu := contextStyle().Render(strings.Join(menuItems, "•"))
		s.WriteString("\n" + menu)

		return s.String()
	}

	return m.list.View()
}

func printHelp() {
	fmt.Println(`Den - A Cozy Home for Your Repos

Usage:
    den [flags]

Flags:
    --help      Show this help message
    --reset     Reset all configuration and start fresh
    --debug     Enable debug logging
    
No flags will start the interactive UI.`)
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--help", "-h":
			printHelp()
			os.Exit(0)

		case "--reset":
			configPath, err := config.GetConfigPath()
			if err != nil {
				fmt.Printf("Error getting config path: %v\n", err)
				os.Exit(1)
			}

			// Remove the config file
			if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
				fmt.Printf("Error removing config: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("Configuration has been reset.")
			os.Exit(0)

		case "--debug":
			// Setup logging to file
			f, err := tea.LogToFile("debug.log", "debug")
			if err != nil {
				fmt.Printf("Error setting up logging: %v\n", err)
				os.Exit(1)
			}
			defer f.Close()

			// Enable verbose logging in scanForProjects
			debugMode = true

		default:
			fmt.Printf("Unknown flag: %s\n\n", os.Args[1])
			printHelp()
			os.Exit(1)
		}
	}

	// Normal program execution
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
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

func scanForProjects(dirs []string, config *config.Config) []Project {
	var projects []Project

	if debugMode {
		log.Printf("Scanning directories: %v", dirs)
	}

	for _, dir := range dirs {
		if debugMode {
			log.Printf("Scanning directory: %s", dir)
		}

		// Check if directory exists and is accessible
		if _, err := os.Stat(dir); err != nil {
			if debugMode {
				log.Printf("Error accessing directory %s: %v", dir, err)
			}
			continue
		}

		// Read directory entries
		entries, err := os.ReadDir(dir)
		if err != nil {
			if debugMode {
				log.Printf("Error reading directory %s: %v", dir, err)
			}
			continue
		}

		if debugMode {
			log.Printf("Found %d entries in %s", len(entries), dir)
		}

		// Check each entry
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			fullPath := filepath.Join(dir, entry.Name())
			if debugMode {
				log.Printf("Checking potential project: %s", fullPath)
			}

			project, err := detectProject(fullPath, config)
			if err != nil {
				if debugMode {
					log.Printf("Error detecting project at %s: %v", fullPath, err)
				}
				continue
			}

			if debugMode {
				log.Printf("Found project: %s (Git: %s)", project.Name, project.GitState)
			}

			projects = append(projects, *project)
		}
	}

	if debugMode {
		log.Printf("Total projects found: %d", len(projects))
	}

	return projects
}

type projectsLoadedMsg []Project

func hasProjectFile(path string) bool {
	projectFiles := []string{
		"package.json",     // Node.js
		"Cargo.toml",       // Rust
		"go.mod",           // Go
		"requirements.txt", // Python
		"pom.xml",          // Java/Maven
		"build.gradle",     // Java/Gradle
		"Gemfile",          // Ruby
		"composer.json",    // PHP
	}

	for _, file := range projectFiles {
		if _, err := os.Stat(filepath.Join(path, file)); err == nil {
			return true
		}
	}
	return false
}

func getGitStatus(path string) string {
	cmd := exec.Command("git", "-C", path, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return "no git"
	}

	if len(output) > 0 {
		return "git (modified)"
	}
	return "git (clean)"
}

func openInEditor(path string, config *config.Config) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		// Use configured default editor
		editor = config.Preferences.DefaultEditor
		if editor == "" {
			// Try to find a default editor from the configured list
			for _, ed := range config.Preferences.EditorList {
				if _, err := exec.LookPath(ed); err == nil {
					editor = ed
					break
				}
			}
			if editor == "" {
				return fmt.Errorf("no editor found")
			}
		}
	}

	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func openInFileExplorer(path string, config *config.Config) error {
	var cmd *exec.Cmd

	// Use configured file manager if set
	if config.Preferences.DefaultFileManager != "" {
		cmd = exec.Command(config.Preferences.DefaultFileManager, path)
	} else {
		// Fall back to OS defaults
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", path)
		case "windows":
			cmd = exec.Command("explorer", path)
		default: // Linux and others
			cmd = exec.Command("xdg-open", path)
		}
	}

	return cmd.Run()
}

func detectSystemPreferences() config.UserPreferences {
	prefs := config.UserPreferences{
		ShowGitStatus:    true,
		ShowHiddenFiles:  false,
		Theme:            "default",
		ProjectListTitle: "Your Projects",
	}

	// Detect default editor
	if editor := os.Getenv("EDITOR"); editor != "" {
		prefs.DefaultEditor = editor
	} else {
		// Try to find common editors
		for _, ed := range defaultEditors[runtime.GOOS] {
			if _, err := exec.LookPath(ed); err == nil {
				prefs.DefaultEditor = ed
				break
			}
		}
	}

	// Detect default file manager
	switch runtime.GOOS {
	case "darwin":
		prefs.DefaultFileManager = "open"
	case "windows":
		prefs.DefaultFileManager = "explorer"
	default:
		if _, err := exec.LookPath("xdg-open"); err == nil {
			prefs.DefaultFileManager = "xdg-open"
		}
	}

	return prefs
}

func detectProject(path string, config *config.Config) (*Project, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("not a directory")
	}

	name := filepath.Base(path)
	lastMod := info.ModTime().Format("2006-01-02 15:04:05")

	gitState := "no git"
	// Only check git status if enabled in preferences
	if config.Preferences.ShowGitStatus {
		if _, err := os.Stat(filepath.Join(path, ".git")); err == nil {
			gitState = getGitStatus(path)
		}
	}

	return &Project{
		Name:     name,
		Path:     path,
		LastMod:  lastMod,
		GitState: gitState,
	}, nil
}

func convertCacheToProjects(cached []cache.Project) []Project {
	projects := make([]Project, len(cached))
	for i, p := range cached {
		projects[i] = Project{
			Name:     p.Name,
			Path:     p.Path,
			LastMod:  p.LastMod.Format("2006-01-02 15:04:05"),
			GitState: p.GitState,
		}
	}
	return projects
}

func convertProjectsToCache(projects []Project) []cache.Project {
	cached := make([]cache.Project, len(projects))
	for i, p := range projects {
		lastMod, _ := time.Parse("2006-01-02 15:04:05", p.LastMod)
		cached[i] = cache.Project{
			Name:     p.Name,
			Path:     p.Path,
			LastMod:  lastMod,
			GitState: p.GitState,
		}
	}
	return cached
}

func copyToClipboard(text string) error {
	return clipboard.WriteAll(text)
}
