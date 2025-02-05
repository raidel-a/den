package editor

import (
	"den/internal/config"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

var defaultEditors = map[string][]string{
	"darwin":  {"code", "vim", "nano"},     // macOS
	"linux":   {"code", "vim", "nano"},     // Linux
	"windows": {"code.exe", "notepad.exe"}, // Windows
}

// OpenInEditor opens the given path in the configured editor
func OpenInEditor(path string, config *config.Config) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		// Use configured default editor
		editor = config.Preferences.DefaultEditor
		if editor == "" {
			// Try to find a default editor from the configured list
			for _, ed := range config.Preferences.EditorList {
				if _, err := exec.LookPath(ed); err == nil {
					editor = ed
					break
				}
			}
			if editor == "" {
				return fmt.Errorf("no editor found")
			}
		}
	}

	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// OpenInFileExplorer opens the given path in the system's file explorer
func OpenInFileExplorer(path string, config *config.Config) error {
	var cmd *exec.Cmd

	// Use configured file manager if set
	if config.Preferences.DefaultFileManager != "" {
		cmd = exec.Command(config.Preferences.DefaultFileManager, path)
	} else {
		// Fall back to OS defaults
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", path)
		case "windows":
			cmd = exec.Command("explorer", path)
		default: // Linux and others
			cmd = exec.Command("xdg-open", path)
		}
	}

	return cmd.Run()
}

// DetectSystemPreferences detects system-specific editor and file manager preferences
func DetectSystemPreferences() config.UserPreferences {
	prefs := config.UserPreferences{
		ShowGitStatus:    true,
		ShowHiddenFiles:  false,
		Theme:            "default",
		ProjectListTitle: "Projects",
	}

	// Detect default editor
	if editor := os.Getenv("EDITOR"); editor != "" {
		prefs.DefaultEditor = editor
	} else {
		// Try to find common editors
		for _, ed := range defaultEditors[runtime.GOOS] {
			if _, err := exec.LookPath(ed); err == nil {
				prefs.DefaultEditor = ed
				break
			}
		}
	}

	// Detect default file manager
	switch runtime.GOOS {
	case "darwin":
		prefs.DefaultFileManager = "open"
	case "windows":
		prefs.DefaultFileManager = "explorer"
	default:
		if _, err := exec.LookPath("xdg-open"); err == nil {
			prefs.DefaultFileManager = "xdg-open"
		}
	}

	return prefs
}