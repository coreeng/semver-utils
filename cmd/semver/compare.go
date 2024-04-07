package main

import (
	"github.com/coreeng/semver-utils/internal/cli"
	"github.com/coreeng/semver-utils/pkg/semver"
	"github.com/spf13/cobra"
)

var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Compare two semantic versions",
	Long: `Compare two semantic versions using one of the following subcommands.

Usage:
  semver compare [subcommand] <version1> <version2>

Subcommands:
  gt  <v1> <v2>   Return 0 if v1 > v2, 1 if false
  gte <v1> <v2>   Return 0 if v1 >= v2, 1 if false
  eq  <v1> <v2>   Return 0 if v1 == v2, 1 if false
  lt  <v1> <v2>   Return 0 if v1 < v2, 1 if false
  lte <v1> <v2>   Return 0 if v1 <= v2, 1 if false

Exit codes:
  0  if the comparison is true
  1  if the comparison is false
  2  if an error occurs (e.g., invalid version string or argument count)

Examples:
  semver compare gt 1.2.3 1.2.0
  semver compare eq 1.2.3 1.2.3
`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var compareGtCmd = &cobra.Command{
	Use:   "gt <version1> <version2>",
	Short: "Check if version1 > version2",
	Long: `Check if version1 is greater than version2.

Usage:
  semver compare gt <version1> <version2>
`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cli.CompareOrExit("compare gt", args, func(v1, v2 semver.SemVer) bool {
			return v1.Compare(v2) > 0
		})
	},
}

var compareGteCmd = &cobra.Command{
	Use:   "gte <version1> <version2>",
	Short: "Check if version1 >= version2",
	Long: `Check if version1 is greater than or equal to version2.

Usage:
  semver compare gte <version1> <version2>
`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cli.CompareOrExit("compare gte", args, func(v1, v2 semver.SemVer) bool {
			return v1.Compare(v2) >= 0
		})
	},
}

var compareEqCmd = &cobra.Command{
	Use:   "eq <version1> <version2>",
	Short: "Check if version1 == version2",
	Long: `Check if version1 is exactly equal to version2.

Usage:
  semver compare eq <version1> <version2>
`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cli.CompareOrExit("compare eq", args, func(v1, v2 semver.SemVer) bool {
			return v1.Compare(v2) == 0
		})
	},
}

var compareLtCmd = &cobra.Command{
	Use:   "lt <version1> <version2>",
	Short: "Check if version1 < version2",
	Long: `Check if version1 is less than version2.

Usage:
  semver compare lt <version1> <version2>
`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cli.CompareOrExit("compare lt", args, func(v1, v2 semver.SemVer) bool {
			return v1.Compare(v2) < 0
		})
	},
}

var compareLteCmd = &cobra.Command{
	Use:   "lte <version1> <version2>",
	Short: "Check if version1 <= version2",
	Long: `Check if version1 is less than or equal to version2.

Usage:
  semver compare lte <version1> <version2>
`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cli.CompareOrExit("compare lte", args, func(v1, v2 semver.SemVer) bool {
			return v1.Compare(v2) <= 0
		})
	},
}

func init() {
	rootCmd.AddCommand(compareCmd)
	compareCmd.AddCommand(compareGtCmd)
	compareCmd.AddCommand(compareGteCmd)
	compareCmd.AddCommand(compareEqCmd)
	compareCmd.AddCommand(compareLtCmd)
	compareCmd.AddCommand(compareLteCmd)
}
