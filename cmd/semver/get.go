package main

import (
	"fmt"
	"github.com/coreeng/semver-utils/internal/cli"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieve parts of a semantic version",
	Long: `Extract a specific field from a semantic version.

Usage:
  semver get [subcommand] <version>

Subcommands:
  major          Get the major field
  minor          Get the minor field
  patch          Get the patch field
  prerelease     Get the prerelease field
  buildmetadata  Get the buildmetadata field

Examples:
  semver get major 1.2.3-alpha+build123
  # Outputs: 1
`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var getMajorCmd = &cobra.Command{
	Use:   "major <version>",
	Short: "Get the major field",
	Long: `Retrieve the major field from a semantic version.

Usage:
  semver get major <version>
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		v := cli.ParseOrExit("get major", "version", args[0])
		fmt.Println(v.Major)
	},
}

var getMinorCmd = &cobra.Command{
	Use:   "minor <version>",
	Short: "Get the minor field",
	Long: `Retrieve the minor field from a semantic version.

Usage:
  semver get minor <version>
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		v := cli.ParseOrExit("get minor", "version", args[0])
		fmt.Println(v.Minor)
	},
}

var getPatchCmd = &cobra.Command{
	Use:   "patch <version>",
	Short: "Get the patch field",
	Long: `Retrieve the patch field from a semantic version.

Usage:
  semver get patch <version>
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		v := cli.ParseOrExit("get patch", "version", args[0])
		fmt.Println(v.Patch)
	},
}

var getPreCmd = &cobra.Command{
	Use:   "prerelease <version>",
	Short: "Get the prerelease field",
	Long: `Retrieve the prerelease field from a semantic version.

Usage:
  semver get prerelease <version>
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		v := cli.ParseOrExit("get prerelease", "version", args[0])
		fmt.Println(v.PreRelease)
	},
}

var getBuildCmd = &cobra.Command{
	Use:   "buildmetadata <version>",
	Short: "Get the buildmetadata field",
	Long: `Retrieve the buildmetadata field from a semantic version.

Usage:
  semver get buildmetadata <version>
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		v := cli.ParseOrExit("get buildmetadata", "version", args[0])
		fmt.Println(v.BuildMetadata)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(getMajorCmd)
	getCmd.AddCommand(getMinorCmd)
	getCmd.AddCommand(getPatchCmd)
	getCmd.AddCommand(getPreCmd)
	getCmd.AddCommand(getBuildCmd)
}
