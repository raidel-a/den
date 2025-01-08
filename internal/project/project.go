package project

import (
	"den/internal/cache"
	"den/internal/config"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Project represents a development project
type Project struct {
	Name     string
	Path     string
	LastMod  string
	GitState string
	Favorite bool
}

// DetectProject attempts to identify a project at the given path
func DetectProject(path string, config *config.Config) (*Project, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("not a directory")
	}

	name := filepath.Base(path)
	lastMod := info.ModTime().Format("2006-01-02 15:04:05")

	gitState := "no git"
	if config.Preferences.ShowGitStatus {
		if _, err := os.Stat(filepath.Join(path, ".git")); err == nil {
			gitState = getGitStatus(path)
		}
	}

	// Check if project is in favorites
	favorite := false
	for _, favPath := range config.Favorites {
		if favPath == path {
			favorite = true
			break
		}
	}

	return &Project{
		Name:     name,
		Path:     path,
		LastMod:  lastMod,
		GitState: gitState,
		Favorite: favorite,
	}, nil
}

// ScanForProjects scans directories for projects
func ScanForProjects(dirs []string, config *config.Config) []Project {
	var projects []Project

	for _, dir := range dirs {
		// Check if directory exists and is accessible
		if _, err := os.Stat(dir); err != nil {
			continue
		}

		// Read directory entries
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		// Check each entry
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			fullPath := filepath.Join(dir, entry.Name())
			project, err := DetectProject(fullPath, config)
			if err != nil {
				continue
			}

			projects = append(projects, *project)
		}
	}

	return projects
}

// HasProjectFile checks if the directory contains common project files
func HasProjectFile(path string) bool {
	projectFiles := []string{
		"package.json",     // Node.js
		"Cargo.toml",       // Rust
		"go.mod",           // Go
		"requirements.txt", // Python
		"pom.xml",          // Java/Maven
		"build.gradle",     // Java/Gradle
		"Gemfile",          // Ruby
		"composer.json",    // PHP
	}

	for _, file := range projectFiles {
		if _, err := os.Stat(filepath.Join(path, file)); err == nil {
			return true
		}
	}
	return false
}

func getGitStatus(path string) string {
	cmd := exec.Command("git", "-C", path, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return "no git"
	}

	if len(output) > 0 {
		return "git (modified)"
	}
	return "git (clean)"
}

// ConvertCacheToProjects converts cached projects to Project structs
func ConvertCacheToProjects(cached []cache.Project) []Project {
	projects := make([]Project, len(cached))
	for i, p := range cached {
		projects[i] = Project{
			Name:     p.Name,
			Path:     p.Path,
			LastMod:  p.LastMod.Format("2006-01-02 15:04:05"),
			GitState: p.GitState,
			Favorite: p.Favorite,
		}
	}
	return projects
}

// ConvertProjectsToCache converts Project structs to cached projects
func ConvertProjectsToCache(projects []Project) []cache.Project {
	cached := make([]cache.Project, len(projects))
	for i, p := range projects {
		lastMod, _ := time.Parse("2006-01-02 15:04:05", p.LastMod)
		cached[i] = cache.Project{
			Name:     p.Name,
			Path:     p.Path,
			LastMod:  lastMod,
			GitState: p.GitState,
			Favorite: p.Favorite,
		}
	}
	return cached
}
