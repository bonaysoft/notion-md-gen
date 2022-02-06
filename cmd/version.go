package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print the version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("release: %s, repo: %s, commit: %s", release, repo, commit)
	},
}

var (
	// RELEASE returns the release version
	release = "unknown"
	// REPO returns the git repository URL
	repo = "unknown"
	// COMMIT returns the short sha from git
	commit = "unknown"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}
