package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version information (set from main.go)
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Print version information including version number, commit hash, and build date.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("pyhub-docs version %s\n", Version)
		fmt.Printf("  Commit: %s\n", Commit)
		fmt.Printf("  Built:  %s\n", BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}