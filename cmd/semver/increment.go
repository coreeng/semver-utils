package main

import (
	"fmt"
	"github.com/coreeng/semver-utils/internal/cli"

	"github.com/spf13/cobra"
)

var incrementCmd = &cobra.Command{
	Use:   "increment",
	Short: "Increment a semantic version",
	Long: `Increment the major, minor, or patch field of a semantic version.
Usage:
  semver increment [subcommand] <version>`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var incrementMajorCmd = &cobra.Command{
	Use:   "major <version>",
	Short: "Increment the major version (minor/patch reset to 0)",
	Run: func(cmd *cobra.Command, args []string) {
		v := cli.ParseOrExit("increment major", "version", args[0])
		bumped := v.BumpMajor()
		fmt.Println(bumped.String())
	},
}

var incrementMinorCmd = &cobra.Command{
	Use:   "minor <version>",
	Short: "Increment the minor version (patch reset to 0)",
	Run: func(cmd *cobra.Command, args []string) {
		v := cli.ParseOrExit("increment minor", "version", args[0])
		bumped := v.BumpMinor()
		fmt.Println(bumped.String())
	},
}

var incrementPatchCmd = &cobra.Command{
	Use:   "patch <version>",
	Short: "Increment the patch version",
	Run: func(cmd *cobra.Command, args []string) {
		v := cli.ParseOrExit("increment patch", "version", args[0])
		bumped := v.BumpPatch()
		fmt.Println(bumped.String())
	},
}

// init registers the "increment" command and subcommands with the rootCmd.
func init() {
	rootCmd.AddCommand(incrementCmd)
	incrementCmd.AddCommand(incrementMajorCmd)
	incrementCmd.AddCommand(incrementMinorCmd)
	incrementCmd.AddCommand(incrementPatchCmd)
}
