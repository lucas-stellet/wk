package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/lucas-stellet/wk/internal/config"
	"github.com/lucas-stellet/wk/internal/hooks"
	"github.com/lucas-stellet/wk/internal/worktree"
)

var setupQuiet bool

var setupCmd = &cobra.Command{
	Use:   "setup [path]",
	Short: "Run copy and post hooks on an existing worktree",
	Long: `Run the setup steps (file copy + post hooks) on an existing worktree.

If path is not specified, uses the current directory.

This is useful for worktrees created externally (e.g. by Claude Code)
that need the same setup that 'wk new' provides.

Use -q/--quiet to suppress wk messages (hook output still shown).`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSetup,
}

func init() {
	rootCmd.AddCommand(setupCmd)
	setupCmd.Flags().BoolVarP(&setupQuiet, "quiet", "q", false, "Suppress wk messages (hook output still shown)")
}

func runSetup(cmd *cobra.Command, args []string) error {
	// Determine destination directory
	var dstDir string
	if len(args) == 1 {
		abs, err := filepath.Abs(args[0])
		if err != nil {
			return fmt.Errorf("resolve path: %w", err)
		}
		dstDir = abs
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}
		dstDir = wd
	}

	// Get main worktree as source for config and file copy
	srcDir, err := worktree.GetMainWorktreePath()
	if err != nil {
		return fmt.Errorf("get main worktree: %w", err)
	}

	// Load config from main worktree
	configPath, err := config.FindConfig(srcDir)
	if os.IsNotExist(err) {
		// No config found — exit silently (graceful degradation)
		return nil
	}
	if err != nil {
		return fmt.Errorf("find config: %w", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Copy files (skip if src == dst to avoid copying onto itself)
	if srcDir != dstDir && len(cfg.Copy) > 0 {
		if !setupQuiet {
			fmt.Println("Copying files...")
		}
		if err := hooks.CopyFiles(srcDir, dstDir, cfg.Copy); err != nil {
			return fmt.Errorf("copy files: %w", err)
		}
	}

	// Run post hooks
	if len(cfg.PostHooks) > 0 {
		if !setupQuiet {
			fmt.Println("Running post hooks...")
		}
		if err := hooks.RunPostHooks(dstDir, cfg.PostHooks); err != nil {
			return fmt.Errorf("run hooks: %w", err)
		}
	}

	if !setupQuiet {
		fmt.Println("Setup complete!")
	}

	return nil
}
