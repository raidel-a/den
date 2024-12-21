package theme

import "github.com/charmbracelet/lipgloss"

type Theme struct {
    Name            string
    Primary         lipgloss.Color
    Secondary       lipgloss.Color
    Background      lipgloss.Color
    Text            lipgloss.Color
    SelectedText    lipgloss.Color
    Border          lipgloss.Color
    Error           lipgloss.Color
    Success         lipgloss.Color
}

var (
    Themes = map[string]Theme{
        "default": {
            Name:         "default",
            Primary:      lipgloss.Color("205"),  // Pink
            Secondary:    lipgloss.Color("245"),  // Gray
            Background:   lipgloss.Color("0"),    // Black
            Text:        lipgloss.Color("252"),   // Light gray
            SelectedText: lipgloss.Color("255"),   // White
            Border:      lipgloss.Color("205"),   // Pink
            Error:       lipgloss.Color("196"),   // Red
            Success:     lipgloss.Color("46"),    // Green
        },
        "dracula": {
            Name:         "dracula",
            Primary:      lipgloss.Color("141"),  // Purple
            Secondary:    lipgloss.Color("61"),   // Blue
            Background:   lipgloss.Color("236"),  // Dark gray
            Text:        lipgloss.Color("253"),   // Light gray
            SelectedText: lipgloss.Color("255"),  // White
            Border:      lipgloss.Color("141"),   // Purple
            Error:       lipgloss.Color("203"),   // Red
            Success:     lipgloss.Color("84"),    // Green
        },
        "nord": {
            Name:         "nord",
            Primary:      lipgloss.Color("110"),  // Light blue
            Secondary:    lipgloss.Color("109"),  // Blue
            Background:   lipgloss.Color("237"),  // Dark blue-gray
            Text:        lipgloss.Color("254"),   // Off-white
            SelectedText: lipgloss.Color("255"),  // White
            Border:      lipgloss.Color("110"),   // Light blue
            Error:       lipgloss.Color("167"),   // Red
            Success:     lipgloss.Color("108"),   // Green
        },
        "gruvbox": {
            Name:         "gruvbox",
            Primary:      lipgloss.Color("214"),  // Orange
            Secondary:    lipgloss.Color("142"),  // Green
            Background:   lipgloss.Color("235"),  // Dark gray
            Text:        lipgloss.Color("223"),   // Off-white
            SelectedText: lipgloss.Color("229"),  // Light yellow
            Border:      lipgloss.Color("214"),   // Orange
            Error:       lipgloss.Color("167"),   // Red
            Success:     lipgloss.Color("142"),   // Green
        },
        "solarized": {
            Name:         "solarized",
            Primary:      lipgloss.Color("136"),  // Yellow
            Secondary:    lipgloss.Color("37"),   // Cyan
            Background:   lipgloss.Color("234"),  // Dark blue
            Text:        lipgloss.Color("247"),   // Gray
            SelectedText: lipgloss.Color("254"),  // Light gray
            Border:      lipgloss.Color("136"),   // Yellow
            Error:       lipgloss.Color("160"),   // Red
            Success:     lipgloss.Color("64"),    // Green
        },
    }

    Current Theme
)

// GetTheme returns the theme by name or default if not found
func GetTheme(name string) Theme {
    if theme, ok := Themes[name]; ok {
        return theme
    }
    return Themes["default"]
}

// ListThemes returns all available theme names
func ListThemes() []string {
    var names []string
    for name := range Themes {
        names = append(names, name)
    }
    return names
} 