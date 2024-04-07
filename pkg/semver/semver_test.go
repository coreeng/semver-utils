package semver

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParse(t *testing.T) {
	assert := assert.New(t)

	goodVersions := []string{
		"1.0.0",
		"v1.0.0",
		"1.0.0-alpha",
		"1.0.0-alpha.1",
		"1.0.0-0.3.7",
		"1.0.0-x.7.z.92",
		"1.0.0+20130313144700",
		"1.0.0-beta.1+exp.sha.5114f85",
	}

	for _, v := range goodVersions {
		parsed, err := Parse(v)

		assert.Nil(err)
		assert.NotNil(parsed)
	}

	badVersions := []string{
		"1",                    // missing minor and patch
		"1.0",                  // missing patch
		"1.0.0-",               // empty prerelease
		"1.0.0-alpha.",         // empty identifier in prerelease (trailing dot)
		"1.0.0+",               // empty buildmetadata
		"1.0.0+meta.",          // empty identifier in buildmetadata (trailing dot)
		"1.0.0-+",              // empty prerelease and buildmetadata
		"1.0.0-alpha+",         // valid prerelease, empty buildmetadata
		"1.0.0-+meta",          // empty prerelease, valid buildmetadata
		"1.0.0-alpha.+1234",    // invalid prerelease (trailing dot), valid buildmetadata
		"1.0.0-alpha+meta.",    // valid prerelease, invalid buildmetadata (trailing period)
		"1.0.0-beta.01+meta.1", // leading zero in prerelease identifier
	}

	for _, v := range badVersions {
		_, err := Parse(v)

		assert.Error(err)
	}
}

// TestCompare tests the Compare function of the SemVer struct
func TestCompare(t *testing.T) {
	tests := []struct {
		lhs, rhs SemVer
		expected int
		comment  string
	}{
		// Major version comparison
		{SemVer{Major: 1, Minor: 0, Patch: 0}, SemVer{Major: 1, Minor: 0, Patch: 0}, 0, "Equal versions"},
		{SemVer{Major: 1, Minor: 0, Patch: 0}, SemVer{Major: 2, Minor: 0, Patch: 0}, -1, "lhs < rhs because Major is smaller"},
		{SemVer{Major: 2, Minor: 0, Patch: 0}, SemVer{Major: 1, Minor: 0, Patch: 0}, 1, "lhs > rhs because Major is larger"},

		// Minor version comparison
		{SemVer{Major: 1, Minor: 1, Patch: 0}, SemVer{Major: 1, Minor: 0, Patch: 0}, 1, "lhs > rhs because Minor is larger"},
		{SemVer{Major: 1, Minor: 0, Patch: 0}, SemVer{Major: 1, Minor: 1, Patch: 0}, -1, "lhs < rhs because Minor is smaller"},

		// Patch version comparison
		{SemVer{Major: 1, Minor: 0, Patch: 1}, SemVer{Major: 1, Minor: 0, Patch: 0}, 1, "lhs > rhs because Patch is larger"},
		{SemVer{Major: 1, Minor: 0, Patch: 0}, SemVer{Major: 1, Minor: 0, Patch: 1}, -1, "lhs < rhs because Patch is smaller"},

		// Pre-release version comparison
		{SemVer{Major: 1, Minor: 0, Patch: 0, PreRelease: "alpha"}, SemVer{Major: 1, Minor: 0, Patch: 0}, -1, "lhs < rhs because pre-release has lower precedence"},
		{SemVer{Major: 1, Minor: 0, Patch: 0}, SemVer{Major: 1, Minor: 0, Patch: 0, PreRelease: "alpha"}, 1, "lhs > rhs because rhs is a pre-release"},
		{SemVer{Major: 1, Minor: 0, Patch: 0, PreRelease: "alpha"}, SemVer{Major: 1, Minor: 0, Patch: 0, PreRelease: "beta"}, -1, "lhs < rhs because 'alpha' < 'beta' lexically"},
		{SemVer{Major: 1, Minor: 0, Patch: 0, PreRelease: "beta"}, SemVer{Major: 1, Minor: 0, Patch: 0, PreRelease: "alpha"}, 1, "lhs > rhs because 'beta' > 'alpha' lexically"},

		// Numeric pre-release comparison
		{SemVer{Major: 1, Minor: 0, Patch: 0, PreRelease: "1.0.0"}, SemVer{Major: 1, Minor: 0, Patch: 0, PreRelease: "1.0.1"}, -1, "lhs < rhs because '1.0.0' < '1.0.1' numerically"},
		{SemVer{Major: 1, Minor: 0, Patch: 0, PreRelease: "1.0.1"}, SemVer{Major: 1, Minor: 0, Patch: 0, PreRelease: "1.0.0"}, 1, "lhs > rhs because '1.0.1' > '1.0.0' numerically"},

		// Mixed pre-release comparison (numeric and alphanumeric)
		{SemVer{Major: 1, Minor: 0, Patch: 0, PreRelease: "1.beta"}, SemVer{Major: 1, Minor: 0, Patch: 0, PreRelease: "1.alpha"}, 1, "lhs > rhs because 'beta' > 'alpha' lexically"},
		{SemVer{Major: 1, Minor: 0, Patch: 0, PreRelease: "1.alpha"}, SemVer{Major: 1, Minor: 0, Patch: 0, PreRelease: "1.beta"}, -1, "lhs < rhs because 'alpha' < 'beta' lexically"},
		{SemVer{Major: 1, Minor: 0, Patch: 0, PreRelease: "1.0"}, SemVer{Major: 1, Minor: 0, Patch: 0, PreRelease: "1.alpha"}, -1, "lhs < rhs because '0' < 'beta' lexically"},

		// Pre-release with different lengths
		{SemVer{Major: 1, Minor: 0, Patch: 0, PreRelease: "1.alpha.1"}, SemVer{Major: 1, Minor: 0, Patch: 0, PreRelease: "1.alpha"}, 1, "lhs > rhs because longer pre-release with equal prefix"},
		{SemVer{Major: 1, Minor: 0, Patch: 0, PreRelease: "1.alpha"}, SemVer{Major: 1, Minor: 0, Patch: 0, PreRelease: "1.alpha.1"}, -1, "lhs < rhs because shorter pre-release with equal prefix"},
	}

	for _, test := range tests {
		result := test.lhs.Compare(test.rhs)
		if result != test.expected {
			t.Errorf("Compare(%v, %v) = %d; expected %d. %s", test.lhs, test.rhs, result, test.expected, test.comment)
		}
	}
}

