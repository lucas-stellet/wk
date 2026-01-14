package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/lucas-stellet/wk/internal/updater"
	"github.com/spf13/cobra"
)

var (
	forceUpdate bool
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update wk to the latest version",
	Long: `Update wk to the latest version from GitHub releases.

This command checks for updates and offers to download and install
the latest version if one is available.`,
	RunE: runUpdate,
}

func init() {
	updateCmd.Flags().BoolVarP(&forceUpdate, "force", "f", false, "Skip confirmation prompt")
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	// Check install method first
	method := updater.DetectInstallMethod()

	switch method {
	case updater.InstallMethodHomebrew:
		fmt.Println("wk was installed via Homebrew.")
		fmt.Println("Run 'brew upgrade wk' to update.")
		return nil
	case updater.InstallMethodGo:
		fmt.Println("wk was installed via 'go install'.")
		fmt.Println("Run 'go install github.com/lucas-stellet/wk@latest' to update.")
		return nil
	}

	fmt.Println("Checking for updates...")

	info, err := updater.CheckForUpdate(version)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if !info.UpdateAvailable {
		fmt.Printf("wk is up to date (%s)\n", version)
		return nil
	}

	fmt.Printf("\nCurrent version: %s\n", info.CurrentVersion)
	fmt.Printf("Latest version:  %s\n", info.LatestVersion)
	fmt.Println()

	if info.DownloadURL == "" {
		fmt.Println("No binary available for your platform.")
		fmt.Printf("Visit %s to download manually.\n", info.ReleaseURL)
		return nil
	}

	if !forceUpdate {
		fmt.Print("Do you want to update? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			fmt.Println("Update cancelled.")
			return nil
		}
	}

	fmt.Printf("\nDownloading wk %s...\n", info.LatestVersion)

	if err := updater.PerformUpdate(info); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	fmt.Printf("\nSuccessfully updated to %s\n", info.LatestVersion)
	return nil
}
