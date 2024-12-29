package man

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const manPageTemplate = `.TH DEN 1 "%s" "den %s" "Den Manual"
.SH NAME
den \- A Cozy Home for Your Repos
.SH SYNOPSIS
.B den
[\fB\-\-help\fR]
[\fB\-\-version\fR]
[\fB\-\-reset\fR]
[\fB\-\-debug\fR]
.SH DESCRIPTION
.B den
is a terminal-based repository manager that provides a comfortable interface for managing and navigating your Git repositories.
.SH OPTIONS
.TP
.BR \-h ", " \-\-help
Display help information and exit.
.TP
.BR \-v ", " \-\-version
Display version information and exit.
.TP
.BR \-\-reset
Reset all configuration and start fresh.
.TP
.BR \-\-debug
Enable debug logging.
.SH FILES
.TP
.I ~/.config/den/config.yaml
Configuration file
.TP
.I ~/.cache/den/
Cache directory
.SH EXAMPLES
.TP
Start Den's interactive UI:
.B den
.TP
Reset configuration:
.B den --reset
.TP
Enable debug mode:
.B den --debug
.SH BUGS
Report bugs at: https://github.com/yourusername/den/issues
.SH AUTHOR
Your Name <your.email@example.com>
`

// InstallManPage generates and installs the man page
func InstallManPage(version string) error {
	// Format man page with current date and version
	manPage := fmt.Sprintf(manPageTemplate, time.Now().Format("January 2006"), version)

	// Determine man page installation directory
	manDir := "/usr/local/share/man/man1"
	if os.Getuid() != 0 {
		// If not root, install in user's home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %v", err)
		}
		manDir = filepath.Join(home, ".local", "share", "man", "man1")
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(manDir, 0755); err != nil {
		return fmt.Errorf("failed to create man page directory: %v", err)
	}

	// Write man page
	manPath := filepath.Join(manDir, "den.1")
	if err := os.WriteFile(manPath, []byte(manPage), 0644); err != nil {
		return fmt.Errorf("failed to write man page: %v", err)
	}

	return nil
}
