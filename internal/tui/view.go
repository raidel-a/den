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

	view := m.List.View()

	// Add status indicator for favorite filtering
	if m.ShowFavoritesOnly {
		statusMsg := m.Styles.RegularItem.Copy().
			Background(m.Styles.FavoriteIcon.GetForeground()).
			Foreground(lipgloss.Color("0")).
			Padding(0, 1).
			Render("Showing Favorites Only")

		// Add status message at the bottom
		view = lipgloss.JoinVertical(lipgloss.Left,
			view,
			"\n"+statusMsg,
		)
	}

	return view
}

func (m Model) renderAddingDirView() string {
	var s strings.Builder
	s.WriteString(m.Styles.Title.Render("Add Project Directory\n"))
	s.WriteString(m.Styles.Instruction.Render(
		"Enter the path to your projects directory.\n" +
			"Press Tab to autocomplete, Esc to cancel, Enter to confirm.",
	))

	// Show placeholder text in a dimmed style if input is empty
	inputText := m.Input
	if inputText == "" {
		inputText = m.Styles.Placeholder.Render(InputPlaceholder)
	}
	s.WriteString(m.Styles.Input.Render(inputText))

	if m.TabState != nil && len(m.TabState.Suggestions) > 0 {
		s.WriteString("\nSuggestions:")

		// Calculate pagination
		start := m.TabState.Page * m.TabState.PageSize
		end := start + m.TabState.PageSize
		if end > len(m.TabState.Suggestions) {
			end = len(m.TabState.Suggestions)
		}

		// Show page info if there are multiple pages
		totalPages := (len(m.TabState.Suggestions) + m.TabState.PageSize - 1) / m.TabState.PageSize
		if totalPages > 1 {
			s.WriteString(fmt.Sprintf(" (Page %d/%d)", m.TabState.Page+1, totalPages))
		}

		// Show suggestions for current page
		for i := start; i < end; i++ {
			if i == m.TabState.Index {
				s.WriteString("\n" + m.Styles.SelectedItem.Render("> "+m.TabState.Suggestions[i]))
			} else {
				s.WriteString("\n" + m.Styles.RegularItem.Render("  "+m.TabState.Suggestions[i]))
			}
		}
	}

	if m.Err != nil {
		s.WriteString("\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Render(fmt.Sprintf("Error: %v", m.Err)))
	}

	s.WriteString("\n\n" + m.Styles.RegularItem.Render(
		"Enter: confirm • Tab: complete • ↑/↓: navigate • ←/→: more • Esc: cancel",
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
