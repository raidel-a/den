package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type UserPreferences struct {
	DefaultEditor      string   `json:"defaultEditor"`
	EditorList         []string `json:"editorList"`
	DefaultFileManager string   `json:"defaultFileManager"`
	ShowHiddenFiles    bool     `json:"showHiddenFiles"`
	ShowGitStatus      bool     `json:"showGitStatus"`
	Theme              string   `json:"theme"`
	ProjectListTitle   string   `json:"projectListTitle"`
}

type Config struct {
	ProjectDirs []string        `json:"projectDirs"`
	Preferences UserPreferences `json:"preferences"`
}

func DefaultConfig() *Config {
	return &Config{
		ProjectDirs: []string{},
		Preferences: UserPreferences{
			DefaultEditor:      "code", // VS Code as default
			EditorList:         []string{"code", "vim", "nano"},
			DefaultFileManager: "", // Will be set based on OS
			ShowHiddenFiles:    false,
			ShowGitStatus:      true,
			Theme:              "default",
			ProjectListTitle:   "Your Projects",
		},
	}
}

func ensureConfigDir() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get home directory: %v", err)
	}

	configDir := filepath.Join(homeDir, ".config", "den")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("could not create config directory: %v", err)
	}

	return nil
}

func detectSystemPreferences() UserPreferences {
	prefs := UserPreferences{
		ShowGitStatus:    true,
		ShowHiddenFiles:  false,
		Theme:            "default",
		ProjectListTitle: "Your Projects",
	}

	// Detect default editor based on OS
	switch runtime.GOOS {
	case "darwin":
		prefs.DefaultEditor = "code" // VS Code is common on macOS
	case "linux":
		prefs.DefaultEditor = "vim" // Vim is usually available on Linux
	case "windows":
		prefs.DefaultEditor = "notepad"
	default:
		prefs.DefaultEditor = "vim"
	}

	// Detect default file manager
	switch runtime.GOOS {
	case "darwin":
		prefs.DefaultFileManager = "open"
	case "linux":
		prefs.DefaultFileManager = "xdg-open"
	case "windows":
		prefs.DefaultFileManager = "explorer"
	default:
		prefs.DefaultFileManager = "open"
	}

	return prefs
}

func LoadConfig() (*Config, error) {
	if err := ensureConfigDir(); err != nil {
		return nil, err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not get home directory: %v", err)
	}

	configPath := filepath.Join(homeDir, ".config", "den", "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default config if file doesn't exist
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("could not read config file: %v", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("could not parse config file: %v", err)
	}

	// Merge with defaults to ensure all fields are set
	defaultCfg := DefaultConfig()

	// Only override non-empty values from the loaded config
	if cfg.Preferences.DefaultEditor == "" {
		cfg.Preferences.DefaultEditor = defaultCfg.Preferences.DefaultEditor
	}
	if cfg.Preferences.DefaultFileManager == "" {
		cfg.Preferences.DefaultFileManager = defaultCfg.Preferences.DefaultFileManager
	}
	if cfg.Preferences.Theme == "" {
		cfg.Preferences.Theme = defaultCfg.Preferences.Theme
	}
	if cfg.Preferences.ProjectListTitle == "" {
		cfg.Preferences.ProjectListTitle = defaultCfg.Preferences.ProjectListTitle
	}
	if len(cfg.Preferences.EditorList) == 0 {
		cfg.Preferences.EditorList = defaultCfg.Preferences.EditorList
	}

	return &cfg, nil
}

func SaveConfig(cfg *Config) error {
	if err := ensureConfigDir(); err != nil {
		return err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get home directory: %v", err)
	}

	configPath := filepath.Join(homeDir, ".config", "den", "config.json")
	data, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return fmt.Errorf("could not marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("could not write config file: %v", err)
	}

	return nil
}

func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get home directory: %v", err)
	}

	return filepath.Join(homeDir, ".config", "den", "config.json"), nil
}
