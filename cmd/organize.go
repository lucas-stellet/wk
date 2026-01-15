package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/lucas-stellet/wk/internal/worktree"
)

var organizeCmd = &cobra.Command{
	Use:   "organize",
	Short: "Move worktrees to the standard location",
	Long:  "Move worktrees that are not in the standard location (<repo>.worktrees/<branch>) to the correct path.",
	RunE:  runOrganize,
}

func init() {
	rootCmd.AddCommand(organizeCmd)
}

func runOrganize(cmd *cobra.Command, args []string) error {
	worktrees, err := worktree.List()
	if err != nil {
		return err
	}

	// Find worktrees not in standard location
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

	if len(nonStandard) == 0 {
		fmt.Println("All worktrees are already in the standard location.")
		return nil
	}

	// Show what will be moved
	worktreesDir, err := worktree.GetWorktreesDir()
	if err != nil {
		return err
	}

	fmt.Printf("The following %d worktree(s) will be moved to %s:\n\n", len(nonStandard), worktreesDir)
	for _, wt := range nonStandard {
		fmt.Printf("  %s\n", wt.Branch)
		fmt.Printf("    from: %s\n", wt.Path)
		fmt.Printf("    to:   %s/%s\n\n", worktreesDir, wt.Branch)
	}

	// Ask for confirmation
	fmt.Print("Proceed? [y/N] ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		fmt.Println("Aborted.")
		return nil
	}

	// Move each worktree
	fmt.Println()
	for _, wt := range nonStandard {
		fmt.Printf("Moving %s... ", wt.Branch)
		newPath, err := worktree.Move(wt)
		if err != nil {
			fmt.Printf("failed: %v\n", err)
			continue
		}
		fmt.Printf("done (%s)\n", newPath)
	}

	fmt.Println("\nAll worktrees have been organized.")
	return nil
}
