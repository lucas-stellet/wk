// Package cmd contains CLI commands for wk.
package cmd

import (
	"os"

	"github.com/lucas-stellet/wk/internal/validate"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wk",
	Short: "Git worktree helper with hooks support",
	Long: `wk is a CLI tool that simplifies git worktree management.

It reads .wk.yaml from your project to automatically:
  - Copy files to new worktrees
  - Run post-creation hooks`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return validate.RunPreValidation(cmd)
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
