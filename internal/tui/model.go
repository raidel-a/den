package tui

import (
	"den/internal/config"
	"den/internal/project"
	"den/internal/ui"

	"fmt"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
)

// KeyMap defines keybindings for the application
type KeyMap struct {
	AddDirectory key.Binding
	ShowContext  key.Binding
	Up           key.Binding
	Down         key.Binding
	Enter        key.Binding
	Escape       key.Binding
	Filter       key.Binding
	OpenConfig   key.Binding
}

// DefaultKeyMap returns the default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		AddDirectory: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add directory"),
		),
		ShowContext: key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("enter", "show context menu"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		OpenConfig: key.NewBinding(
			key.WithKeys("."),
			key.WithHelp(".", "open config"),
		),
	}
}

// Model represents the application state
type Model struct {
	Projects      []project.Project
	List          list.Model
	Err           error
	Config        *config.Config
	InputMode     bool
	Input         string
	TabState      *TabCompletionState
	Status        string
	ShowContext   bool
	ContextCursor int
	AddingDir     bool
	Width         int
	Height        int
	Styles        *ui.Styles
	KeyMap        KeyMap
}

// TabCompletionState tracks the state of tab completion
type TabCompletionState struct {
	Suggestions []string
	Index       int
}

// ListItem represents an item in the project list
type ListItem struct {
	Project project.Project
}

func (i ListItem) Title() string { return i.Project.Name }
func (i ListItem) Description() string {
	desc := i.Project.Path
	if i.Project.GitState != "" {
		desc += " (" + i.Project.GitState + ")"
	}
	return desc
}

// FilterValue implements list.Item interface
func (i ListItem) FilterValue() string {
	// Return both name and path for filtering
	return fmt.Sprintf("%s %s", i.Project.Name, i.Project.Path)
}

// ContextOptions defines the available context menu options
var ContextOptions = []string{
	"Editor",
	"File Explorer",
	"Go to Dir",
	"Copy Path",
	"Cancel",
}

// CopyToClipboard copies the given text to the system clipboard
func CopyToClipboard(text string) error {
	return clipboard.WriteAll(text)
}

// ProjectsLoadedMsg is sent when projects are loaded
type ProjectsLoadedMsg []project.Project
