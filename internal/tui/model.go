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
	AddDirectory    key.Binding
	ShowContext     key.Binding
	Up              key.Binding
	Down            key.Binding
	Enter           key.Binding
	Escape          key.Binding
	Filter          key.Binding
	OpenConfig      key.Binding
	ToggleFavorite  key.Binding
	FilterFavorites key.Binding
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
			key.WithHelp("enter", "show context"),
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
			key.WithHelp(".", "config"),
		),
		ToggleFavorite: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "toggle favorite"),
		),
		FilterFavorites: key.NewBinding(
			key.WithKeys("F"),
			key.WithHelp("F", "filter favorites"),
		),
	}
}

// Model represents the application state
type Model struct {
	Projects          []project.Project
	List              list.Model
	Err               error
	Config            *config.Config
	InputMode         bool
	Input             string
	TabState          *TabCompletionState
	Status            string
	ShowContext       bool
	ContextCursor     int
	AddingDir         bool
	Width             int
	Height            int
	Styles            *ui.Styles
	KeyMap            KeyMap
	ShowFavoritesOnly bool
}

// TabCompletionState tracks the state of tab completion
type TabCompletionState struct {
	Suggestions []string
	Index       int
	Page        int
	PageSize    int
}

// ListItem represents an item in the project list
type ListItem struct {
	Project project.Project
}

func (i ListItem) Title() string {
	if i.Project.Favorite {
		return "★ " + i.Project.Name
	}
	return "  " + i.Project.Name
}

func (i ListItem) Description() string {
	desc := i.Project.Path
	if i.Project.GitState != "" {
		desc += " ( " + i.Project.GitState + " )"
	}
	return desc
}

// FilterValue implements list.Item interface
func (i ListItem) FilterValue() string {
	favorite := ""
	if i.Project.Favorite {
		favorite = "favorite starred"
	}
	return fmt.Sprintf("%s %s %s", i.Project.Name, i.Project.Path, favorite)
}

// ContextOptions defines the available context menu options
var ContextOptions = []string{
	// "Go To",
	"Editor",
	"Explorer",
	"Copy Path",
	"Toggle Favorite",
	"Cancel",
}

// InputPlaceholder is the text shown in the input field before user starts typing
const InputPlaceholder = "/path/to/your/projects"

// DefaultPageSize is the number of suggestions shown per page
const DefaultPageSize = 5

// CopyToClipboard copies the given text to the system clipboard
func CopyToClipboard(text string) error {
	return clipboard.WriteAll(text)
}

// ProjectsLoadedMsg is sent when projects are loaded
type ProjectsLoadedMsg []project.Project
