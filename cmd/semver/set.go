package main

import (
	"fmt"
	"github.com/coreeng/semver-utils/internal/cli"
	"os"
	"strconv"

	"github.com/coreeng/semver-utils/pkg/semver"
	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set parts of a semantic version",
	Long: `Set the major, minor, patch, prerelease, or buildmetadata fields of a semantic version,
producing a new version string.

Usage:
  semver set [subcommand] <version> <newValue>

Subcommands:
  major         Set the major field
  minor         Set the minor field
  patch         Set the patch field
  prerelease    Set the prerelease field
  buildmetadata Set the buildmetadata field

Examples:
  semver set major 1.2.3 10
  # Outputs: 10.2.3

  semver set prerelease 1.2.3-alpha rc.1
  # Outputs: 1.2.3-rc.1
`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var setMajorCmd = &cobra.Command{
	Use:   "major <version> <newMajor>",
	Short: "Set the major field",
	Long: `Set the major version field of a semantic version to <newMajor>.

Usage:
  semver set major <version> <newMajor>

Example:
  semver set major 1.2.3 10
  # Outputs: 10.2.3
`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		v := cli.ParseOrExit("set major", "version", args[0])
		newMajor, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "set major: invalid major '%s': %v\n", args[1], err)
			os.Exit(1)
		}
		updated := semver.SemVer{
			Major:         newMajor,
			Minor:         v.Minor,
			Patch:         v.Patch,
			PreRelease:    v.PreRelease,
			BuildMetadata: v.BuildMetadata,
		}
		fmt.Println(updated.String())
	},
}

var setMinorCmd = &cobra.Command{
	Use:   "minor <version> <newMinor>",
	Short: "Set the minor field",
	Long: `Set the minor version field of a semantic version to <newMinor>.

Usage:
  semver set minor <version> <newMinor>

Example:
  semver set minor 1.2.3 10
  # Outputs: 1.10.3
`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		v := cli.ParseOrExit("set minor", "version", args[0])
		newMinor, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "set minor: invalid minor '%s': %v\n", args[1], err)
			os.Exit(1)
		}
		updated := semver.SemVer{
			Major:         v.Major,
			Minor:         newMinor,
			Patch:         v.Patch,
			PreRelease:    v.PreRelease,
			BuildMetadata: v.BuildMetadata,
		}
		fmt.Println(updated.String())
	},
}

var setPatchCmd = &cobra.Command{
	Use:   "patch <version> <newPatch>",
	Short: "Set the patch field",
	Long: `Set the patch version field of a semantic version to <newPatch>.

Usage:
  semver set patch <version> <newPatch>

Example:
  semver set patch 1.2.3 99
  # Outputs: 1.2.99
`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		v := cli.ParseOrExit("set patch", "version", args[0])
		newPatch, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "set patch: invalid patch '%s': %v\n", args[1], err)
			os.Exit(1)
		}
		updated := semver.SemVer{
			Major:         v.Major,
			Minor:         v.Minor,
			Patch:         newPatch,
			PreRelease:    v.PreRelease,
			BuildMetadata: v.BuildMetadata,
		}
		fmt.Println(updated.String())
	},
}

var setPreCmd = &cobra.Command{
	Use:   "prerelease <version> <newPrerelease>",
	Short: "Set the prerelease field",
	Long: `Set the prerelease field of a semantic version to <newPrerelease>.

Usage:
  semver set prerelease <version> <newPrerelease>

Example:
  semver set prerelease 1.2.3-alpha rc.1
  # Outputs: 1.2.3-rc.1
`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		v := cli.ParseOrExit("set prerelease", "version", args[0])
		newPre := semver.PreRelease(args[1])
		updated, err := v.SetPreRelease(newPre)
		if err != nil {
			fmt.Fprintf(os.Stderr, "set prerelease: invalid prerelease '%s': %v\n", args[1], err)
			os.Exit(1)
		}
		fmt.Println(updated.String())
	},
}

var setBuildCmd = &cobra.Command{
	Use:   "buildmetadata <version> <newBuildmetadata>",
	Short: "Set the buildmetadata field",
	Long: `Set the buildmetadata field of a semantic version to <newBuildmetadata>.

Usage:
  semver set buildmetadata <version> <newBuildmetadata>

Example:
  semver set buildmetadata 1.2.3-alpha+abc build456
  # Outputs: 1.2.3-alpha+build456
`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		v := cli.ParseOrExit("set buildmetadata", "version", args[0])
		newBuild := semver.BuildMetadata(args[1])
		updated, err := v.SetBuildMetadata(newBuild)
		if err != nil {
			fmt.Fprintf(os.Stderr, "set buildmetadata: invalid buildmetadata '%s': %v\n", args[1], err)
			os.Exit(1)
		}
		fmt.Println(updated.String())
	},
}

func init() {
	rootCmd.AddCommand(setCmd)
	setCmd.AddCommand(setMajorCmd)
	setCmd.AddCommand(setMinorCmd)
	setCmd.AddCommand(setPatchCmd)
	setCmd.AddCommand(setPreCmd)
	setCmd.AddCommand(setBuildCmd)
}
