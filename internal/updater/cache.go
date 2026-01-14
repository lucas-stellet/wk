package updater

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const (
	cacheTTL      = 24 * time.Hour
	cacheFileName = "update-check.json"
)

// CacheEntry represents a cached update check result.
type CacheEntry struct {
	CheckedAt       time.Time `json:"checked_at"`
	CurrentVersion  string    `json:"current_version"`
	LatestVersion   string    `json:"latest_version"`
	UpdateAvailable bool      `json:"update_available"`
	DownloadURL     string    `json:"download_url"`
	ReleaseURL      string    `json:"release_url"`
}

// CachedCheck returns cached update info if valid, otherwise fetches new info.
func CachedCheck(currentVersion string) (*Info, error) {
	cache, err := loadCache()
	if err == nil && cache.isValid(currentVersion) {
		return cache.toInfo(), nil
	}

	info, err := CheckForUpdate(currentVersion)
	if err != nil {
		return nil, err
	}

	saveCache(info)
	return info, nil
}

// getCacheDir returns the wk cache directory path.
func getCacheDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".wk"), nil
}

// getCachePath returns the full path to the cache file.
func getCachePath() (string, error) {
	dir, err := getCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, cacheFileName), nil
}

// loadCache loads the cached update check from disk.
func loadCache() (*CacheEntry, error) {
	path, err := getCachePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cache CacheEntry
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	return &cache, nil
}

// saveCache saves the update info to cache.
func saveCache(info *Info) error {
	dir, err := getCacheDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	cache := &CacheEntry{
		CheckedAt:       time.Now(),
		CurrentVersion:  info.CurrentVersion,
		LatestVersion:   info.LatestVersion,
		UpdateAvailable: info.UpdateAvailable,
		DownloadURL:     info.DownloadURL,
		ReleaseURL:      info.ReleaseURL,
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}

	path, err := getCachePath()
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// isValid checks if the cache entry is still valid.
func (c *CacheEntry) isValid(currentVersion string) bool {
	if c.CurrentVersion != currentVersion {
		return false
	}
	return time.Since(c.CheckedAt) < cacheTTL
}

// toInfo converts a CacheEntry to Info.
func (c *CacheEntry) toInfo() *Info {
	return &Info{
		CurrentVersion:  c.CurrentVersion,
		LatestVersion:   c.LatestVersion,
		UpdateAvailable: c.UpdateAvailable,
		DownloadURL:     c.DownloadURL,
		ReleaseURL:      c.ReleaseURL,
	}
}

// InvalidateCache removes the cache file.
func InvalidateCache() error {
	path, err := getCachePath()
	if err != nil {
		return err
	}
	return os.Remove(path)
}
