// Package validate provides pre-command validation for wk CLI.
package validate

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/lucas-stellet/wk/internal/config"
	"github.com/spf13/cobra"
)

// IsGitRepository reports whether the current directory is inside a git repository.
func IsGitRepository() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	err := cmd.Run()
	return err == nil
}

// CheckConfig validates the existence and format of .wk.yaml.
// Returns exists=true if file found, valid=true if YAML parses correctly.
func CheckConfig() (exists, valid bool, err error) {
	wd, err := os.Getwd()
	if err != nil {
		return false, false, fmt.Errorf("get working directory: %w", err)
	}

	configPath, err := config.FindConfig(wd)
	if os.IsNotExist(err) {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}

	_, err = config.Load(configPath)
	if err != nil {
		return true, false, err
	}

	return true, true, nil
}

// RunPreValidation performs validation checks before command execution.
// It skips validation for help, version, and update commands.
func RunPreValidation(cmd *cobra.Command) error {
	if isHelpCommand(cmd) {
		return nil
	}

	// Skip validation for commands that don't need git repo
	if shouldSkipValidation(cmd) {
		return nil
	}

	if !IsGitRepository() {
		return fmt.Errorf("not a git repository (or any parent up to mount point /)\n\nRun this command from inside a git repository")
	}

	if cmd.Name() == "init" {
		return nil
	}

	exists, valid, err := CheckConfig()

	if !exists {
		fmt.Fprintln(os.Stderr, "hint: no .wk.yaml found. Run 'wk init' to create one.\n")
		return nil
	}

	if !valid {
		return fmt.Errorf("invalid .wk.yaml: %w\n\nFix the YAML syntax in your configuration file", err)
	}

	return nil
}

// isHelpCommand reports whether cmd is a help command or has --help flag.
func isHelpCommand(cmd *cobra.Command) bool {
	if cmd.Name() == "help" {
		return true
	}

	helpFlag := cmd.Flags().Lookup("help")
	if helpFlag != nil && helpFlag.Changed {
		return true
	}

	return false
}

// shouldSkipValidation returns true for commands that don't need git repo validation.
func shouldSkipValidation(cmd *cobra.Command) bool {
	skipCommands := []string{"version", "update", "completion"}
	name := cmd.Name()
	for _, skip := range skipCommands {
		if name == skip {
			return true
		}
	}
	return false
}