// TestBumpMajor tests the BumpMajor function of the SemVer struct
func TestBumpMajor(t *testing.T) {
	tests := []struct {
		input    SemVer
		expected SemVer
		comment  string
	}{
		{SemVer{Major: 1, Minor: 0, Patch: 0}, SemVer{Major: 2, Minor: 0, Patch: 0}, "Bumping major version"},
		{SemVer{Major: 0, Minor: 1, Patch: 1}, SemVer{Major: 1, Minor: 0, Patch: 0}, "Bumping major version with non-zero minor and patch"},
	}

	for _, test := range tests {
		result := test.input.BumpMajor()
		if result != test.expected {
			t.Errorf("BumpMajor(%v) = %v; expected %v. %s", test.input, result, test.expected, test.comment)
		}
	}
}

// TestBumpMinor tests the BumpMinor function of the SemVer struct
func TestBumpMinor(t *testing.T) {
	tests := []struct {
		input    SemVer
		expected SemVer
		comment  string
	}{
		{SemVer{Major: 1, Minor: 0, Patch: 0}, SemVer{Major: 1, Minor: 1, Patch: 0}, "Bumping minor version"},
		{SemVer{Major: 1, Minor: 1, Patch: 1}, SemVer{Major: 1, Minor: 2, Patch: 0}, "Bumping minor version with non-zero patch"},
	}

	for _, test := range tests {
		result := test.input.BumpMinor()
		if result != test.expected {
			t.Errorf("BumpMinor(%v) = %v; expected %v. %s", test.input, result, test.expected, test.comment)
		}
	}
}

