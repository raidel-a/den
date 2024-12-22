package main

import (
	"den/internal/cache"
	"den/internal/config"
	"den/internal/project"
	"den/internal/theme"
	"den/internal/tui"
	"den/internal/ui"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

var debugMode bool

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

func initialModel() tui.Model {
	cfg, err := config.LoadConfig()
	if err != nil {
		return tui.Model{
			Err:           fmt.Errorf("failed to load config: %v", err),
			List:          list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
			TabState:      nil,
			ShowContext:   false,
			ContextCursor: 0,
		}
	}

	// Set active theme and create styles
	activeTheme := theme.GetTheme(cfg.Preferences.Theme)
	styles := ui.NewStyles(activeTheme)

	// Try to load from cache first
	projectCache, err := cache.LoadCache()
	if err != nil {
		if debugMode {
			fmt.Printf("Error loading cache: %v\n", err)
		}
	}

	var projects []project.Project
	if projectCache != nil && projectCache.IsCacheValid(cfg) {
		// Use cached projects
		projects = project.ConvertCacheToProjects(projectCache.Projects)
		if debugMode {
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

		if err := projectCache.SaveCache(); err != nil && debugMode {
			fmt.Printf("Error updating cache: %v\n", err)
		}
	}

	// Create themed delegate
	delegate := ui.CreateThemedDelegate(activeTheme)

	// Create list with themed delegate
	projectList := list.New([]list.Item{}, delegate, 0, 0)
	projectList.Title = cfg.Preferences.ProjectListTitle
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

	items := make([]list.Item, len(projects))
	for i, p := range projects {
		items[i] = tui.ListItem{Project: p}
	}
	projectList.SetItems(items)

	return tui.Model{
		Config:        cfg,
		List:          projectList,
		TabState:      nil,
		ShowContext:   false,
		ContextCursor: 0,
		InputMode:     len(cfg.ProjectDirs) == 0,
		Styles:        styles,
	}
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

			// Enable verbose logging
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
