package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/coreeng/semver-utils/internal/build"

	igit "github.com/coreeng/semver-utils/pkg/git"
	"github.com/coreeng/semver-utils/pkg/semver"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/spf13/cobra"
)

// outputErrorAndExit prints an error message as JSON and exits the program.
func outputErrorAndExit(errMsg string) {
	if err := json.NewEncoder(os.Stdout).Encode(map[string]string{"error": errMsg}); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
	}
	os.Exit(1)
}

var rootCmd = &cobra.Command{
	Use:   "semver-utils",
	Short: "A utility for managing semantic versioning with Git",
}

// fetch-tag command: calls FetchVersionTag with all its parameters exposed as flags.
var fetchTagCmd = &cobra.Command{
	Use:   "fetch-tag",
	Short: "Fetch the semantic version tag from the repository",
	Run: func(cmd *cobra.Command, args []string) {
		repoPath, _ := cmd.Flags().GetString("repo")
		commitRef, _ := cmd.Flags().GetString("commit")
		prefix, _ := cmd.Flags().GetString("prefix")
		exact, _ := cmd.Flags().GetBool("exact")

		repository, err := git.PlainOpen(repoPath)
		if err != nil {
			outputErrorAndExit(fmt.Sprintf("failed to open repository: %v", err))
		}

		commit, err := igit.FetchCommitObject(repository, commitRef)
		if err != nil {
			outputErrorAndExit(fmt.Sprintf("failed to fetch commit object: %v", err))
		}

		tagName, version, tagCommit, err := igit.FetchVersionTag(repository, commit, prefix, exact)
		if err != nil {
			outputErrorAndExit(fmt.Sprintf("failed to fetch version tag: %v", err))
		}
		if tagName == "" {
			outputErrorAndExit("No matching version tag found.")
		}

		result := map[string]string{
			"tag":     tagName,
			"version": version.String(),
			"commit":  tagCommit.Hash.String(),
		}
		if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
			outputErrorAndExit(fmt.Sprintf("Error encoding JSON: %v", err))
		}
	},
}

// create-tag command: calls FetchVersionTag first. If a previous tag is found,
// it increments the desired version field and then calls CreateVersionTag.
// If no previous tag is found and --create-initial-version=true, it uses the provided --initial-version.
var createTagCmd = &cobra.Command{
	Use:   "create-tag",
	Short: "Create an incremented semantic version tag on the specified commit",
	Run: func(cmd *cobra.Command, args []string) {
		repoPath, _ := cmd.Flags().GetString("repo")
		commitRef, _ := cmd.Flags().GetString("commit")
		prefix, _ := cmd.Flags().GetString("prefix")
		incrementType, _ := cmd.Flags().GetString("increment-type")
		annotated, _ := cmd.Flags().GetBool("annotated")
		prerelease, _ := cmd.Flags().GetString("prerelease")
		buildMetadata, _ := cmd.Flags().GetString("build-metadata")
		pushTag, _ := cmd.Flags().GetBool("push")
		upstream, _ := cmd.Flags().GetString("upstream")
		createInitialVersion, _ := cmd.Flags().GetBool("create-initial-version")
		initialVersionStr, _ := cmd.Flags().GetString("initial-version")

		repository, err := git.PlainOpen(repoPath)
		if err != nil {
			outputErrorAndExit(fmt.Sprintf("failed to open repository: %v", err))
		}

		commit, err := igit.FetchCommitObject(repository, commitRef)
		if err != nil {
			outputErrorAndExit(fmt.Sprintf("failed to fetch commit object: %v", err))
		}

		// Try to fetch a previous version tag
		prevTag, currentVersion, _, err := igit.FetchVersionTag(repository, commit, prefix, false)
		if err != nil {
			// We ignore the error here as it's not critical for version bumping
			prevTag = ""
			currentVersion = semver.SemVer{}
		}

		var newVersion semver.SemVer
		if prevTag != "" {
			switch strings.ToLower(incrementType) {
			case "major":
				newVersion = currentVersion.BumpMajor()
			case "minor":
				newVersion = currentVersion.BumpMinor()
			case "patch":
				newVersion = currentVersion.BumpPatch()
			default:
				outputErrorAndExit("invalid increment type: must be 'major', 'minor', or 'patch'")
			}
		} else {
			// No previous tag found; create an initial version if allowed.
			if createInitialVersion {
				if initialVersionStr == "" {
					outputErrorAndExit("initial-version must be specified when create-initial-version is true")
				}
				newVersion, err = semver.Parse(initialVersionStr)
				if err != nil {
					outputErrorAndExit(fmt.Sprintf("failed to parse initial version: %v", err))
				}
			} else {
				outputErrorAndExit("No previous version tag found and create-initial-version is false.")
			}
		}

		// Apply prerelease and build metadata if provided.
		if prerelease != "" {
			newVersion, err = newVersion.SetPreRelease(semver.PreRelease(prerelease))
			if err != nil {
				outputErrorAndExit(fmt.Sprintf("failed to set prerelease: %v", err))
			}
		}
		if buildMetadata != "" {
			newVersion, err = newVersion.SetBuildMetadata(semver.BuildMetadata(buildMetadata))
			if err != nil {
				outputErrorAndExit(fmt.Sprintf("failed to set build metadata: %v", err))
			}
		}

		// Create the new version tag.
		newTag, err := igit.CreateVersionTag(repository, commit, newVersion, prefix, annotated)
		if err != nil {
			outputErrorAndExit(fmt.Sprintf("failed to create new tag: %v", err))
		}

		response := map[string]interface{}{
			"tag":     newTag,
			"version": newVersion.String(),
			"commit":  commit.Hash.String(),
		}

		// Push the tag to remote if requested.
		if pushTag {
			pushOpts := &git.PushOptions{
				RemoteName: upstream,
				RefSpecs: []config.RefSpec{
					config.RefSpec("refs/tags/" + newTag + ":refs/tags/" + newTag),
				},
			}
			if err := repository.Push(pushOpts); err != nil {
				outputErrorAndExit(fmt.Sprintf("failed to push tag %s to remote %s: %v", newTag, upstream, err))
			}
			response["pushed"] = true
			response["upstream"] = upstream
		}

		if err := json.NewEncoder(os.Stdout).Encode(response); err != nil {
			outputErrorAndExit(fmt.Sprintf("Error encoding JSON: %v", err))
		}
	},
}