// TestBumpPatch tests the BumpPatch function of the SemVer struct
func TestBumpPatch(t *testing.T) {
	tests := []struct {
		input    SemVer
		expected SemVer
		comment  string
	}{
		{SemVer{Major: 1, Minor: 0, Patch: 0}, SemVer{Major: 1, Minor: 0, Patch: 1}, "Bumping patch version"},
		{SemVer{Major: 1, Minor: 1, Patch: 1}, SemVer{Major: 1, Minor: 1, Patch: 2}, "Bumping patch version with non-zero patch"},
	}

	for _, test := range tests {
		result := test.input.BumpPatch()
		if result != test.expected {
			t.Errorf("BumpPatch(%v) = %v; expected %v. %s", test.input, result, test.expected, test.comment)
		}
	}
}

// TestSetPreRelease tests the SetPreRelease function of the SemVer struct
func TestSetPreRelease(t *testing.T) {
	tests := []struct {
		input       SemVer
		preRelease  PreRelease
		expected    SemVer
		expectError bool
		comment     string
	}{
		{SemVer{Major: 1, Minor: 0, Patch: 0}, "alpha", SemVer{Major: 1, Minor: 0, Patch: 0, PreRelease: "alpha"}, false, "Valid pre-release version 'alpha'"},
		{SemVer{Major: 1, Minor: 0, Patch: 0}, "alpha.1", SemVer{Major: 1, Minor: 0, Patch: 0, PreRelease: "alpha.1"}, false, "Valid pre-release version 'alpha.1'"},
		{SemVer{Major: 1, Minor: 0, Patch: 0, PreRelease: "1"}, "", SemVer{Major: 1, Minor: 0, Patch: 0, PreRelease: ""}, false, "Valid pre-release version '' (empty)"},
		{SemVer{Major: 1, Minor: 0, Patch: 0}, "1.01", SemVer{}, true, "Invalid pre-release version '1.01'"},
		{SemVer{Major: 1, Minor: 0, Patch: 0}, "1.", SemVer{}, true, "Invalid pre-release version '1.'"},
		{SemVer{Major: 1, Minor: 0, Patch: 0}, "invalid@pre", SemVer{}, true, "Invalid pre-release version 'invalid@pre'"},
	}

	for _, test := range tests {
		result, err := test.input.SetPreRelease(test.preRelease)
		if (err != nil) != test.expectError {
			t.Errorf("SetPreRelease(%v, %s) error = %v, expectError %v. %s", test.input, test.preRelease, err, test.expectError, test.comment)
		}
		if result != test.expected {
			t.Errorf("SetPreRelease(%v, %s) = %v; expected %v. %s", test.input, test.preRelease, result, test.expected, test.comment)
		}
	}
}

