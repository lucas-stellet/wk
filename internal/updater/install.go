package updater

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// InstallMethod represents how wk was installed.
type InstallMethod string

const (
	InstallMethodBinary   InstallMethod = "binary"
	InstallMethodHomebrew InstallMethod = "homebrew"
	InstallMethodGo       InstallMethod = "go"
	InstallMethodUnknown  InstallMethod = "unknown"
)

// DetectInstallMethod returns how wk was installed.
func DetectInstallMethod() InstallMethod {
	execPath, err := os.Executable()
	if err != nil {
		return InstallMethodUnknown
	}

	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return InstallMethodUnknown
	}

	// Check for Homebrew
	if isHomebrewInstall(execPath) {
		return InstallMethodHomebrew
	}

	// Check for Go install
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		home, _ := os.UserHomeDir()
		gopath = filepath.Join(home, "go")
	}
	if strings.HasPrefix(execPath, filepath.Join(gopath, "bin")) {
		return InstallMethodGo
	}

	// Default to binary install
	if strings.HasPrefix(execPath, "/usr/local/bin") {
		return InstallMethodBinary
	}

	return InstallMethodUnknown
}

// isHomebrewInstall checks if the binary is installed via Homebrew.
func isHomebrewInstall(execPath string) bool {
	// Check if path contains Cellar or homebrew (Homebrew's storage locations)
	if strings.Contains(execPath, "Cellar") || strings.Contains(execPath, "homebrew") {
		return true
	}

	// Don't use brew list command - it checks if any wk is installed via brew,
	// not if THIS specific binary is from brew
	return false
}

// PerformUpdate downloads and installs the new version.
func PerformUpdate(info *Info) error {
	if info.DownloadURL == "" {
		return fmt.Errorf("no download URL available for your platform")
	}

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "wk-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Download the archive
	archivePath := filepath.Join(tmpDir, "wk.tar.gz")
	if err := downloadFile(info.DownloadURL, archivePath); err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}

	// Extract the binary
	newBinaryPath := filepath.Join(tmpDir, "wk")
	if err := extractBinary(archivePath, newBinaryPath); err != nil {
		return fmt.Errorf("failed to extract update: %w", err)
	}

	// Create backup of current binary
	backupPath := execPath + ".backup"
	if err := os.Rename(execPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	// Move new binary to target location
	if err := copyFile(newBinaryPath, execPath); err != nil {
		// Rollback on failure
		os.Rename(backupPath, execPath)
		return fmt.Errorf("failed to install new binary: %w", err)
	}

	// Make executable
	if err := os.Chmod(execPath, 0755); err != nil {
		// Rollback on failure
		os.Remove(execPath)
		os.Rename(backupPath, execPath)
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Verify the new binary works
	cmd := exec.Command(execPath, "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Rollback on failure
		os.Remove(execPath)
		os.Rename(backupPath, execPath)
		return fmt.Errorf("new binary verification failed, rolled back: %w\nOutput: %s", err, string(output))
	}

	// Remove backup
	os.Remove(backupPath)

	// Invalidate cache after successful update
	InvalidateCache()

	return nil
}

// downloadFile downloads a file from URL to the specified path.
func downloadFile(url, destPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// extractBinary extracts the wk binary from a tar.gz archive.
func extractBinary(archivePath, destPath string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if header.Name == "wk" && header.Typeflag == tar.TypeReg {
			out, err := os.Create(destPath)
			if err != nil {
				return err
			}
			defer out.Close()

			if _, err := io.Copy(out, tr); err != nil {
				return err
			}
			return nil
		}
	}

	return fmt.Errorf("wk binary not found in archive")
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
