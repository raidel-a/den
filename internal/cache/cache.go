package cache

import (
	"den/internal/config"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type ProjectCache struct {
	Projects     []Project      `json:"projects"`
	LastUpdated  time.Time      `json:"lastUpdated"`
	DirectoryMap map[string]int `json:"directoryMap"` // maps directory to number of projects
}

type Project struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	LastMod  time.Time `json:"lastMod"`
	GitState string    `json:"gitState"`
}

func GetCachePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".cache", "den", "projects.json"), nil
}

func LoadCache() (*ProjectCache, error) {
	cachePath, err := GetCachePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &ProjectCache{
				Projects:     []Project{},
				LastUpdated:  time.Time{},
				DirectoryMap: make(map[string]int),
			}, nil
		}
		return nil, err
	}

	var cache ProjectCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	return &cache, nil
}

func SaveCache(cache *ProjectCache) error {
	cachePath, err := GetCachePath()
	if err != nil {
		return err
	}

	// Ensure cache directory exists
	cacheDir := filepath.Dir(cachePath)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cache, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(cachePath, data, 0644)
}

func IsCacheValid(cache *ProjectCache, cfg *config.Config) bool {
	if cache == nil || len(cache.Projects) == 0 {
		return false
	}

	// Check if directories have changed
	for _, dir := range cfg.ProjectDirs {
		if _, err := os.Stat(dir); err != nil {
			return false
		}
	}

	// Check if cache is too old (e.g., more than 1 hour)
	if time.Since(cache.LastUpdated) > time.Hour {
		return false
	}

	return true
}

func (c *ProjectCache) IsCacheValid(cfg *config.Config) bool {
	if c == nil || len(c.Projects) == 0 {
		return false
	}

	// Check if directories have changed
	for _, dir := range cfg.ProjectDirs {
		if _, err := os.Stat(dir); err != nil {
			return false
		}
	}

	// Check if cache is too old (e.g., more than 1 hour)
	if time.Since(c.LastUpdated) > time.Hour {
		return false
	}

	return true
}

func (c *ProjectCache) SaveCache() error {
	cachePath, err := GetCachePath()
	if err != nil {
		return err
	}

	// Ensure cache directory exists
	cacheDir := filepath.Dir(cachePath)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(cachePath, data, 0644)
}
