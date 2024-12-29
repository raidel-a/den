package cli

import (
	"den/internal/cache"
	"den/internal/cli/completion"
	"den/internal/cli/man"
	"den/internal/config"
	"den/internal/project"
	"den/internal/theme"
	"den/internal/tui"
	"den/internal/ui"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// Version information
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// Command line flags
const (
	helpFlag    = "--help"
	versionFlag = "--version"
	resetFlag   = "--reset"
	debugFlag   = "--debug"
	installFlag = "--install"
)

type CLI struct {
	debugMode bool
}

func New() *CLI {
	return &CLI{}
}

// Run executes the CLI application
func (c *CLI) Run() error {
	if len(os.Args) > 1 {
		return c.handleFlags()
	}
	return c.startUI()
}

// handleFlags processes command line flags
func (c *CLI) handleFlags() error {
	switch os.Args[1] {
	case helpFlag, "-h":
		c.printHelp()
		return nil
	case versionFlag, "-v":
		c.printVersion()
		return nil
	case resetFlag:
		return c.resetConfig()
	case debugFlag:
		return c.enableDebugMode()
	case installFlag:
		return c.install()
	default:
		fmt.Printf("Unknown flag: %s\n\n", os.Args[1])
		c.printHelp()
		return fmt.Errorf("invalid flag")
	}
}

// printHelp displays comprehensive help information
func (c *CLI) printHelp() {
	fmt.Print(`Den - A Cozy Home for Your Repos

Description:
    Den is a terminal-based repository manager that provides a comfortable
    interface for managing and navigating your Git repositories.

Usage:
    den [flags]

Flags:
    -h, --help      Show help information
    -v, --version   Display version information
    --reset         Reset all configuration and start fresh
    --debug         Enable debug logging
    --install       Install shell completions and man pages

Examples:
    # Start Den's interactive UI
    den

    # Reset configuration
    den --reset

    # Enable debug mode
    den --debug

Configuration:
    Den stores its configuration in ~/.config/den/config.yaml
    Cache is stored in ~/.cache/den/

For more information, visit: https://github.com/yourusername/den
Report bugs at: https://github.com/yourusername/den/issues
`)
}

// printVersion displays version information
func (c *CLI) printVersion() {
	if info, ok := debug.ReadBuildInfo(); ok && version == "dev" {
		version = info.Main.Version
	}
	fmt.Printf("den version %s\n", version)
	fmt.Printf("commit: %s\n", commit)
	fmt.Printf("built: %s\n", date)
}

// enableDebugMode sets up debug logging
func (c *CLI) enableDebugMode() error {
	c.debugMode = true
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		return fmt.Errorf("failed to setup logging: %v", err)
	}
	defer f.Close()
	return c.startUI()
}

// resetConfig removes the configuration file
func (c *CLI) resetConfig() error {
	configPath, err := config.GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %v", err)
	}

	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove config: %v", err)
	}

	fmt.Println("Configuration has been reset.")
	return nil
}

// install installs shell completions and man pages
func (c *CLI) install() error {
	if err := completion.InstallCompletions(); err != nil {
		return fmt.Errorf("failed to install shell completions: %v", err)
	}

	if err := man.InstallManPage(version); err != nil {
		return fmt.Errorf("failed to install man page: %v", err)
	}

	fmt.Println("Successfully installed shell completions and man pages.")
	return nil
}

// startUI initializes and runs the terminal UI
func (c *CLI) startUI() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// Set active theme and create styles
	activeTheme := theme.GetTheme(cfg.Preferences.Theme)
	styles := ui.NewStyles(activeTheme)

	// Try to load from cache first
	projectCache, err := cache.LoadCache()
	if err != nil && c.debugMode {
		fmt.Printf("Error loading cache: %v\n", err)
	}

	var projects []project.Project
	if projectCache != nil && projectCache.IsCacheValid(cfg) {
		// Use cached projects
		projects = project.ConvertCacheToProjects(projectCache.Projects)
		if c.debugMode {
			fmt.Printf("Using cached projects (%d items)\n", len(projects))
		}
	} else {
		// Scan directories and update cache
		projects = project.ScanForProjects(cfg.ProjectDirs, cfg)

		// Update cache
		projectCache = &cache.ProjectCache{
			Projects:     project.ConvertProjectsToCache(projects),
			LastUpdated:  time.Now(),
			DirectoryMap: make(map[string]int),
		}

		// Update directory map
		for _, p := range projects {
			dir := filepath.Dir(p.Path)
			projectCache.DirectoryMap[dir]++
		}

		if err := projectCache.SaveCache(); err != nil && c.debugMode {
			fmt.Printf("Error updating cache: %v\n", err)
		}
	}

	// Create themed delegate
	delegate := ui.CreateThemedDelegate(activeTheme)

	// Create list with themed delegate
	projectList := list.New([]list.Item{}, delegate, 0, 0)
	projectList.Title = cfg.Preferences.ProjectListTitle
	projectList.SetShowHelp(true)
	projectList.SetFilteringEnabled(true)
	projectList.SetShowFilter(true)
	projectList.KeyMap.Filter.SetEnabled(true)
	projectList.KeyMap.ShowFullHelp.SetEnabled(true)
	projectList.KeyMap.CancelWhileFiltering.SetEnabled(true)
	projectList.KeyMap.AcceptWhileFiltering.SetEnabled(true)
	// projectList.DisableQuitKeybindings()

	// Add custom help
	projectList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("a"),
				key.WithHelp("a", "add directory"),
			),
			key.NewBinding(
				key.WithKeys("."),
				key.WithHelp(".", "open config"),
			),
		}
	}

	items := make([]list.Item, len(projects))
	for i, p := range projects {
		items[i] = tui.ListItem{Project: p}
	}
	projectList.SetItems(items)

	model := tui.Model{
		Config:        cfg,
		List:          projectList,
		Projects:      projects,
		TabState:      nil,
		ShowContext:   false,
		ContextCursor: 0,
		InputMode:     len(cfg.ProjectDirs) == 0,
		Styles:        styles,
		KeyMap:        tui.DefaultKeyMap(),
	}

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run UI: %v", err)
	}

	return nil
}