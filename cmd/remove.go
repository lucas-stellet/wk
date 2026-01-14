package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"wk/internal/worktree"
)

var removeCmd = &cobra.Command{
	Use:     "remove <branch>",
	Aliases: []string{"rm"},
	Short:   "Remove a worktree",
	Args:    cobra.ExactArgs(1),
	RunE:    runRemove,
}

func init() {
	rootCmd.AddCommand(removeCmd)
}

func runRemove(cmd *cobra.Command, args []string) error {
	target := args[0]

	fmt.Printf("Removing worktree '%s'...\n", target)
	if err := worktree.Remove(target); err != nil {
		return err
	}

	fmt.Printf("Worktree '%s' removed\n", target)
	return nil
}
