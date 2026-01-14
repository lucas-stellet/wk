package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/lucas-stellet/wk/internal/config"
	"github.com/lucas-stellet/wk/internal/hooks"
	"github.com/lucas-stellet/wk/internal/worktree"
)

var newCmd = &cobra.Command{
	Use:   "new <branch>",
	Short: "Create a new worktree",
	Long: `Create a new git worktree and run post-creation hooks.

This command:
  1. Creates a new worktree using git worktree add
  2. Copies files listed in .wk.yaml
  3. Runs post_hooks from .wk.yaml`,
	Args: cobra.ExactArgs(1),
	RunE: runNew,
}

func init() {
	rootCmd.AddCommand(newCmd)
}

func runNew(cmd *cobra.Command, args []string) error {
	branch := args[0]

	// Get current directory (source worktree)
	srcDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	// Create worktree
	fmt.Printf("Creating worktree for branch '%s'...\n", branch)
	dstDir, err := worktree.Add(branch)
	if err != nil {
		return err
	}
	fmt.Printf("Created worktree at %s\n", dstDir)

	// Load config
	configPath, err := config.FindConfig(srcDir)
	if os.IsNotExist(err) {
		fmt.Println("No .wk.yaml found, skipping hooks")
		return nil
	}
	if err != nil {
		return fmt.Errorf("find config: %w", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Copy files
	if len(cfg.Copy) > 0 {
		fmt.Println("\nCopying files...")
		if err := hooks.CopyFiles(srcDir, dstDir, cfg.Copy); err != nil {
			return fmt.Errorf("copy files: %w", err)
		}
	}

	// Run post hooks
	if len(cfg.PostHooks) > 0 {
		fmt.Println("\nRunning post hooks...")
		if err := hooks.RunPostHooks(dstDir, cfg.PostHooks); err != nil {
			return fmt.Errorf("run hooks: %w", err)
		}
	}

	fmt.Printf("\nWorktree '%s' is ready!\n", branch)
	return nil
}
