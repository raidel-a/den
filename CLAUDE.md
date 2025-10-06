# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Den is a terminal-based project manager built with Go and the Charm Bracelet TUI framework (Bubbletea, Bubbles, Lipgloss). It provides an interactive interface for discovering, navigating, and managing development projects across configured directories.

## Build and Development Commands

```bash
# Build the project
go build

# Development with hot reload
./scripts/dev.sh  # Uses air for live reload

# Install shell completions, man pages, and shell integration
./den --install

# Run with debug logging
./den --debug

# Reset configuration
./den --reset

# Create test environment
./scripts/test-setup.sh
```

## Architecture

### Core Application Flow

1. **Entry Point** (`main.go`): Initializes CLI and starts the application
2. **CLI Layer** (`internal/cli/cli.go`): Handles command-line flags, version info, and bootstraps the TUI
3. **TUI Layer** (`internal/tui/`): Implements the Bubbletea Model-View-Update pattern
4. **Business Logic**: Project scanning, config management, caching

### Key Components

**TUI (internal/tui/)**
- `model.go`: Defines application state (Model struct), keybindings (KeyMap), and list items
- `update.go`: Handles all state transitions via the Update function - processes user input, context menus, directory addition, and tab completion
- `view.go`: Renders the UI based on current state

**Configuration (internal/config/config.go)**
- Uses TOML format stored in `~/.config/den/config.toml`
- Manages ProjectDirs, Favorites, and UserPreferences
- Auto-detects system defaults for editor and file manager

**Caching (internal/cache/cache.go)**
- Stores project metadata in JSON at `~/.cache/den/projects.json`
- Invalidates after 1 hour or when directories change
- Improves startup performance by avoiding repeated directory scans

**Project Scanning (internal/project/project.go)**
- Scans configured directories for subdirectories
- Detects Git repositories and their status (clean/modified/no git)
- Tracks favorites and last modification times

**Theme System (internal/theme/theme.go)**
- Five built-in themes: Default, Dracula, Nord, Gruvbox, Solarized
- Themes are applied to all UI components via lipgloss styles

### State Management Pattern

Den uses the Elm Architecture (Model-View-Update):
- **Model**: Application state including projects, config, UI state (context menu, input mode, filter state)
- **Update**: Pure functions that transform state based on messages (keyboard input, window resize, project loading)
- **View**: Renders UI from current model state

### Important State Modes

The TUI has several mutually exclusive modes:
- **Normal Mode**: Browse project list, trigger actions
- **AddingDir Mode**: Adding a new project directory with tab completion
- **ShowContext Mode**: Display context menu for project actions
- **Filtering Mode**: Handled by the bubbles list component

Mode transitions are managed in `update.go` with careful checking to prevent state conflicts.

### Shell Integration

Den requires shell integration for the "Change Directory" feature to work properly:
- Integration scripts are injected into shell RC files (`~/.bashrc`, `~/.zshrc`, `~/.config/fish/config.fish`)
- The `--install` command handles this setup
- Shell completion is also installed for Bash, Zsh, and Fish

### Project Actions

Context menu actions (`internal/tui/model.go` ContextOptions):
1. **Editor**: Opens project in configured editor (respects $EDITOR env var)
2. **Explorer**: Opens in system file manager
3. **Copy Path**: Copies project path to clipboard
4. **Toggle Favorite**: Adds/removes from favorites list

The "Go To" action is implemented via shell integration and requires the shell function to be loaded.

## Configuration File Structure

```toml
projectDirs = ["/path/to/projects"]

favorites = ["/path/to/favorite/project"]

[preferences]
defaultEditor = "code"
editorList = ["code", "vim", "nano"]
defaultFileManager = "open"  # macOS: "open", Linux: "xdg-open", Windows: "explorer"
showHiddenFiles = false
showGitStatus = true
theme = "default"  # Options: default, dracula, nord, gruvbox, solarized
projectListTitle = "Projects"
```

## Adding New Features

**To add a new keybinding:**
1. Add to KeyMap struct in `internal/tui/model.go`
2. Initialize in DefaultKeyMap()
3. Handle in Update() in `internal/tui/update.go`
4. Add to AdditionalFullHelpKeys if it should appear in help

**To add a new theme:**
1. Define Theme struct in `internal/theme/theme.go`
2. Add to the themes map in GetTheme()
3. Theme will automatically be available via config

**To add a new context menu action:**
1. Add to ContextOptions slice in `internal/tui/model.go`
2. Add case in handleContextMenuSelection() in `internal/tui/update.go`

## Linux-Specific Requirements

Clipboard functionality requires `xsel` or `xclip`:
```bash
sudo apt-get install xsel  # or xclip
```