// versionCmd prints version/build info.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the semver-utils version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("semver-git version: %s (commit: %s, built at: %s)\n",
			build.BuildVersion, build.BuildCommit, build.BuildDate)
	},
}

func init() {
	// Flags for fetch-tag command.
	fetchTagCmd.Flags().String("repo", ".", "Path to the Git repository")
	fetchTagCmd.Flags().String("commit", "HEAD", "Git reference (commit hash, branch, tag, etc.)")
	fetchTagCmd.Flags().String("prefix", "", "If set, the tag fetched will be formatted as <prefix>/v<semver>")
	fetchTagCmd.Flags().Bool("exact", false, "Match only if tag commit exactly equals the provided commit")

	// Flags for create-tag command.
	createTagCmd.Flags().String("repo", ".", "Path to the Git repository")
	createTagCmd.Flags().String("commit", "HEAD", "Git reference (commit hash, branch, tag, etc.)")
	createTagCmd.Flags().String("prefix", "", "If set, the tag created (and the previous version searched for) will be formatted as <prefix>/v<semver>")
	createTagCmd.Flags().String("increment-type", "patch", "Version increment type: major, minor, or patch")
	createTagCmd.Flags().Bool("annotated", false, "Create an annotated tag")
	createTagCmd.Flags().String("prerelease", "", "Set the prerelease identifier for the new version (optional)")
	createTagCmd.Flags().String("build-metadata", "", "Set the build metadata for the new version (optional)")
	createTagCmd.Flags().Bool("push", false, "Push the new tag to a remote repository after creation? (default is false)")
	createTagCmd.Flags().String("upstream", "origin", "The remote to push the new tag to (default is 'origin')")
	createTagCmd.Flags().Bool("create-initial-version", false, "If true, create an initial version if no previous version tag is found (default is false)")
	createTagCmd.Flags().String("initial-version", "", "Specify the initial semantic version to use if no previous version tag is found (required if create-initial-version is true)")

	// Add subcommands to the root command.
	rootCmd.AddCommand(fetchTagCmd)
	rootCmd.AddCommand(createTagCmd)
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		outputErrorAndExit(err.Error())
	}
}
