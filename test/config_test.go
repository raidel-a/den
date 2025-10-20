package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pelletier/go-toml/v2"
)

func TestConfigMigration(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Override the home directory for this test
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create config directory
	configDir := filepath.Join(tmpDir, ".config", "den")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Create an old config file without GitStatusStyle
	oldConfigContent := `projectDirs = ["/tmp/test-projects"]

[preferences]
defaultEditor = "vim"
editorList = ["vim", "nano"]
defaultFileManager = "open"
showHiddenFiles = false
showGitStatus = true
theme = "default"
projectListTitle = "My Projects"
`
	configPath := filepath.Join(configDir, "config.toml")
	if err := os.WriteFile(configPath, []byte(oldConfigContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load the config (should trigger migration)
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify GitStatusStyle was set to default
	if cfg.Preferences.GitStatusStyle != "text" {
		t.Errorf("Expected GitStatusStyle to be 'text', got '%s'", cfg.Preferences.GitStatusStyle)
	}

	// Verify other fields were preserved
	if cfg.Preferences.DefaultEditor != "vim" {
		t.Errorf("Expected DefaultEditor to be 'vim', got '%s'", cfg.Preferences.DefaultEditor)
	}
	if cfg.Preferences.ProjectListTitle != "My Projects" {
		t.Errorf("Expected ProjectListTitle to be 'My Projects', got '%s'", cfg.Preferences.ProjectListTitle)
	}

	// Verify the config file was updated on disk
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read updated config: %v", err)
	}

	var reloadedCfg Config
	if err := toml.Unmarshal(data, &reloadedCfg); err != nil {
		t.Fatalf("Failed to parse updated config: %v", err)
	}

	if reloadedCfg.Preferences.GitStatusStyle != "text" {
		t.Errorf("GitStatusStyle was not persisted to disk, got '%s'", reloadedCfg.Preferences.GitStatusStyle)
	}
}

func TestMergeWithDefaults(t *testing.T) {
	// Test that MergeWithDefaults correctly identifies missing fields
	cfg := &Config{
		ProjectDirs: []string{"/test"},
		Preferences: UserPreferences{
			DefaultEditor: "vim",
			Theme:         "dark",
			// GitStatusStyle is missing
		},
	}

	defaults := DefaultConfig()
	merged, didMigrate := MergeWithDefaults(cfg, defaults)

	if !didMigrate {
		t.Error("Expected migration to be detected")
	}

	if merged.Preferences.GitStatusStyle != "text" {
		t.Errorf("Expected GitStatusStyle to be 'text', got '%s'", merged.Preferences.GitStatusStyle)
	}

	// Verify existing values were preserved
	if merged.Preferences.DefaultEditor != "vim" {
		t.Error("Existing DefaultEditor was overwritten")
	}
	if merged.Preferences.Theme != "dark" {
		t.Error("Existing Theme was overwritten")
	}
}

func TestMergeWithDefaults_NoMigrationNeeded(t *testing.T) {
	// Test that no migration is reported when all fields are present
	defaults := DefaultConfig()

	// Create a complete config
	cfg := &Config{
		ProjectDirs: []string{"/test"},
		Preferences: UserPreferences{
			DefaultEditor:      "vim",
			EditorList:         []string{"vim"},
			DefaultFileManager: "open",
			ShowHiddenFiles:    false,
			ShowGitStatus:      true,
			GitStatusStyle:     "nerd",
			Theme:              "dark",
			ProjectListTitle:   "Projects",
		},
	}

	merged, didMigrate := MergeWithDefaults(cfg, defaults)

	if didMigrate {
		t.Error("Expected no migration for complete config")
	}

	// Verify values were preserved
	if merged.Preferences.GitStatusStyle != "nerd" {
		t.Error("GitStatusStyle should not have been changed")
	}
}
