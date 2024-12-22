package ui

import (
	"den/internal/theme"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// Styles holds all the UI styles for the application
type Styles struct {
	Input           lipgloss.Style
	Instruction     lipgloss.Style
	Title           lipgloss.Style
	SelectedItem    lipgloss.Style
	RegularItem     lipgloss.Style
	Context         lipgloss.Style
	MenuItem        lipgloss.Style
	SelectedMenuItem lipgloss.Style
}

// NewStyles creates a new Styles instance with the given theme
func NewStyles(activeTheme theme.Theme) *Styles {
	return &Styles{
		Input: lipgloss.NewStyle().
			Foreground(activeTheme.Primary).
			MarginTop(1),

		Instruction: lipgloss.NewStyle().
			Foreground(activeTheme.Primary).
			Margin(1, 0).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(activeTheme.Border),

		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(activeTheme.Primary).
			Background(activeTheme.Secondary).
			MarginBottom(1),

		SelectedItem: lipgloss.NewStyle().
			Foreground(activeTheme.Primary).
			Margin(0, 2).
			Bold(true),

		RegularItem: lipgloss.NewStyle().
			Foreground(activeTheme.Secondary),

		Context: lipgloss.NewStyle().
			Align(lipgloss.Center),

		MenuItem: lipgloss.NewStyle().
			Faint(true),

		SelectedMenuItem: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("87")),
	}
}

// CreateThemedDelegate creates a new list delegate with themed styles
func CreateThemedDelegate(activeTheme theme.Theme) list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()

	delegate.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(activeTheme.Text)

	delegate.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(activeTheme.Secondary)

	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(activeTheme.Primary).
		Foreground(activeTheme.Primary).
		Padding(0, 1).
		MarginLeft(5).
		Bold(true)

	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(activeTheme.Text).
		Underline(true)

	return delegate
}