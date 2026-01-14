package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/lucas-stellet/wk/internal/worktree"
)

var switchCmd = &cobra.Command{
	Use:   "switch [branch]",
	Short: "Switch to another worktree",
	Long: `Switch to another worktree by opening a new shell in its directory.

If branch is not specified, shows a list of available worktrees to choose from.
If there are uncommitted changes, offers to stash them before switching.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSwitch,
}

func init() {
	rootCmd.AddCommand(switchCmd)
}

func runSwitch(cmd *cobra.Command, args []string) error {
	worktrees, err := worktree.List()
	if err != nil {
		return err
	}

	if len(worktrees) == 0 {
		return fmt.Errorf("no worktrees found")
	}

	var targetBranch string

	if len(args) == 1 {
		targetBranch = args[0]
	} else {
		targetBranch, err = selectWorktree(worktrees)
		if err != nil {
			return err
		}
	}

	wt, err := worktree.FindByBranch(targetBranch)
	if err != nil {
		return err
	}

	if err := handleStashIfNeeded(); err != nil {
		return err
	}

	fmt.Printf("Switching to worktree '%s' at %s\n", wt.Branch, wt.Path)
	fmt.Println("Type 'exit' to return to the previous shell.")
	return openShellAt(wt.Path)
}

func selectWorktree(worktrees []worktree.Worktree) (string, error) {
	fmt.Println("Available worktrees:")
	fmt.Println()

	for i, wt := range worktrees {
		fmt.Printf("  [%d] %s (%s)\n", i+1, wt.Branch, wt.Path)
	}

	fmt.Println()
	fmt.Print("Select worktree (number): ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read input: %w", err)
	}

	input = strings.TrimSpace(input)
	num, err := strconv.Atoi(input)
	if err != nil || num < 1 || num > len(worktrees) {
		return "", fmt.Errorf("invalid selection: %s", input)
	}

	return worktrees[num-1].Branch, nil
}

func handleStashIfNeeded() error {
	hasChanges, err := worktree.HasUncommittedChanges()
	if err != nil {
		return err
	}

	if !hasChanges {
		return nil
	}

	fmt.Print("You have uncommitted changes. Create stash before switching? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input != "y" && input != "yes" {
		return nil
	}

	branch, err := worktree.GetCurrentBranch()
	if err != nil {
		return err
	}

	stashName := generateStashName(branch)
	fmt.Printf("Creating stash: %s\n", stashName)

	return worktree.CreateStash(stashName)
}

func generateStashName(branch string) string {
	now := time.Now()
	timestamp := now.Format("150405-02012006")
	return fmt.Sprintf("%s-%s", branch, timestamp)
}

func openShellAt(dir string) error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "bash"
	}

	cmd := exec.Command(shell)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
