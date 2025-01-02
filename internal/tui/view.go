package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the current state of the model
func (m Model) View() string {
	if m.AddingDir {
		return m.renderAddingDirView()
	}

	if m.ShowContext {
		return m.renderContextView()
	}

	return m.List.View()
}

func (m Model) renderAddingDirView() string {
	var s strings.Builder
	s.WriteString(m.Styles.Title.Render("Add Project Directory"))
	s.WriteString("\n")
	s.WriteString(m.Styles.Instruction.Render(
		"Enter the path to your projects directory.\n" +
			"Press Tab to autocomplete, Esc to cancel and Enter to confirm.",
	))
	s.WriteString("\n")

	// Show either input or placeholder
	if m.Input == "" {
		s.WriteString(m.Styles.Placeholder.Render("Enter directory path..."))
	} else {
		s.WriteString(m.Styles.Input.Render(m.Input))
	}

	if m.TabState != nil && len(m.TabState.Suggestions) > 0 {
		s.WriteString("\n\nSuggestions:\n")
		for i, sugg := range m.TabState.Suggestions {
			if i == m.TabState.Index {
				s.WriteString(m.Styles.SelectedItem.Render("> " + sugg + "\n"))
			} else {
				s.WriteString(m.Styles.RegularItem.Render("  " + sugg + "\n"))
			}
		}
	}

	if m.Err != nil {
		s.WriteString("\n\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Render(fmt.Sprintf("Error: %v", m.Err)))
	}

	s.WriteString("\n\n" + m.Styles.RegularItem.Render(
		"Tab: autocomplete • Enter: confirm • Esc: cancel",
	))

	return s.String()
}

func (m Model) renderContextView() string {
	var s strings.Builder
	s.WriteString(m.List.View())

	// Create horizontal menu
	var menuItems []string
	width := 0
	maxWidth := m.List.Width() - 4 // Account for margins

	for i, opt := range ContextOptions {
		item := opt
		if i == m.ContextCursor {
			item = m.Styles.SelectedMenuItem.Render(opt)
		} else {
			item = m.Styles.MenuItem.Render(opt)
		}

		// Check if adding this item would exceed available width
		itemWidth := lipgloss.Width(item) + 1 // +1 for separator
		if width+itemWidth > maxWidth {
			break
		}

		menuItems = append(menuItems, item)
		width += itemWidth
	}

	menu := m.Styles.Context.Render(strings.Join(menuItems, " • "))
	s.WriteString("\n" + menu)

	return s.String()
}
