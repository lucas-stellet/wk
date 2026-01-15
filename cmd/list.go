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
	if err := w.Flush(); err != nil {
		return err
	}

	// Detect worktrees not in standard location
	var nonStandard []worktree.Worktree
	for _, wt := range worktrees {
		isStandard, err := worktree.IsInStandardLocation(wt.Path)
		if err != nil {
			continue
		}
		if !isStandard {
			nonStandard = append(nonStandard, wt)
		}
	}

	if len(nonStandard) > 0 {
		fmt.Println()
		fmt.Printf("Warning: %d worktree(s) not in standard location:\n", len(nonStandard))
		for _, wt := range nonStandard {
			fmt.Printf("  - %s (%s)\n", wt.Branch, wt.Path)
		}
		fmt.Println()
		fmt.Println("Run 'wk organize' to move them to the standard location.")
	}

	return nil
}
