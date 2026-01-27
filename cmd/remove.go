package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/lucas-stellet/wk/internal/selector"
	"github.com/lucas-stellet/wk/internal/worktree"
)

var removeForce bool

var removeCmd = &cobra.Command{
	Use:     "remove [branch]",
	Aliases: []string{"rm"},
	Short:   "Remove a worktree",
	Long: `Remove a git worktree by branch name.

If branch is not specified, opens an interactive selector to choose which
worktree to remove.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runRemove,
}

func init() {
	rootCmd.AddCommand(removeCmd)
	removeCmd.Flags().BoolVarP(&removeForce, "force", "f", false, "Force removal even if worktree has uncommitted changes")
}

func runRemove(cmd *cobra.Command, args []string) error {
	var target string

	if len(args) == 1 {
		target = args[0]
	} else {
		selected, err := selector.SelectWorktree()
		if err != nil {
			if errors.Is(err, selector.ErrCancelled) {
				return nil
			}
			return err
		}
		target = selected
	}

	fmt.Printf("Removing worktree '%s'...\n", target)
	if err := worktree.Remove(target, removeForce); err != nil {
		return err
	}

	fmt.Printf("Worktree '%s' removed\n", target)
	return nil
}
