package main

import (
	"fmt"
	"github.com/coreeng/semver-utils/internal/build"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd is the base command for the CLI.
var rootCmd = &cobra.Command{
	Use:   "semver",
	Short: "A CLI for interacting with semantic versions",
}

// versionCmd prints version/build info.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the semver-utils version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("semver version: %s (commit: %s, built at: %s)\n",
			build.BuildVersion, build.BuildCommit, build.BuildDate)
	},
}

func init() {
	// Add the versionCmd to the root command.
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
	}
}
