package completion

import (
	"fmt"
	"os"
	"path/filepath"
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
)

// InstallCompletions installs shell completion scripts
func InstallCompletions() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}

	// Install Zsh completion
	zshCompletionDir := filepath.Join(home, ".zsh", "completion")
	if err := os.MkdirAll(zshCompletionDir, 0755); err != nil {
		return fmt.Errorf("failed to create zsh completion directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(zshCompletionDir, "_den"), []byte(zshCompletion), 0644); err != nil {
		return fmt.Errorf("failed to write zsh completion: %v", err)
	}

	// Install Bash completion
	bashCompletionDir := filepath.Join(home, ".bash_completion.d")
	if err := os.MkdirAll(bashCompletionDir, 0755); err != nil {
		return fmt.Errorf("failed to create bash completion directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(bashCompletionDir, "den"), []byte(bashCompletion), 0644); err != nil {
		return fmt.Errorf("failed to write bash completion: %v", err)
	}

	// Install Fish completion
	fishCompletionDir := filepath.Join(home, ".config", "fish", "completions")
	if err := os.MkdirAll(fishCompletionDir, 0755); err != nil {
		return fmt.Errorf("failed to create fish completion directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(fishCompletionDir, "den.fish"), []byte(fishCompletion), 0644); err != nil {
		return fmt.Errorf("failed to write fish completion: %v", err)
	}

	return nil
}
