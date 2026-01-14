// Package worktree provides operations for managing git worktrees.
package worktree

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Worktree represents a git worktree entry.
type Worktree struct {
	Path   string
	Commit string
	Branch string
}

// Add creates a new worktree for the given branch.
// Returns the path where the worktree was created.
func Add(branch string) (string, error) {
	cmd := exec.Command("git", "worktree", "add", branch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git worktree add failed: %s", strings.TrimSpace(string(output)))
	}

	// Parse output to find the created path
	// Git outputs: "Preparing worktree (new branch 'branch')" or similar
	// The path is typically "../branch" relative to current dir
	path, err := filepath.Abs(branch)
	if err != nil {
		return "", err
	}

	return path, nil
}

// List returns all worktrees in the repository.
func List() ([]Worktree, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git worktree list failed: %w", err)
	}

	return parseWorktreeList(output)
}

func parseWorktreeList(data []byte) ([]Worktree, error) {
	var worktrees []Worktree
	var current Worktree

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "worktree "):
			if current.Path != "" {
				worktrees = append(worktrees, current)
			}
			current = Worktree{Path: strings.TrimPrefix(line, "worktree ")}
		case strings.HasPrefix(line, "HEAD "):
			current.Commit = strings.TrimPrefix(line, "HEAD ")
		case strings.HasPrefix(line, "branch "):
			branch := strings.TrimPrefix(line, "branch ")
			// Remove refs/heads/ prefix
			current.Branch = strings.TrimPrefix(branch, "refs/heads/")
		case line == "detached":
			current.Branch = "(detached)"
		}
	}

	if current.Path != "" {
		worktrees = append(worktrees, current)
	}

	return worktrees, scanner.Err()
}

// Remove removes a worktree by path or branch name.
func Remove(target string) error {
	cmd := exec.Command("git", "worktree", "remove", target)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree remove failed: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

// GetMainWorktreePath returns the path of the main worktree (bare repo or main checkout).
func GetMainWorktreePath() (string, error) {
	worktrees, err := List()
	if err != nil {
		return "", err
	}

	if len(worktrees) == 0 {
		return "", fmt.Errorf("no worktrees found")
	}

	return worktrees[0].Path, nil
}

// HasUncommittedChanges checks if there are uncommitted changes in the working directory.
func HasUncommittedChanges() (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("git status failed: %w", err)
	}
	return len(bytes.TrimSpace(output)) > 0, nil
}

// GetCurrentBranch returns the name of the current branch.
func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse failed: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// CreateStash creates a stash with the given message.
func CreateStash(message string) error {
	cmd := exec.Command("git", "stash", "push", "-m", message)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git stash failed: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

// FindByBranch finds a worktree by its branch name.
func FindByBranch(branch string) (*Worktree, error) {
	worktrees, err := List()
	if err != nil {
		return nil, err
	}

	for _, wt := range worktrees {
		if wt.Branch == branch {
			return &wt, nil
		}
	}
	return nil, fmt.Errorf("worktree for branch '%s' not found", branch)
}
