package cli

import (
	"fmt"
	"os"

	"github.com/coreeng/semver-utils/pkg/semver"
)

// ParseOrExit attempts to parse a single version string into a SemVer.
// If parsing fails, it prints an error and exits with code 2.
func ParseOrExit(cmdName, argName, versionStr string) semver.SemVer {
	v, err := semver.Parse(versionStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s command: error parsing %s '%s': %v\n",
			cmdName, argName, versionStr, err)
		os.Exit(2)
	}
	return v
}

// CompareOrExit parses two version strings, applies a comparison function,
// and exits with 0 if true, 1 if false, or 2 on error.
func CompareOrExit(cmdName string, args []string, cmp func(v1, v2 semver.SemVer) bool) {
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "%s command: requires exactly 2 arguments <version1> <version2>\n", cmdName)
		os.Exit(2)
	}
	v1 := ParseOrExit(cmdName, "version1", args[0])
	v2 := ParseOrExit(cmdName, "version2", args[1])

	if cmp(v1, v2) {
		os.Exit(0)
	}
	os.Exit(1)
}
