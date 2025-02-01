package completion

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	zshCompletion = `#compdef den

_den() {
    local -a commands
    commands=(
        'help:Show help information'
        'version:Show version information'
        'reset:Reset configuration'
        'debug:Enable debug mode'
    )

    _arguments -C \
        '--help[Show help information]' \
        '--version[Show version information]' \
        '--reset[Reset configuration]' \
        '--debug[Enable debug mode]' \
        '*:: :->args'

    case $state in
        args)
            _describe -t commands 'den commands' commands
            ;;
    esac
}

_den "$@"`

	bashCompletion = `_den() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    opts="--help --version --reset --debug"

    if [[ ${cur} == -* ]] ; then
        COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
        return 0
    fi
}
complete -F _den den`

	fishCompletion = `complete -c den -l help -d 'Show help information'
complete -c den -l version -d 'Show version information'
complete -c den -l reset -d 'Reset configuration'
complete -c den -l debug -d 'Enable debug mode'`

	// Shell function for directory changing
	zshFunction = `
# Den shell integration
den() {
    command den "$@"
    local tmpfile="${TMPDIR:-/tmp}/den_last_dir"
    if [ -f "$tmpfile" ]; then
        cd "$(cat $tmpfile)"
        rm -f "$tmpfile"
    fi
}
`

	fishFunction = `
# Den shell integration
function den
    command den $argv
    set -l tmpfile "/tmp/den_last_dir"
    if test -f "$tmpfile"
        cd (cat $tmpfile)
        rm -f "$tmpfile"
    end
end
`

	bashFunction = `
# Den shell integration
den() {
    command den "$@"
    local tmpfile="${TMPDIR:-/tmp}/den_last_dir"
    if [ -f "$tmpfile" ]; then
        cd "$(cat $tmpfile)"
        rm -f "$tmpfile"
    fi
}
`
)

// InstallCompletions installs shell completion scripts
func InstallCompletions(shells map[string]bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}

	if shells["Fish"] {
		if err := installFishCompletions(home); err != nil {
			return err
		}
	}
	if shells["Bash"] {
		if err := installBashCompletions(home); err != nil {
			return err
		}
	}
	if shells["Zsh"] {
		if err := installZshCompletions(home); err != nil {
			return err
		}
	}
	return nil
}

func InstallShellIntegration(home string, shells map[string]bool) error {
	if shells["Bash"] {
		if err := installBashIntegration(home); err != nil {
			return err
		}
	}
	if shells["Zsh"] {
		if err := installZshIntegration(home); err != nil {
			return err
		}
	}
	if shells["Fish"] {
		if err := installFishIntegration(home); err != nil {
			return err
		}
	}
	return nil
}

func appendToFileIfNotExists(filepath string, content string) error {
	// Read existing file
	existing, err := os.ReadFile(filepath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Check if integration is already installed
	if strings.Contains(string(existing), "Den shell integration") {
		return nil
	}

	// Append integration
	f, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(content)
	return err
}

func ensureDirectoryWithSudo(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		// If permission denied, try with sudo
		if os.IsPermission(err) {
			cmd := exec.Command("sudo", "mkdir", "-p", path)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		}
		return err
	}
	return nil
}

func installFishCompletions(home string) error {
	fishCompletionDir := filepath.Join(home, ".config", "fish", "completions")
	if err := ensureDirectoryWithSudo(fishCompletionDir); err != nil {
		return fmt.Errorf("failed to create fish completion directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(fishCompletionDir, "den.fish"), []byte(fishCompletion), 0644); err != nil {
		return fmt.Errorf("failed to write fish completion: %v", err)
	}
	return nil
}

func installBashCompletions(home string) error {
	bashCompletionDir := filepath.Join(home, ".bash_completion.d")
	if err := ensureDirectoryWithSudo(bashCompletionDir); err != nil {
		return fmt.Errorf("failed to create bash completion directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(bashCompletionDir, "den"), []byte(bashCompletion), 0644); err != nil {
		return fmt.Errorf("failed to write bash completion: %v", err)
	}
	return nil
}

func installZshCompletions(home string) error {
	zshCompletionDir := filepath.Join(home, ".zsh", "completion")
	if err := ensureDirectoryWithSudo(zshCompletionDir); err != nil {
		return fmt.Errorf("failed to create zsh completion directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(zshCompletionDir, "_den"), []byte(zshCompletion), 0644); err != nil {
		return fmt.Errorf("failed to write zsh completion: %v", err)
	}
	return nil
}

func installBashIntegration(home string) error {
	bashrcPath := filepath.Join(home, ".bashrc")
	return appendToFileIfNotExists(bashrcPath, bashFunction)
}

func installZshIntegration(home string) error {
	zshrcPath := filepath.Join(home, ".config/zsh/.zshrc")
	if _, err := os.Stat(zshrcPath); os.IsNotExist(err) {
		zshrcPath = filepath.Join(home, ".zshrc")
	}
	return appendToFileIfNotExists(zshrcPath, zshFunction)
}

func installFishIntegration(home string) error {
	fishConfigPath := filepath.Join(home, ".config", "fish", "config.fish")
	return appendToFileIfNotExists(fishConfigPath, fishFunction)
}
