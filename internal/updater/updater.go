// Package updater provides self-update functionality for wk.
package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
)

const (
	repoOwner = "lucas-stellet"
	repoName  = "wk"
	apiURL    = "https://api.github.com/repos/" + repoOwner + "/" + repoName + "/releases/latest"
)

// Info contains version and update information.
type Info struct {
	CurrentVersion  string
	LatestVersion   string
	UpdateAvailable bool
	DownloadURL     string
	ReleaseURL      string
}

// githubRelease represents the GitHub API response for a release.
type githubRelease struct {
	TagName string  `json:"tag_name"`
	HTMLURL string  `json:"html_url"`
	Assets  []asset `json:"assets"`
}

type asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// CheckForUpdate queries GitHub API for the latest release and compares versions.
func CheckForUpdate(currentVersion string) (*Info, error) {
	release, err := fetchLatestRelease()
	if err != nil {
		return nil, err
	}

	info := &Info{
		CurrentVersion:  currentVersion,
		LatestVersion:   release.TagName,
		UpdateAvailable: isNewerVersion(currentVersion, release.TagName),
		ReleaseURL:      release.HTMLURL,
	}

	if info.UpdateAvailable {
		info.DownloadURL = findDownloadURL(release.Assets)
	}

	return info, nil
}

// fetchLatestRelease fetches the latest release from GitHub API.
func fetchLatestRelease() (*githubRelease, error) {
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &release, nil
}

// isNewerVersion compares two version strings and returns true if latest is newer.
func isNewerVersion(current, latest string) bool {
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")

	if current == "dev" || current == "" {
		return true
	}

	return latest != current && compareVersions(latest, current) > 0
}

// compareVersions compares two semver strings.
// Returns: 1 if a > b, -1 if a < b, 0 if equal.
func compareVersions(a, b string) int {
	partsA := strings.Split(a, ".")
	partsB := strings.Split(b, ".")

	for i := 0; i < 3; i++ {
		var numA, numB int
		if i < len(partsA) {
			fmt.Sscanf(partsA[i], "%d", &numA)
		}
		if i < len(partsB) {
			fmt.Sscanf(partsB[i], "%d", &numB)
		}

		if numA > numB {
			return 1
		}
		if numA < numB {
			return -1
		}
	}

	return 0
}

// findDownloadURL finds the appropriate download URL for the current OS/arch.
func findDownloadURL(assets []asset) string {
	os := runtime.GOOS
	arch := runtime.GOARCH

	// Map Go arch names to release asset names
	if arch == "amd64" {
		arch = "amd64"
	} else if arch == "arm64" {
		arch = "arm64"
	}

	expected := fmt.Sprintf("wk_")
	suffix := fmt.Sprintf("_%s_%s.tar.gz", os, arch)

	for _, a := range assets {
		if strings.HasPrefix(a.Name, expected) && strings.HasSuffix(a.Name, suffix) {
			return a.BrowserDownloadURL
		}
	}

	return ""
}

// GetAssetFilename returns the expected asset filename for the current platform.
func GetAssetFilename(version string) string {
	version = strings.TrimPrefix(version, "v")
	return fmt.Sprintf("wk_%s_%s_%s.tar.gz", version, runtime.GOOS, runtime.GOARCH)
}
