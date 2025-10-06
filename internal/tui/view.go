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

	// Create gradient/shadow header that fades from solid in middle to light on edges
	titleText := m.Config.Preferences.ProjectListTitle
	// Use full screen width if available, otherwise use title width
	totalWidth := len(titleText) + 6
	if m.Width > 0 {
		totalWidth = m.Width
	}

	// Create gradient pattern: light -> medium -> dark -> solid -> dark -> medium -> light
	var topBar, bottomBar string
	shades := []string{"░", "▒", "▓", "█"}

	for i := 0; i < totalWidth; i++ {
		// Calculate distance from center
		center := totalWidth / 2
		distFromCenter := center - i
		if distFromCenter < 0 {
			distFromCenter = -distFromCenter
		}

		// Map distance to shade (closer to center = darker)
		shadeIndex := 3 - (distFromCenter * 4 / (totalWidth / 2))
		if shadeIndex < 0 {
			shadeIndex = 0
		}
		if shadeIndex > 3 {
			shadeIndex = 3
		}

		topBar += shades[shadeIndex]
		bottomBar += shades[shadeIndex]
	}

	// Center the title text within the full width
	centeredTitle := lipgloss.NewStyle().
		Width(totalWidth).
		Align(lipgloss.Center).
		Render(titleText)

	gradientHeader := m.Styles.ListTitle.Render(
		topBar + "\n" +
		centeredTitle + "\n" +
		bottomBar,
	)

	listView := m.List.View()

	// Add status indicator for favorite filtering
	if m.ShowFavoritesOnly {
		statusMsg := m.Styles.RegularItem.Copy().
			Background(m.Styles.FavoriteIcon.GetForeground()).
			Foreground(lipgloss.Color("0")).
			Padding(0, 1).
			Render("Showing Favorites Only")

		// Add status message at the bottom
		listView = lipgloss.JoinVertical(lipgloss.Left,
			listView,
			"\n"+statusMsg,
		)
	}

	// Center each line of the list individually
	if m.Width > 0 {
		lines := strings.Split(listView, "\n")
		centeredLines := make([]string, len(lines))
		for i, line := range lines {
			centeredLines[i] = lipgloss.NewStyle().
				Width(m.Width).
				Align(lipgloss.Center).
				Render(line)
		}
		listView = strings.Join(centeredLines, "\n")
	}

	view := "\n" + gradientHeader + "\n\n" + listView

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

	view := s.String()

	// Center each line individually
	if m.Width > 0 {
		lines := strings.Split(view, "\n")
		centeredLines := make([]string, len(lines))
		for i, line := range lines {
			centeredLines[i] = lipgloss.NewStyle().
				Width(m.Width).
				Align(lipgloss.Center).
				Render(line)
		}
		view = strings.Join(centeredLines, "\n")
	}

	return view
}

func (m Model) renderContextView() string {
	// Create gradient/shadow header that fades from solid in middle to light on edges
	titleText := m.Config.Preferences.ProjectListTitle
	// Use full screen width if available, otherwise use title width
	totalWidth := len(titleText) + 6
	if m.Width > 0 {
		totalWidth = m.Width
	}

	// Create gradient pattern: light -> medium -> dark -> solid -> dark -> medium -> light
	var topBar, bottomBar string
	shades := []string{"░", "▒", "▓", "█"}

	for i := 0; i < totalWidth; i++ {
		// Calculate distance from center
		center := totalWidth / 2
		distFromCenter := center - i
		if distFromCenter < 0 {
			distFromCenter = -distFromCenter
		}

		// Map distance to shade (closer to center = darker)
		shadeIndex := 3 - (distFromCenter * 4 / (totalWidth / 2))
		if shadeIndex < 0 {
			shadeIndex = 0
		}
		if shadeIndex > 3 {
			shadeIndex = 3
		}

		topBar += shades[shadeIndex]
		bottomBar += shades[shadeIndex]
	}

	// Center the title text within the full width
	centeredTitle := lipgloss.NewStyle().
		Width(totalWidth).
		Align(lipgloss.Center).
		Render(titleText)

	gradientHeader := m.Styles.ListTitle.Render(
		topBar + "\n" +
		centeredTitle + "\n" +
		bottomBar,
	)

	listView := m.List.View()

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
	listView = listView + "\n" + menu

	// Center each line of the list individually
	if m.Width > 0 {
		lines := strings.Split(listView, "\n")
		centeredLines := make([]string, len(lines))
		for i, line := range lines {
			centeredLines[i] = lipgloss.NewStyle().
				Width(m.Width).
				Align(lipgloss.Center).
				Render(line)
		}
		listView = strings.Join(centeredLines, "\n")
	}

	return "\n" + gradientHeader + "\n\n" + listView
}
