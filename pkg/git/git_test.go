package git

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"io"

	"github.com/coreeng/semver-utils/pkg/semver"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// commitSpec defines a commit's content, timestamp, associated tags, and branches.
type commitSpec struct {
	Message   string
	Timestamp int64
	Tags      []string
	Branches  []string
}

// setupRepo creates an in-memory Git repository with a linear commit history.
func setupRepo() (*git.Repository, []*object.Commit, error) {
	repo, err := git.Init(memory.NewStorage(), memfs.New())
	if err != nil {
		return nil, nil, err
	}
	w, err := repo.Worktree()
	if err != nil {
		return nil, nil, err
	}

	commitData := []commitSpec{
		{
			Message:   "First commit",
			Timestamp: 100,
			Tags:      []string{"first"},
			Branches:  []string{},
		},
		{
			Message:   "Second commit",
			Timestamp: 200,
			Tags:      nil,
			Branches:  []string{"test-branch"},
		},
		{
			Message:   "Third commit",
			Timestamp: 300,
			Tags:      []string{"v1.0.0", "annotated-tag"},
			Branches:  nil,
		},
		{
			Message:   "Fourth commit",
			Timestamp: 400,
			Tags:      []string{"release/v1.0.0"},
			Branches:  []string{},
		},
		{
			Message:   "Fifth commit",
			Timestamp: 500,
			Tags:      []string{"v1.4.0-alpha.1", "prefixed/v1.0.0", "non-semver-tag", "pre/non-semver"},
			Branches:  nil,
		},
	}

	commits := make([]*object.Commit, len(commitData))
	for i, spec := range commitData {
		file, err := w.Filesystem.Create("file.txt")
		if err != nil {
			return nil, nil, err
		}
		if _, err := file.Write([]byte(spec.Message)); err != nil {
			return nil, nil, err
		}
		_ = file.Close()

		if _, err := w.Add("file.txt"); err != nil {
			return nil, nil, err
		}

		commitHash, err := w.Commit(spec.Message, &git.CommitOptions{
			Author: &object.Signature{
				Name:  "test",
				Email: "test@example.com",
				When:  time.Unix(spec.Timestamp, 0),
			},
		})
		if err != nil {
			return nil, nil, err
		}

		commitObj, err := repo.CommitObject(commitHash)
		if err != nil {
			return nil, nil, err
		}
		commits[i] = commitObj

		for _, tag := range spec.Tags {
			if tag == "annotated-tag" {
				_, err = repo.CreateTag(tag, commitHash, &git.CreateTagOptions{
					Tagger: &object.Signature{
						Name:  "test",
						Email: "test@example.com",
						When:  time.Now(),
					},
					Message: "This is an annotated tag",
				})
			} else {
				_, err = repo.CreateTag(tag, commitHash, nil)
			}
			if err != nil {
				return nil, nil, err
			}
		}

		for _, branchName := range spec.Branches {
			branchRef := plumbing.NewBranchReferenceName(branchName)
			ref := plumbing.NewHashReference(branchRef, commitHash)
			if err := repo.Storer.SetReference(ref); err != nil {
				return nil, nil, err
			}
		}
	}

	// Print commits for debugging purposes.
	logIter, err := repo.Log(&git.LogOptions{
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, nil, err
	}

	for {
		c, err := logIter.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, err
		}
		parentCount := c.NumParents()
		fmt.Printf("Commit: %s | Parents: %d | Message: %s\n",
			c.Hash.String()[:7], parentCount, c.Message)
	}
	return repo, commits, nil
}

