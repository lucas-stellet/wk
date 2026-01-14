package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/lucas-stellet/wk/internal/worktree"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all worktrees",
	RunE:    runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	worktrees, err := worktree.List()
	if err != nil {
		return err
	}

	if len(worktrees) == 0 {
		fmt.Println("No worktrees found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "BRANCH\tPATH\tCOMMIT")
	for _, wt := range worktrees {
		commit := wt.Commit
		if len(commit) > 7 {
			commit = commit[:7]
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", wt.Branch, wt.Path, commit)
	}
	return w.Flush()
}
