// Package cmd contains CLI commands for wk.
package cmd

import (
	"fmt"
	"os"

	"github.com/lucas-stellet/wk/internal/updater"
	"github.com/lucas-stellet/wk/internal/validate"
	"github.com/spf13/cobra"
)

var version = "dev"

// SetVersion sets the version string from main.
func SetVersion(v string) {
	version = v
}

var rootCmd = &cobra.Command{
	Use:   "wk",
	Short: "Git worktree helper with hooks support",
	Long: `wk is a CLI tool that simplifies git worktree management.

It reads .wk.yaml from your project to automatically:
  - Copy files to new worktrees
  - Run post-creation hooks`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := validate.RunPreValidation(cmd); err != nil {
			return err
		}

		// Check for updates (skip for certain commands)
		if shouldCheckUpdate(cmd) {
			checkAndNotifyUpdate()
		}

		return nil
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// shouldCheckUpdate returns true if we should check for updates for this command.
func shouldCheckUpdate(cmd *cobra.Command) bool {
	name := cmd.Name()
	// Skip update check for these commands
	skipCommands := []string{"help", "version", "update", "completion"}
	for _, skip := range skipCommands {
		if name == skip {
			return false
		}
	}
	return true
}

// checkAndNotifyUpdate checks for updates using cache and notifies if available.
func checkAndNotifyUpdate() {
	// Run in background to not slow down command execution
	info, err := updater.CachedCheck(version)
	if err != nil {
		// Silently ignore errors - don't interrupt user workflow
		return
	}

	if info.UpdateAvailable {
		fmt.Fprintf(os.Stderr, "hint: A new version of wk is available (%s). Run 'wk update' to upgrade.\n\n", info.LatestVersion)
	}
}