func TestFetchCommitObject(t *testing.T) {
	repo, commits, err := setupRepo()
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	assert.Len(t, commits, 5)

	// HEAD should be on the "Fifth commit" (index 4)
	headRef, err := repo.Head()
	assert.NoError(t, err)
	assert.Equal(t, commits[4].Hash, headRef.Hash())

	t.Run("Fetch HEAD", func(t *testing.T) {
		commit, err := FetchCommitObject(repo, "HEAD")
		assert.NoError(t, err)
		assert.NotNil(t, commit)
		assert.Equal(t, commits[4].Hash, commit.Hash)
	})

	t.Run("Fetch commit by hash", func(t *testing.T) {
		headHash := commits[4].Hash.String()
		commit, err := FetchCommitObject(repo, headHash)
		assert.NoError(t, err)
		assert.NotNil(t, commit)
		assert.Equal(t, commits[4].Hash, commit.Hash)
	})

	t.Run("Fetch commit by branch name", func(t *testing.T) {
		commit, err := FetchCommitObject(repo, "test-branch")
		assert.NoError(t, err)
		assert.NotNil(t, commit)
		assert.Equal(t, commits[1].Hash, commit.Hash)
	})

	t.Run("Fetch commit by lightweight tag (v1.0.0)", func(t *testing.T) {
		commit, err := FetchCommitObject(repo, "v1.0.0")
		assert.NoError(t, err)
		assert.NotNil(t, commit)
		assert.Equal(t, commits[2].Hash, commit.Hash)
	})

	t.Run("Fetch commit by annotated tag", func(t *testing.T) {
		commit, err := FetchCommitObject(repo, "annotated-tag")
		assert.NoError(t, err)
		assert.NotNil(t, commit)
		assert.Equal(t, commits[2].Hash, commit.Hash)
	})

	t.Run("Error fetching non-existent commit", func(t *testing.T) {
		commit, err := FetchCommitObject(repo, "non-existent-ref")
		assert.Error(t, err)
		assert.Nil(t, commit)
	})
}

func TestFetchVersionTagExact(t *testing.T) {
	repo, _, err := setupRepo()
	assert.NoError(t, err)

	head, err := repo.Head()
	assert.NoError(t, err)

	headCommit, err := repo.CommitObject(head.Hash())
	assert.NoError(t, err)

	t.Run("Find tag without prefix", func(t *testing.T) {
		tag, ver, commitObj, err := FetchVersionTag(repo, headCommit, "", true)
		assert.NoError(t, err)
		assert.Equal(t, "v1.4.0-alpha.1", tag)
		expectedVer, err := semver.Parse("1.4.0-alpha.1")
		assert.NoError(t, err)
		assert.True(t, ver.Compare(expectedVer) == 0)
		assert.Equal(t, headCommit.Hash, commitObj.Hash)
	})

	t.Run("Find tag with prefix", func(t *testing.T) {
		tag, ver, commitObj, err := FetchVersionTag(repo, headCommit, "prefixed", true)
		assert.NoError(t, err)
		assert.Equal(t, "prefixed/v1.0.0", tag)
		expectedVer, err := semver.Parse("1.0.0")
		assert.NoError(t, err)
		assert.True(t, ver.Compare(expectedVer) == 0)
		assert.Equal(t, headCommit.Hash, commitObj.Hash)
	})

	t.Run("No matching tag", func(t *testing.T) {
		tag, _, _, err := FetchVersionTag(repo, headCommit, "non-existent-prefix", true)
		assert.NoError(t, err)
		assert.Empty(t, tag)
	})

	t.Run("No matching tag with prefix", func(t *testing.T) {
		tag, _, _, err := FetchVersionTag(repo, headCommit, "some", true)
		assert.NoError(t, err)
		assert.Empty(t, tag)
	})

	t.Run("Tag on different commit", func(t *testing.T) {
		commitsIter, err := repo.Log(&git.LogOptions{})
		assert.NoError(t, err)
		_, err = commitsIter.Next() // Skip HEAD
		assert.NoError(t, err)
		secondCommit, err := commitsIter.Next()
		assert.NoError(t, err)

		tag, _, _, err := FetchVersionTag(repo, secondCommit, "", true)
		assert.NoError(t, err)
		assert.Empty(t, tag)
	})
}

