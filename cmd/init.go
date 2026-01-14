package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"wk/internal/config"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a .wk.yaml configuration file",
	Long: `Interactively create a .wk.yaml configuration file.

This command guides you through setting up:
  - Files to copy to new worktrees
  - Post-creation hooks to run`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	// Check if config already exists
	if _, err := os.Stat(config.ConfigFileName); err == nil {
		fmt.Printf("%s already exists. Overwrite? [y/N]: ", config.ConfigFileName)
		if !confirmPrompt() {
			fmt.Println("Aborted")
			return nil
		}
	}

	reader := bufio.NewReader(os.Stdin)
	cfg := &config.Config{}

	fmt.Println("Creating .wk.yaml configuration")
	fmt.Println("Press Enter to skip any section\n")

	// Files to copy
	fmt.Println("Files/directories to copy to new worktrees")
	fmt.Println("(comma-separated, e.g.: .env,.env.local,tmp/)")
	fmt.Print("> ")
	copyInput, _ := reader.ReadString('\n')
	copyInput = strings.TrimSpace(copyInput)
	if copyInput != "" {
		cfg.Copy = parseCSV(copyInput)
	}

	fmt.Println()

	// Post hooks
	fmt.Println("Post-creation hooks (commands to run after creating worktree)")
	fmt.Println("Enter one command per line, empty line to finish:")
	for {
		fmt.Print("> ")
		hookInput, _ := reader.ReadString('\n')
		hookInput = strings.TrimSpace(hookInput)
		if hookInput == "" {
			break
		}
		cfg.PostHooks = append(cfg.PostHooks, hookInput)
	}

	// Generate YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	// Write file
	if err := os.WriteFile(config.ConfigFileName, data, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	fmt.Printf("\nCreated %s:\n", config.ConfigFileName)
	fmt.Println("---")
	fmt.Print(string(data))
	fmt.Println("---")

	return nil
}

func confirmPrompt() bool {
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}

func parseCSV(input string) []string {
	parts := strings.Split(input, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
