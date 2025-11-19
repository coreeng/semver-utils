# semver-utils

This repository provides Go libraries and CLIs for working with semantic versioning:

- `semver` for interacting with semantic version strings
- `semver-git` for managing semantic version tags in Git repositories

## Installation

### Via Docker (recommended)

Both CLIs are available in a single Docker image from GitHub Container Registry:

```bash
# Pull the latest image
docker pull ghcr.io/coreeng/semver-utils:latest

# Run semver CLI
docker run --rm ghcr.io/coreeng/semver-utils:latest semver --help
docker run --rm ghcr.io/coreeng/semver-utils:latest semver increment minor 1.2.3

# Run semver-git CLI
docker run --rm ghcr.io/coreeng/semver-utils:latest semver-git --help

# Run semver-git in a git repository (mount current directory)
docker run --rm -v $(pwd):/workspace -w /workspace \
  ghcr.io/coreeng/semver-utils:latest semver-git fetch-tag --repo .
```

The Docker image is available for `linux/amd64` and `linux/arm64` platforms.

### Via `brew`

```bash
brew tap coreeng/public
brew install coreeng/public/semver-utils
```

### Via `go install`

```bash
go install github.com/coreeng/semver-utils/cmd/semver@latest
go install github.com/coreeng/semver-utils/cmd/semver-git@latest
```

Once installed, the `semver` and `semver-git` binaries will be available in your Go bin directory.

# Prefixed vs. non-prefixed version tags

Sometimes you want to manage a single authoritative semantic version for an entire repository, sometimes you want to manage multiple semantic version tags within a repository.

All commands on `semver-git` accept the optional `--prefix` parameter:

- If this parameter is provided, the CLI will search for and create Git tags in the format `<prefix>/v<semver>`
- If the `--prefix` parameter is omitted, the Git tags searched and created will be in the format `v<semver>`

The `--prefix` string itself can contain any characters that form a valid Git tag.

# Usage

Detailed usage for both CLIs can be found in the [USAGE.md](USAGE.md) file.

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