func TestFetchVersionTag(t *testing.T) {
	repo, commits, err := setupRepo()
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	assert.Len(t, commits, 5)

	t.Run("Find previous version tag without prefix", func(t *testing.T) {
		tag, ver, commitObj, err := FetchVersionTag(repo, commits[4], "", false)
		assert.NoError(t, err)
		assert.Equal(t, "v1.4.0-alpha.1", tag)

		expectedVer, err := semver.Parse("1.4.0-alpha.1")
		assert.NoError(t, err)
		assert.True(t, ver.Compare(expectedVer) == 0)
		assert.Equal(t, commits[4].Hash, commitObj.Hash)
	})

	t.Run("Find previous version tag with prefix", func(t *testing.T) {
		tag, ver, commitObj, err := FetchVersionTag(repo, commits[4], "release", false)
		assert.NoError(t, err)
		assert.Equal(t, "release/v1.0.0", tag)
		expectedVer, err := semver.Parse("1.0.0")
		assert.NoError(t, err)
		assert.True(t, ver.Compare(expectedVer) == 0)
		assert.Equal(t, commits[3].Hash, commitObj.Hash)
	})

	t.Run("No previous version tag", func(t *testing.T) {
		tag, ver, commitObj, err := FetchVersionTag(repo, commits[0], "", false)
		assert.NoError(t, err)
		assert.Empty(t, tag)
		assert.Equal(t, semver.SemVer{}, ver)
		assert.Nil(t, commitObj)
	})

	t.Run("Ignore non-semver tags", func(t *testing.T) {
		tag, _, commitObj, err := FetchVersionTag(repo, commits[4], "", false)
		assert.NoError(t, err)
		assert.NotEqual(t, "non-semver-tag", tag)
		assert.NotEqual(t, "pre/non-semver", tag)
		assert.Equal(t, "v1.4.0-alpha.1", tag)
		assert.Equal(t, commits[4].Hash, commitObj.Hash)
	})

	t.Run("Tag on second commit, no previous version", func(t *testing.T) {
		tag, ver, commitObj, err := FetchVersionTag(repo, commits[1], "", false)
		assert.NoError(t, err)
		assert.Empty(t, tag)
		assert.Equal(t, semver.SemVer{}, ver)
		assert.Nil(t, commitObj)
	})
}

