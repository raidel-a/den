package ui

import (
	"den/internal/theme"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// Styles holds all the UI styles for the application
type Styles struct {
	Input            lipgloss.Style
	Instruction      lipgloss.Style
	Title            lipgloss.Style
	ListTitle        lipgloss.Style
	SelectedItem     lipgloss.Style
	RegularItem      lipgloss.Style
	Context          lipgloss.Style
	MenuItem         lipgloss.Style
	SelectedMenuItem lipgloss.Style
	Placeholder      lipgloss.Style
}

// NewStyles creates a new Styles instance with the given theme
func NewStyles(activeTheme theme.Theme) *Styles {
	return &Styles{
		Input: lipgloss.NewStyle().
			Foreground(activeTheme.Primary).
			MarginTop(1),

		Placeholder: lipgloss.NewStyle().
			Foreground(activeTheme.Secondary).
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
			Width(100).
			Align(lipgloss.Center),

		SelectedItem: lipgloss.NewStyle().
			Foreground(activeTheme.Primary).
			Margin(2, 0, 2, 0).
			Bold(true),

		RegularItem: lipgloss.NewStyle().
			Foreground(activeTheme.Secondary),

		Context: lipgloss.NewStyle().
			Margin(1, 0, 0, 2),

		MenuItem: lipgloss.NewStyle().
			Faint(true),

		SelectedMenuItem: lipgloss.NewStyle().
			Bold(true).
			Foreground(activeTheme.Primary),

		ListTitle: lipgloss.NewStyle().
			Bold(true).
			Padding(0,2,0,2).
			Foreground(activeTheme.Primary).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderBottom(true),
	}
}

// CreateThemedDelegate creates a new list delegate with themed styles
func CreateThemedDelegate(activeTheme theme.Theme) list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()

	// Set base styles
	baseStyle := lipgloss.NewStyle()

	// Title styles
	titleStyle := baseStyle.
		Foreground(activeTheme.Text).
		Bold(true)

	delegate.Styles.SelectedTitle = titleStyle.
		BorderStyle(lipgloss.DoubleBorder()).
		BorderForeground(activeTheme.Primary).
		Foreground(activeTheme.Primary).
		Padding(0, 1)

	delegate.Styles.NormalTitle = titleStyle

	// Description styles
	descStyle := baseStyle.
		Foreground(activeTheme.Secondary)

	delegate.Styles.SelectedDesc = descStyle.
		Foreground(activeTheme.Text)

	delegate.Styles.NormalDesc = descStyle

	// Set spacing for better readability
	delegate.SetSpacing(1)

	return delegate
}
