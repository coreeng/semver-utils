package semver

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	FullPattern          = regexp.MustCompile(`^v?(?P<major>0|[1-9]\d*)\.(?P<minor>0|[1-9]\d*)\.(?P<patch>0|[1-9]\d*)(?:-(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)
	PrereleasePattern    = regexp.MustCompile(`^(?P<prerelease>(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*)?$`)
	BuildmetadataPattern = regexp.MustCompile(`^(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*)?$`)
)

type PreRelease string
type BuildMetadata string

// SemVer represents all components of a Semantic Version
type SemVer struct {
	Major         int           `json:"major"`
	Minor         int           `json:"minor"`
	Patch         int           `json:"patch"`
	PreRelease    PreRelease    `json:"preRelease"`
	BuildMetadata BuildMetadata `json:"buildMetadata"`
}

func Parse(v string) (SemVer, error) {
	if !FullPattern.MatchString(v) {
		return SemVer{}, fmt.Errorf("unable to parse '%s' as a Semantic Version, please see the formatting requirements at: https://semver.org/#semantic-versioning-specification-semver", v)
	}

	matches := FullPattern.FindStringSubmatch(v)

	ver := SemVer{Major: 0, Minor: 0, Patch: 0}

	// discard errors from strconv.Atoi as we've already validated them from the parsing regex
	for i, name := range FullPattern.SubexpNames() {
		switch name {
		case "major":
			ver.Major, _ = strconv.Atoi(matches[i])
		case "minor":
			ver.Minor, _ = strconv.Atoi(matches[i])
		case "patch":
			ver.Patch, _ = strconv.Atoi(matches[i])
		case "prerelease":
			ver.PreRelease = PreRelease(matches[i])
		case "buildmetadata":
			ver.BuildMetadata = BuildMetadata(matches[i])
		}
	}

	return ver, nil
}

func (v SemVer) Compare(rhs SemVer) int {
	// Compare major version
	if v.Major < rhs.Major {
		return -1
	} else if v.Major > rhs.Major {
		return 1
	}

	// Compare minor version
	if v.Minor < rhs.Minor {
		return -1
	} else if v.Minor > rhs.Minor {
		return 1
	}

	// Compare patch version
	if v.Patch < rhs.Patch {
		return -1
	} else if v.Patch > rhs.Patch {
		return 1
	}

	// Compare pre-release versions
	preReleaseComparison := comparePreRelease(v.PreRelease, rhs.PreRelease)
	if preReleaseComparison != 0 {
		return preReleaseComparison
	}

	// Build metadata does not affect precedence
	return 0
}

func (v SemVer) BumpMajor() SemVer {
	return SemVer{Major: v.Major + 1, Minor: 0, Patch: 0, PreRelease: "", BuildMetadata: ""}
}

func (v SemVer) BumpMinor() SemVer {
	return SemVer{Major: v.Major, Minor: v.Minor + 1, Patch: 0, PreRelease: "", BuildMetadata: ""}
}

func (v SemVer) BumpPatch() SemVer {
	return SemVer{Major: v.Major, Minor: v.Minor, Patch: v.Patch + 1, PreRelease: "", BuildMetadata: ""}
}

func (v SemVer) SetPreRelease(preRelease PreRelease) (SemVer, error) {
	// validate it matches the prerelease required pattern
	if !PrereleasePattern.MatchString(string(preRelease)) && preRelease != "" {
		return SemVer{}, fmt.Errorf("unable to set '%s' as a PreRelease, please see the formatting requirements at: https://semver.org/#semantic-versioning-specification-semver", preRelease)
	}

	return SemVer{Major: v.Major, Minor: v.Minor, Patch: v.Patch, PreRelease: preRelease, BuildMetadata: v.BuildMetadata}, nil
}

func (v SemVer) SetBuildMetadata(buildMetadata BuildMetadata) (SemVer, error) {
	// validate it matches the buildmetadata required pattern
	if !BuildmetadataPattern.MatchString(string(buildMetadata)) && buildMetadata != "" {
		return SemVer{}, fmt.Errorf("unable to set '%s' as a BuildMetadata, please see the formatting requirements at: https://semver.org/#semantic-versioning-specification-semver", buildMetadata)
	}

	return SemVer{Major: v.Major, Minor: v.Minor, Patch: v.Patch, PreRelease: v.PreRelease, BuildMetadata: buildMetadata}, nil
}

func (v SemVer) String() string {
	result := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.PreRelease != "" {
		result += "-" + string(v.PreRelease)
	}
	if v.BuildMetadata != "" {
		result += "+" + string(v.BuildMetadata)
	}
	return result
}

func comparePreRelease(lhs, rhs PreRelease) int {
	// If both are empty, they are equal
	if lhs == "" && rhs == "" {
		return 0
	}

	// If one is empty, the non-empty one has lower precedence
	if lhs == "" {
		return 1
	}
	if rhs == "" {
		return -1
	}

	// Split pre-releases into identifiers
	lhsIdentifiers := strings.Split(string(lhs), ".")
	rhsIdentifiers := strings.Split(string(rhs), ".")

	// Compare each identifier
	for i := 0; i < len(lhsIdentifiers) && i < len(rhsIdentifiers); i++ {
		lhsID := lhsIdentifiers[i]
		rhsID := rhsIdentifiers[i]

		// Check if identifiers are numeric
		lhsIsNumeric := isNumeric(lhsID)
		rhsIsNumeric := isNumeric(rhsID)

		// Numeric identifiers have lower precedence than non-numeric
		if lhsIsNumeric && !rhsIsNumeric {
			return -1
		}
		if !lhsIsNumeric && rhsIsNumeric {
			return 1
		}

		// Compare numeric identifiers numerically
		if lhsIsNumeric && rhsIsNumeric {
			lhsNum, _ := strconv.Atoi(lhsID)
			rhsNum, _ := strconv.Atoi(rhsID)
			if lhsNum < rhsNum {
				return -1
			}
			if lhsNum > rhsNum {
				return 1
			}
		} else {
			// Compare non-numeric identifiers lexically
			if lhsID < rhsID {
				return -1
			}
			if lhsID > rhsID {
				return 1
			}
		}
	}

	// If all compared identifiers are equal, the one with more identifiers has higher precedence
	if len(lhsIdentifiers) < len(rhsIdentifiers) {
		return -1
	}
	if len(lhsIdentifiers) > len(rhsIdentifiers) {
		return 1
	}

	return 0
}

func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