func TestCreateIncrementedVersionTag(t *testing.T) {
	tests := []struct {
		name            string
		prefix          string
		incrementType   string // "major", "minor", or "patch"
		annotated       bool
		prerelease      string
		buildMetadata   string
		expectedVersion string // expected semantic version string after bump/modification
		expectedTagName string // expected full tag name (with prefix, if any)
		expectErr       bool
		errContains     string
	}{
		// --- Cases without a prefix (using commit[4]'s tag "v1.4.0-alpha.1" as previous version) ---
		{
			name:            "No prefix, major, lightweight, no prerelease, no buildmetadata",
			prefix:          "",
			incrementType:   "major",
			annotated:       false,
			prerelease:      "",
			buildMetadata:   "",
			expectedVersion: "2.0.0",
			expectedTagName: "v2.0.0",
		},
		{
			name:            "No prefix, minor, annotated, no prerelease, no buildmetadata",
			prefix:          "",
			incrementType:   "minor",
			annotated:       true,
			prerelease:      "",
			buildMetadata:   "",
			expectedVersion: "1.5.0",
			expectedTagName: "v1.5.0",
		},
		{
			name:            "No prefix, patch, lightweight, with prerelease",
			prefix:          "",
			incrementType:   "patch",
			annotated:       false,
			prerelease:      "beta.2",
			buildMetadata:   "",
			expectedVersion: "1.4.1-beta.2",
			expectedTagName: "v1.4.1-beta.2",
		},
		{
			name:            "No prefix, major, annotated, with buildmetadata",
			prefix:          "",
			incrementType:   "major",
			annotated:       true,
			prerelease:      "",
			buildMetadata:   "build.123",
			expectedVersion: "2.0.0+build.123",
			expectedTagName: "v2.0.0+build.123",
		},
		{
			name:            "No prefix, minor, annotated, with prerelease and buildmetadata",
			prefix:          "",
			incrementType:   "minor",
			annotated:       true,
			prerelease:      "rc.1",
			buildMetadata:   "meta.456",
			expectedVersion: "1.5.0-rc.1+meta.456",
			expectedTagName: "v1.5.0-rc.1+meta.456",
		},
		// --- Cases with a prefix (using commit[3]'s tag "release/v1.0.0" as previous version) ---
		{
			name:            "With prefix, major, lightweight, no prerelease, no buildmetadata",
			prefix:          "release",
			incrementType:   "major",
			annotated:       false,
			prerelease:      "",
			buildMetadata:   "",
			expectedVersion: "2.0.0",
			expectedTagName: "release/v2.0.0",
		},
		{
			name:            "With prefix, minor, annotated, no prerelease, no buildmetadata",
			prefix:          "release",
			incrementType:   "minor",
			annotated:       true,
			prerelease:      "",
			buildMetadata:   "",
			expectedVersion: "1.1.0",
			expectedTagName: "release/v1.1.0",
		},
		{
			name:            "With prefix, patch, lightweight, with prerelease",
			prefix:          "release",
			incrementType:   "patch",
			annotated:       false,
			prerelease:      "beta.2",
			buildMetadata:   "",
			expectedVersion: "1.0.1-beta.2",
			expectedTagName: "release/v1.0.1-beta.2",
		},
		{
			name:            "With prefix, major, annotated, with buildmetadata",
			prefix:          "release",
			incrementType:   "major",
			annotated:       true,
			prerelease:      "",
			buildMetadata:   "build.123",
			expectedVersion: "2.0.0+build.123",
			expectedTagName: "release/v2.0.0+build.123",
		},
		{
			name:            "With prefix, minor, annotated, with prerelease and buildmetadata",
			prefix:          "release",
			incrementType:   "minor",
			annotated:       true,
			prerelease:      "rc.1",
			buildMetadata:   "meta.456",
			expectedVersion: "1.1.0-rc.1+meta.456",
			expectedTagName: "release/v1.1.0-rc.1+meta.456",
		},
		// --- Error cases ---
		{
			name:          "Invalid increment type",
			prefix:        "",
			incrementType: "invalid",
			annotated:     false,
			prerelease:    "",
			buildMetadata: "",
			expectErr:     true,
			errContains:   "unknown increment type: invalid",
		},
		{
			name:          "No previous version found",
			prefix:        "nonexistent",
			incrementType: "patch",
			annotated:     false,
			prerelease:    "",
			buildMetadata: "",
			expectErr:     true,
			errContains:   "no valid previous version tag found",
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			repo, commits, err := setupRepo()
			require.NoError(t, err)
			require.Len(t, commits, 5)
			targetCommit := commits[4]

			// Use FetchVersionTag instead of the removed FindPreviousVersionTag.
			prevTag, prevVersion, _, err := FetchVersionTag(repo, targetCommit, tc.prefix, false)
			require.NoError(t, err)
			if prevTag == "" {
				if tc.expectErr && strings.Contains(tc.errContains, "no valid previous version tag found") {
					// Expected error: no valid previous version tag found.
					return
				}
				t.Fatalf("no valid previous version tag found")
			}

			// Bump the version based on the incrementType.
			var newVersion semver.SemVer
			var bumpErr error
			switch strings.ToLower(tc.incrementType) {
			case "major":
				newVersion = prevVersion.BumpMajor()
			case "minor":
				newVersion = prevVersion.BumpMinor()
			case "patch":
				newVersion = prevVersion.BumpPatch()
			default:
				bumpErr = fmt.Errorf("unknown increment type: %s", tc.incrementType)
			}
			if bumpErr != nil {
				if tc.expectErr {
					require.Error(t, bumpErr)
					require.Contains(t, bumpErr.Error(), tc.errContains)
					return
				}
				require.NoError(t, bumpErr)
			}

			// Apply prerelease and build metadata modifications if provided.
			if tc.prerelease != "" {
				newVersion, err = newVersion.SetPreRelease(semver.PreRelease(tc.prerelease))
				require.NoError(t, err)
			}
			if tc.buildMetadata != "" {
				newVersion, err = newVersion.SetBuildMetadata(semver.BuildMetadata(tc.buildMetadata))
				require.NoError(t, err)
			}

			newTagName, err := CreateVersionTag(repo, targetCommit, newVersion, tc.prefix, tc.annotated)
			if tc.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expectedTagName, newTagName)
			require.Equal(t, tc.expectedVersion, newVersion.String())

			// Verify that the new tag exists in the repository.
			ref, err := repo.Reference(plumbing.NewTagReferenceName(newTagName), true)
			require.NoError(t, err)
			if tc.annotated {
				tagObj, err := repo.TagObject(ref.Hash())
				require.NoError(t, err)
				actualMsg := strings.TrimSpace(tagObj.Message)
				expectedMsg := fmt.Sprintf("Version %s", newVersion.String())
				require.Equal(t, expectedMsg, actualMsg)
			} else {
				_, err = repo.TagObject(ref.Hash())
				require.Error(t, err)
			}
		})
	}
}
