package tui

import (
	"den/internal/config"
	"den/internal/project"
	"den/internal/ui"

	"fmt"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
)

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
