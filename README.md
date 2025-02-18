# semver-utils

This repository provides Go libraries and CLIs for working with semantic versioning:

- `semver` for interacting with semantic version strings
- `semver-git` for managing semantic version tags in Git repositories

## Installation

Both CLIs require Go to be installed. Then install each CLI using:

### Via `brew` (recommended)

```bash
brew tap coreeng/tap
brew install coreeng/tap/semver-utils
```

> [!NOTE]
> This will require you to have `HOMEBREW_GITHUB_API_TOKEN` set in your environment to a PAT that can access private coreeng repos

### Via `go install`

```bash
export GOPRIVATE=github.com/coreeng/semver-utils
go install github.com/coreeng/semver-utils/cmd/semver@latest
go install github.com/coreeng/semver-utils/cmd/semver-git@latest
```

Once installed, the `semver` and `semver-git` binaries will be available in your Go bin directory.

## `semver` CLI

Use the `semver` CLI to parse, set, compare, or increment semantic version components. For example:

- `semver get major 1.2.3`
- `semver set patch 1.2.3 4`
- `semver compare gt 1.2.3 1.2.0`
- `semver increment minor 1.2.3`

For complete usage details, run:

```bash
semver --help
```

## `semver-git` CLI

Use the `semver-git` CLI to fetch existing semantic version tags or create new tags on a specific Git commit. For example:

- `semver-git fetch-tag --repo .`
- `semver-git create-tag --increment-type patch --push`

For complete usage details, run:

```bash
semver-git --help
```

### Prefixed vs. non-prefixed version tags

Sometimes you want to manage a single authoritative semantic version for an entire repository, sometimes you want to manage multiple semantic version tags for multiple elements within a repository.

All commands on `semver-git` accept the optional `--prefix` parameter. If this parameter is provided, the CLI will search for and create Git tags in the format `<prefix>/v<semver>`. If the `--prefix` parameter is omitted, the Git tags searched and created will be in the format `v<semver>`

The `--prefix` string itself can contain any characters that form a valid Git tag.