// TestSetBuildMetadata tests the SetBuildMetadata function of the SemVer struct
func TestSetBuildMetadata(t *testing.T) {
	tests := []struct {
		input         SemVer
		buildMetadata BuildMetadata
		expected      SemVer
		expectError   bool
		comment       string
	}{
		{SemVer{Major: 1, Minor: 0, Patch: 0}, "001", SemVer{Major: 1, Minor: 0, Patch: 0, BuildMetadata: "001"}, false, "Valid build metadata '001'"},
		{SemVer{Major: 1, Minor: 0, Patch: 0}, "exp.sha.5114f85", SemVer{Major: 1, Minor: 0, Patch: 0, BuildMetadata: "exp.sha.5114f85"}, false, "Valid build metadata 'exp.sha.5114f85'"},
		{SemVer{Major: 1, Minor: 0, Patch: 0, BuildMetadata: "a"}, "", SemVer{Major: 1, Minor: 0, Patch: 0, BuildMetadata: ""}, false, "Valid build metadata '' (empty)"},
		{SemVer{Major: 1, Minor: 0, Patch: 0}, "invalid@meta", SemVer{}, true, "Invalid build metadata 'invalid@meta'"},
		{SemVer{Major: 1, Minor: 0, Patch: 0}, "a.", SemVer{}, true, "Invalid build metadata 'a.'"},
	}

	for _, test := range tests {
		result, err := test.input.SetBuildMetadata(test.buildMetadata)
		if (err != nil) != test.expectError {
			t.Errorf("SetBuildMetadata(%v, %s) error = %v, expectError %v. %s", test.input, test.buildMetadata, err, test.expectError, test.comment)
		}
		if result != test.expected {
			t.Errorf("SetBuildMetadata(%v, %s) = %v; expected %v. %s", test.input, test.buildMetadata, result, test.expected, test.comment)
		}
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		input    SemVer
		expected string
		comment  string
	}{
		{SemVer{Major: 1, Minor: 2, Patch: 3}, "1.2.3", "Basic version without pre-release or build metadata"},
		{SemVer{Major: 0, Minor: 0, Patch: 1}, "0.0.1", "Version with zeros"},
		{SemVer{Major: 1, Minor: 2, Patch: 3, PreRelease: "alpha"}, "1.2.3-alpha", "Version with pre-release"},
		{SemVer{Major: 1, Minor: 2, Patch: 3, BuildMetadata: "001"}, "1.2.3+001", "Version with build metadata"},
		{SemVer{Major: 1, Minor: 2, Patch: 3, PreRelease: "beta", BuildMetadata: "exp.sha.5114f85"}, "1.2.3-beta+exp.sha.5114f85", "Version with pre-release and build metadata"},
		{SemVer{Major: 1, Minor: 0, Patch: 0, PreRelease: "alpha.1"}, "1.0.0-alpha.1", "Version with multi-part pre-release"},
		{SemVer{Major: 1, Minor: 0, Patch: 0, BuildMetadata: "20130313144700"}, "1.0.0+20130313144700", "Version with numeric build metadata"},
		{SemVer{Major: 1, Minor: 0, Patch: 0, PreRelease: "", BuildMetadata: ""}, "1.0.0", "Version with empty pre-release and build metadata"},
	}

	for _, test := range tests {
		result := test.input.String()
		if result != test.expected {
			t.Errorf("String(%v) = %s; expected %s. %s", test.input, result, test.expected, test.comment)
		}
	}

	// Additional test for SetPreRelease interaction
	t.Run("SetPreRelease interaction", func(t *testing.T) {
		// 1. Create a SemVer with a valid non-empty prerelease field
		initialVersion := SemVer{Major: 2, Minor: 1, Patch: 0, PreRelease: "alpha.1"}

		// 2. Call String and verify
		initialResult := initialVersion.String()
		expectedInitial := "2.1.0-alpha.1"
		if initialResult != expectedInitial {
			t.Errorf("Initial String() = %s; expected %s", initialResult, expectedInitial)
		}

		// 3. Use SetPreRelease to set the prerelease field to empty
		updatedVersion, err := initialVersion.SetPreRelease("")
		if err != nil {
			t.Errorf("SetPreRelease(\"\") returned an unexpected error: %v", err)
		}

		// 4. Call String again and verify
		updatedResult := updatedVersion.String()
		expectedUpdated := "2.1.0"
		if updatedResult != expectedUpdated {
			t.Errorf("Updated String() = %s; expected %s", updatedResult, expectedUpdated)
		}
	})

	// Additional test for SetBuildMetadata interaction
	t.Run("SetBuildMetadata interaction", func(t *testing.T) {
		// 1. Create a SemVer with a valid non-empty buildmetadata field
		initialVersion := SemVer{Major: 1, Minor: 2, Patch: 3, BuildMetadata: "build.123"}

		// 2. Call String and verify
		initialResult := initialVersion.String()
		expectedInitial := "1.2.3+build.123"
		if initialResult != expectedInitial {
			t.Errorf("Initial String() = %s; expected %s", initialResult, expectedInitial)
		}

		// 3. Use SetBuildMetadata to set the buildmetadata field to empty
		updatedVersion, err := initialVersion.SetBuildMetadata("")
		if err != nil {
			t.Errorf("SetBuildMetadata(\"\") returned an unexpected error: %v", err)
		}

		// 4. Call String again and verify
		updatedResult := updatedVersion.String()
		expectedUpdated := "1.2.3"
		if updatedResult != expectedUpdated {
			t.Errorf("Updated String() = %s; expected %s", updatedResult, expectedUpdated)
		}
	})
}
