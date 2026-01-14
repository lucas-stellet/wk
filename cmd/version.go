package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show wk version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("wk version %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
