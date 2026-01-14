// Package hooks handles file copying and post-hook execution.
package hooks

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// CopyFiles copies files and directories from src to dst.
func CopyFiles(src, dst string, files []string) error {
	for _, file := range files {
		srcPath := filepath.Join(src, file)
		dstPath := filepath.Join(dst, file)

		info, err := os.Stat(srcPath)
		if os.IsNotExist(err) {
			fmt.Printf("  skipping %s (not found)\n", file)
			continue
		}
		if err != nil {
			return fmt.Errorf("stat %s: %w", srcPath, err)
		}

		if info.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return fmt.Errorf("copy directory %s: %w", file, err)
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return fmt.Errorf("copy file %s: %w", file, err)
			}
		}
		fmt.Printf("  copied %s\n", file)
	}
	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath)
	})
}

// RunPostHooks executes commands in the specified directory.
func RunPostHooks(dir string, commands []string) error {
	for _, cmdStr := range commands {
		fmt.Printf("  running: %s\n", cmdStr)

		cmd := exec.Command("sh", "-c", cmdStr)
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("command %q failed: %w", cmdStr, err)
		}
	}
	return nil
}
