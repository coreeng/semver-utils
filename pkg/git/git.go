package git

import (
	"fmt"
	"strings"

	"github.com/coreeng/semver-utils/pkg/semver"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// FetchCommitObject retrieves the commit associated with the given Git reference.
// It accepts HEAD, branch, tag, commit hash, or annotated tag.
func FetchCommitObject(repo *git.Repository, gitRef string) (*object.Commit, error) {
	var hash plumbing.Hash
	var err error

	if gitRef == "HEAD" {
		headRef, err := repo.Head()
		if err != nil {
			return nil, fmt.Errorf("failed to resolve HEAD: %w", err)
		}
		hash = headRef.Hash()
	} else {
		hashPtr, err := repo.ResolveRevision(plumbing.Revision(gitRef))
		if err != nil {
			return nil, fmt.Errorf("failed to resolve git ref '%s': %w", gitRef, err)
		}
		hash = *hashPtr
	}

	commit, err := repo.CommitObject(hash)
	if err == nil {
		return commit, nil
	}

	tagObj, err := repo.TagObject(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch commit object: %w", err)
	}
	commit, err = repo.CommitObject(tagObj.Target)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch commit object from annotated tag: %w", err)
	}
	return commit, nil
}

// FetchVersionTag searches for a semantic version tag in the repository that matches the specified prefix.
// If exactCommit is true, only tags pointing exactly to targetCommit are considered.
// Otherwise, it selects the most recent tag from the commit history not after targetCommit.
func FetchVersionTag(repo *git.Repository, targetCommit *object.Commit, prefix string, exactCommit bool) (string, semver.SemVer, *object.Commit, error) {
	tags, err := repo.Tags()
	if err != nil {
		return "", semver.SemVer{}, nil, fmt.Errorf("failed to retrieve tags: %w", err)
	}

	commitCache := make(map[string]*object.Commit)
	var foundTag string
	var foundVersion semver.SemVer
	var foundCommit *object.Commit
	targetTime := targetCommit.Committer.When

	err = tags.ForEach(func(ref *plumbing.Reference) error {
		candidateTagName := ref.Name().Short()
		var versionString string

		if prefix != "" {
			tagPrefix := prefix + "/"
			if !strings.HasPrefix(candidateTagName, tagPrefix) {
				return nil
			}
			versionString = strings.TrimPrefix(candidateTagName, tagPrefix)
		} else {
			versionString = candidateTagName
		}

		if !semver.FullPattern.MatchString(versionString) {
			return nil
		}

		candidateVersion, err := semver.Parse(versionString)
		if err != nil {
			return nil
		}

		refStr := ref.Name().String()
		var candidateCommit *object.Commit
		if cached, ok := commitCache[refStr]; ok {
			candidateCommit = cached
		} else {
			candidateCommit, err = FetchCommitObject(repo, refStr)
			if err != nil {
				return nil
			}
			commitCache[refStr] = candidateCommit
		}

		if exactCommit {
			if candidateCommit.Hash != targetCommit.Hash {
				return nil
			}
			if foundCommit == nil || candidateVersion.Compare(foundVersion) > 0 {
				foundCommit = candidateCommit
				foundTag = candidateTagName
				foundVersion = candidateVersion
			}
		} else {
			if candidateCommit.Committer.When.After(targetTime) {
				return nil
			}
			if foundCommit == nil || candidateCommit.Committer.When.After(foundCommit.Committer.When) {
				foundCommit = candidateCommit
				foundTag = candidateTagName
				foundVersion = candidateVersion
			} else if candidateCommit.Hash == foundCommit.Hash && candidateVersion.Compare(foundVersion) > 0 {
				foundTag = candidateTagName
				foundVersion = candidateVersion
			}
		}
		return nil
	})

	if err != nil {
		return "", semver.SemVer{}, nil, err
	}

	return foundTag, foundVersion, foundCommit, nil
}

// CreateVersionTag creates a new Git tag for the given targetCommit with the specified semantic version.
// It constructs the new tag name using an optional prefix. If annotated is true, the tag will include
// a message and tagger information. It returns the new tag name or an error.
func CreateVersionTag(repo *git.Repository, targetCommit *object.Commit, version semver.SemVer, prefix string, annotated bool) (string, error) {
	newTagName := fmt.Sprintf("v%s", version.String())
	if prefix != "" {
		newTagName = prefix + "/" + newTagName
	}

	var tagOpts *git.CreateTagOptions
	if annotated {
		tagOpts = &git.CreateTagOptions{
			Message: fmt.Sprintf("Version %s", version.String()),
			Tagger:  &targetCommit.Committer,
		}
	}

	if _, err := repo.CreateTag(newTagName, targetCommit.Hash, tagOpts); err != nil {
		return "", fmt.Errorf("failed to create tag: %w", err)
	}

	return newTagName, nil
}
