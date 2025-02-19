# semver-utils CLI Usage Guide

semver-utils provides two command-line utilities for managing semantic versions and semantic version tags on Git repositories.

<!-- TOC -->
* [semver-utils CLI Usage Guide](#semver-utils-cli-usage-guide)
* [`semver-git` usage](#semver-git-usage)
  * [Commands](#commands)
    * [fetch-tag](#fetch-tag)
      * [Syntax](#syntax)
      * [Parameters](#parameters)
      * [Example Usage](#example-usage)
    * [create-tag](#create-tag)
      * [Syntax](#syntax-1)
      * [Parameters](#parameters-1)
      * [Example Usage](#example-usage-1)
  * [Output and Error Handling](#output-and-error-handling)
* [`semver` usage](#semver-usage)
  * [Version](#version)
  * [Compare](#compare)
    * [gt (Greater Than)](#gt-greater-than)
    * [gte (Greater Than or Equal To)](#gte-greater-than-or-equal-to)
    * [eq (Equal To)](#eq-equal-to)
    * [lt (Less Than)](#lt-less-than)
    * [lte (Less Than or Equal To)](#lte-less-than-or-equal-to)
  * [Get](#get)
    * [major](#major)
    * [minor](#minor)
    * [patch](#patch)
    * [prerelease](#prerelease)
    * [buildmetadata](#buildmetadata)
  * [Increment](#increment)
    * [Increment Major](#increment-major)
    * [Increment Minor](#increment-minor)
    * [Increment Patch](#increment-patch)
  * [Set](#set)
    * [Set Major](#set-major)
    * [Set Minor](#set-minor)
    * [Set Patch](#set-patch)
    * [Set Prerelease](#set-prerelease)
    * [Set Buildmetadata](#set-buildmetadata)
  * [Additional Information](#additional-information)
* [Additional Resources](#additional-resources)
<!-- TOC -->

# `semver-git` usage

`semver-git` offers the following two primary commands:

- **fetch-tag**: Search for a semantic version tag from the repository starting from a specific commit
- **create-tag**: Creates a new semantic version tag by incrementing a specified part of an existing version (or by creating an initial version if none exists).

The basic usage of the CLI is as follows:

```bash
semver-git [command] [flags]
```

For help on any command, run:

```bash
semver-git [command] --help
```

---

## Commands

### fetch-tag

Fetches the semantic version tag from the repository based on a given commit reference.

#### Syntax

```
semver-utils fetch-tag \
    [--repo=<repository-path>] \
    [--commit=<git-ref>] \
    [--prefix=<tag-prefix>] \
    [--exact=<true|false>]
```

#### Parameters

| Flag       | Description                                                                               | Default      | Required |
|------------|-------------------------------------------------------------------------------------------|--------------|----------|
| `--repo`   | Path to the Git repository where version tags are maintained.                             | `.`          | No       |
| `--commit` | Git reference identifying the target commit. Can be a commit hash, branch name, tag, etc. | `HEAD`       | No       |
| `--prefix` | If specified, look for semver tags in the format `<prefix>/v<semver>`.                    | `""` (empty) | No       |
| `--exact`  | Boolean flag. If set to `true`, only tags that exactly match the commit are considered.   | `false`      | No       |

#### Example Usage

1. **Fetch the most recent semver tag without a prefix, searching backwards from the HEAD commit**

    ```bash
    semver-git fetch-tag
    ```

2. **Fetch the most recent semver tag with a `backend/` prefix, searching backwards from a specific commit:**

    ```bash
    semver-git fetch-tag --commit=abc123 --prefix=backend
    ```

3. **Fetch the most recent semver tag without a prefix, from a specific commit on a repository:**

    ```bash
    semver-git fetch-tag --commit=abc123 --exact=true
    ```

---

### create-tag

Creates a new semver tag on the specified commit. It first attempts to determine if a previous semver tag exists (with the same `--prefix`, if specified) and then increments the desired part of the version (major, minor, or patch) for the new tag. If no previous semver tag exists, it can optionally create an initial version tag.

#### Syntax

```
semver-git create-tag \
  [--repo=<repository-path>] \
  [--commit=<git-ref>] \
  [--prefix=<tag-prefix>] \
  [--create-initial-version=<true|false>] \
  [--initial-version=<version>] \
  [--increment-type=<major|minor|patch>] \
  [--prerelease=<identifier>] \
  [--build-metadata=<metadata>] \
  [--annotated=<true|false>] \
  [--push=<true|false>] \
  [--upstream=<remote-name>]
```

#### Parameters

| Flag                       | Description                                                                                                                                             | Default      | Required      |
|----------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------|--------------|---------------|
| `--repo`                   | Path to the Git repository where the version tag is to be created.                                                                                      | `.`          | No            |
| `--commit`                 | Git reference identifying the target commit on which the tag will be created. Can be a commit hash, branch name, tag, etc.                              | `HEAD`       | No            |
| `--prefix`                 | If specified, semver tags in the format `<prefix>/v<semver>` will be searched and created.                                                              | `""` (empty) | No            |
| `--increment-type`         | Specifies which part of the version to increment. Supported values are: `major`, `minor`, or `patch`.                                                   | `patch`      | No            |
| `--annotated`              | If set to `true`, creates an annotated Git tag, which includes a message and tagger information.                                                        | `false`      | No            |
| `--prerelease`             | Pre-release identifier (for example, `alpha` or `beta`) to set on the created semver tag. This allows tagging versions such as `1.2.3-alpha`.           | `""` (empty) | No            |
| `--build-metadata`         | Build metadata string to set on the created semver tag. Often used to add additional build or environment information to the tag such as `1.2.3+macos`. | `""` (empty) | No            |
| `--push`                   | If set to `true`, the new tag will be pushed to a remote repository.                                                                                    | `false`      | No            |
| `--upstream`               | The name of the remote repository where the tag should be pushed.                                                                                       | `origin`     | No            |
| `--create-initial-version` | If set to `true`, when no previous semantic tag exists, a new one will be created if `--initial-version` has been specified.                            | `false`      | No            |
| `--initial-version`        | When using `--create-initial-version=true`, this flag must be provided to set the starting semantic version (e.g., `1.0.0`).                            | none         | Conditionally |

#### Example Usage

1. **Search backwards for a semver tag, without prefix (i.e., `vX.Y.Z[-*][+*]`, and create a new semver tag on the HEAD commit with the patch version incremented only if a previous version was found:**

    ```bash
    semver-git create-tag --increment-type=patch
    ```

2. **Same as previous example, but this time an initial version tag of `v0.1.0` will be created if no previous version tag exists:**

    ```bash
    semver-git create-tag --create-initial-version=true --initial-version=0.1.0 --increment-type=patch
    ```

3. **Search backwards for a semver tag, with prefix (i.e., `foobar/vX.Y.Z[-*][+*]`), from a specific commit. If found, increment the major version and create a new tag `foobar/vX+1.0.0` on the specific commit:**

    ```bash
    semver-git create-tag --commit=abc123 --prefix=foobar --increment-type=major --annotated=true
    ```

4. **Create a minor version bump, include build metadata, and push the new tag to an upstream remote:**

    ```bash
    semver-git create-tag --increment-type=minor --build-metadata=build-456 --push=true --upstream=upstream
    ```

---

## Output and Error Handling

- **Successful Execution:**  
  Both commands output a JSON object. For example, a successful fetch might return:

  ```json
  {
      "tag": "v1.2.3",
      "version": "1.2.3",
      "commit": "d4c3b4a..."
  }
  ```

- **Errors:**  
  Errors are output as a JSON object with an "error" key. For example:

  ```json
  {
      "error": "failed to open repository: repository does not exist"
  }
  ```

---

# `semver` usage

The `semver` CLI is a command-line tool for interacting with semantic version strings. It provides commands to display version information, compare semantic versions, extract individual fields, increment parts of a version, and set specific fields.

Below is a comprehensive guide detailing each command, its subcommands, and usage examples.

---

## Version

The `version` command displays the build and version information of the CLI.

**Usage:**

```bash
semver version
```

**Example Output:**
```
semver version: 1.0.0 (commit: abcdef, built at: 2023-10-01T12:34:56Z)
``` 

---

## Compare

The `compare` command is used to compare two semantic versions. It requires one of the comparison subcommands.

**General Usage:**

```bash
semver compare [subcommand] <version1> <version2>
```

The comparison subcommands exit with the following codes:
- **0**: The comparison is true.
- **1**: The comparison is false.
- **2**: An error occurred (e.g., invalid version string or wrong number of arguments).

### gt (Greater Than)

Checks if `version1` is greater than `version2`.

**Usage:**

```bash
semver compare gt <version1> <version2>
```

**Example:**

```bash
semver compare gt 1.2.3 1.2.0
```

---

### gte (Greater Than or Equal To)

Checks if `version1` is greater than or equal to `version2`.

**Usage:**

```bash
semver compare gte <version1> <version2>
```

**Example:**

```bash
semver compare gte 1.2.3 1.2.3
```

---

### eq (Equal To)

Checks if `version1` is exactly equal to `version2`.

**Usage:**

```bash
semver compare eq <version1> <version2>
```

**Example:**

```bash
semver compare eq 1.2.3 1.2.3
```

---

### lt (Less Than)

Checks if `version1` is less than `version2`.

**Usage:**

```bash
semver compare lt <version1> <version2>
```

**Example:**

```bash
semver compare lt 1.2.0 1.2.3
```

---

### lte (Less Than or Equal To)

Checks if `version1` is less than or equal to `version2`.

**Usage:**

```bash
semver compare lte <version1> <version2>
```

**Example:**

```bash
semver compare lte 1.2.3 1.2.3
```

---

## Get

The `get` command extracts a specific field from a semantic version.

**General Usage:**

```bash
semver get [subcommand] <version>
```

### major

Retrieves the **major** version field.

**Usage:**

```bash
semver get major <version>
```

**Example:**

```bash
semver get major 1.2.3-alpha+build123
# Output: 1
```

---

### minor

Retrieves the **minor** version field.

**Usage:**

```bash
semver get minor <version>
```

**Example:**

```bash
semver get minor 1.2.3-alpha+build123
# Output: 2
```

---

### patch

Retrieves the **patch** version field.

**Usage:**

```bash
semver get patch <version>
```

**Example:**

```bash
semver get patch 1.2.3-alpha+build123
# Output: 3
```

---

### prerelease

Retrieves the **prerelease** field.

**Usage:**

```bash
semver get prerelease <version>
```

**Example:**

```bash
semver get prerelease 1.2.3-alpha+build123
# Output: alpha
```

---

### buildmetadata

Retrieves the **build metadata** field.

**Usage:**

```bash
semver get buildmetadata <version>
```

**Example:**

```bash
semver get buildmetadata 1.2.3-alpha+build123
# Output: build123
```

---

## Increment

The `increment` command increases one of the version fields. The minor and patch reset occurs as appropriate when a higher-level field is incremented.

**General Usage:**

```bash
semver increment [subcommand] <version>
```

### Increment Major

Increments the **major** version. Note that when the major version is incremented, the minor and patch fields are reset to 0.

**Usage:**

```bash
semver increment major <version>
```

**Example:**

```bash
semver increment major 1.2.3
# Output: 2.0.0
```

---

### Increment Minor

Increments the **minor** version. When incrementing the minor version, the patch field is reset to 0.

**Usage:**

```bash
semver increment minor <version>
```

**Example:**

```bash
semver increment minor 1.2.3
# Output: 1.3.0
```

---

### Increment Patch

Increments the **patch** version.

**Usage:**

```bash
semver increment patch <version>
```

**Example:**

```bash
semver increment patch 1.2.3
# Output: 1.2.4
```

---

## Set

The `set` command allows you to modify a specific field of a semantic version, generating a new version string with the updated field.

**General Usage:**

```bash
semver set [subcommand] <version> <newValue>
```

### Set Major

Sets the **major** field to a new value. The minor and patch fields remain unchanged.

**Usage:**

```bash
semver set major <version> <newMajor>
```

**Example:**

```bash
semver set major 1.2.3 10
# Output: 10.2.3
```

---

### Set Minor

Sets the **minor** field to a new value while keeping the major and patch fields unchanged.

**Usage:**

```bash
semver set minor <version> <newMinor>
```

**Example:**

```bash
semver set minor 1.2.3 10
# Output: 1.10.3
```

---

### Set Patch

Sets the **patch** field to a new value while retaining the major and minor fields.

**Usage:**

```bash
semver set patch <version> <newPatch>
```

**Example:**

```bash
semver set patch 1.2.3 99
# Output: 1.2.99
```

---

### Set Prerelease

Sets the **prerelease** field to a new value.

**Usage:**

```bash
semver set prerelease <version> <newPrerelease>
```

**Example:**

```bash
semver set prerelease 1.2.3-alpha rc.1
# Output: 1.2.3-rc.1
```

---

### Set Buildmetadata

Sets the **build metadata** field to a new value.

**Usage:**

```bash
semver set buildmetadata <version> <newBuildMetadata>
```

**Example:**

```bash
semver set buildmetadata 1.2.3+build123 release456
# Output: 1.2.3+release456
```

---

## Additional Information

- **Error Handling:**  
  If an invalid version string or incorrect number of arguments is provided, the CLI will display an error message and exit with code 2.

- **Exit Codes Summary:**
    - **0:** Operation successful and the comparison or command evaluated to true.
    - **1:** Comparison evaluated to false or invalid numerical value provided for set operations.
    - **2:** General error (e.g., invalid input syntax).

This guide should serve as a complete reference for using the `semver` CLI. For detailed internal behaviors and further customizations, refer to the built-in help by running:

```bash
semver <command> --help
```

# Additional Resources

- **Semantic Versioning Specification:** [https://semver.org/](https://semver.org/)
- **Git Documentation:** [https://git-scm.com/doc](https://git-scm.com/doc)

---

Happy versioning!
