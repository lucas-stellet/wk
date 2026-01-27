// Package worktree provides operations for managing git worktrees.
package worktree

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
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
// If the branch doesn't exist, it creates a new branch from HEAD.
// Returns the path where the worktree was created.
// Worktrees are created in the standard location: ../<reponame>.worktrees/<branch>
func Add(branch string) (string, error) {
	worktreesDir, err := GetWorktreesDir()
	if err != nil {
		return "", err
	}

	// Create worktrees directory if it doesn't exist
	if err := os.MkdirAll(worktreesDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create worktrees directory: %w", err)
	}

	worktreePath := filepath.Join(worktreesDir, branch)

	var cmd *exec.Cmd
	if BranchExists(branch) {
		// Branch exists, just create worktree
		cmd = exec.Command("git", "worktree", "add", worktreePath, branch)
	} else {
		// Branch doesn't exist, create it from HEAD
		cmd = exec.Command("git", "worktree", "add", "-b", branch, worktreePath, "HEAD")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git worktree add failed: %s", strings.TrimSpace(string(output)))
	}

	return worktreePath, nil
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
// If force is true, removes even if worktree has uncommitted changes.
func Remove(target string, force bool) error {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, target)

	cmd := exec.Command("git", args...)
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

// BranchExists checks if a branch exists locally or as a remote tracking branch.
func BranchExists(branch string) bool {
	// Check local branch
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	if cmd.Run() == nil {
		return true
	}

	// Check remote tracking branch
	cmd = exec.Command("git", "show-ref", "--verify", "--quiet", "refs/remotes/origin/"+branch)
	return cmd.Run() == nil
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

// GetRepoName returns the repository name from the remote origin URL or directory name.
func GetRepoName() (string, error) {
	// Try to get from remote origin
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	output, err := cmd.Output()
	if err == nil {
		url := strings.TrimSpace(string(output))
		name := extractRepoName(url)
		if name != "" {
			return name, nil
		}
	}

	// Fallback: use main worktree directory name
	mainPath, err := GetMainWorktreePath()
	if err != nil {
		return "", err
	}
	return filepath.Base(mainPath), nil
}

// extractRepoName extracts the repository name from a git URL.
// Supports both SSH (git@github.com:user/repo.git) and HTTPS (https://github.com/user/repo.git) formats.
func extractRepoName(url string) string {
	// Remove trailing .git
	url = strings.TrimSuffix(url, ".git")

	// Handle SSH format: git@github.com:user/repo
	if strings.Contains(url, ":") && strings.Contains(url, "@") {
		parts := strings.Split(url, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}

	// Handle HTTPS format: https://github.com/user/repo
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return ""
}

// GetWorktreesDir returns the path to the .worktrees directory.
func GetWorktreesDir() (string, error) {
	repoName, err := GetRepoName()
	if err != nil {
		return "", err
	}

	mainPath, err := GetMainWorktreePath()
	if err != nil {
		return "", err
	}

	// ../reponame.worktrees
	parentDir := filepath.Dir(mainPath)
	return filepath.Join(parentDir, repoName+".worktrees"), nil
}

// IsInStandardLocation checks if a worktree path follows the standard pattern.
func IsInStandardLocation(wtPath string) (bool, error) {
	worktreesDir, err := GetWorktreesDir()
	if err != nil {
		return false, err
	}

	// Main worktree doesn't need to be in standard location
	mainPath, _ := GetMainWorktreePath()
	if wtPath == mainPath {
		return true, nil
	}

	return strings.HasPrefix(wtPath, worktreesDir), nil
}

// Move moves a worktree to the standard location.
func Move(wt Worktree) (string, error) {
	worktreesDir, err := GetWorktreesDir()
	if err != nil {
		return "", err
	}

	// Create worktrees directory if it doesn't exist
	if err := os.MkdirAll(worktreesDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create worktrees directory: %w", err)
	}

	newPath := filepath.Join(worktreesDir, wt.Branch)

	cmd := exec.Command("git", "worktree", "move", wt.Path, newPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git worktree move failed: %s", strings.TrimSpace(string(output)))
	}

	return newPath, nil
}

// Branch represents a git branch with metadata.
type Branch struct {
	Name        string
	IsRemote    bool
	IsLocal     bool
	CommitShort string
	CommitDate  string
}

// ListBranches returns all branches (local and remote) with metadata.
func ListBranches() ([]Branch, error) {
	// Get local branches with commit info
	// Format: %(refname:short)|%(objectname:short)|%(committerdate:relative)
	cmd := exec.Command("git", "for-each-ref",
		"--format=%(refname:short)|%(objectname:short)|%(committerdate:relative)",
		"refs/heads/", "refs/remotes/origin/")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git for-each-ref failed: %w", err)
	}

	localBranches := make(map[string]bool)
	remoteBranches := make(map[string]Branch)
	var branches []Branch

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "|", 3)
		if len(parts) != 3 {
			continue
		}

		name := parts[0]
		commitShort := parts[1]
		commitDate := parts[2]

		if strings.HasPrefix(name, "origin/") {
			// Remote branch
			remoteName := strings.TrimPrefix(name, "origin/")
			if remoteName == "HEAD" {
				continue
			}
			remoteBranches[remoteName] = Branch{
				Name:        remoteName,
				IsRemote:    true,
				IsLocal:     false,
				CommitShort: commitShort,
				CommitDate:  commitDate,
			}
		} else {
			// Local branch
			localBranches[name] = true
			branches = append(branches, Branch{
				Name:        name,
				IsRemote:    false,
				IsLocal:     true,
				CommitShort: commitShort,
				CommitDate:  commitDate,
			})
		}
	}

	// Mark local branches that also exist on remote
	for i, b := range branches {
		if _, exists := remoteBranches[b.Name]; exists {
			branches[i].IsRemote = true
			delete(remoteBranches, b.Name)
		}
	}

	// Add remaining remote-only branches
	for _, b := range remoteBranches {
		branches = append(branches, b)
	}

	return branches, scanner.Err()
}

// ListWorktreeBranches returns the branch names that have existing worktrees.
func ListWorktreeBranches() (map[string]bool, error) {
	worktrees, err := List()
	if err != nil {
		return nil, err
	}

	branches := make(map[string]bool)
	for _, wt := range worktrees {
		if wt.Branch != "" && wt.Branch != "(detached)" {
			branches[wt.Branch] = true
		}
	}
	return branches, nil
}
