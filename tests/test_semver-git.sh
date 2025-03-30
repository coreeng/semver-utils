#!/usr/bin/env bash
set -uo pipefail

# -----------------------------------------------------------------------------
# Build the project with GoReleaser (snapshot mode)
# -----------------------------------------------------------------------------
cd "$(dirname "${BASH_SOURCE[0]}")/.."

echo "==> Building project with GoReleaser (snapshot mode)..."
make build

# -----------------------------------------------------------------------------
# Detect current platform/architecture and locate the binary
# -----------------------------------------------------------------------------
PLATFORM=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64)
        ARCH="amd64"
        ;;
    i386|i686)
        ARCH="386"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

if [[ "$PLATFORM" == "darwin" ]]; then
    PLATFORM="darwin"
elif [[ "$PLATFORM" == "linux" ]]; then
    PLATFORM="linux"
elif [[ "$PLATFORM" == msys* ]] || [[ "$PLATFORM" == cygwin* ]] || [[ "$PLATFORM" == mingw* ]]; then
    PLATFORM="windows"
else
    echo "Unsupported platform: $PLATFORM"
    exit 1
fi

BINARY_EXTENSION=""
if [ "$PLATFORM" == "windows" ]; then
    BINARY_EXTENSION=".exe"
fi

SEMVER_DIR=$(find ./dist -type d -name "semver-git_${PLATFORM}_${ARCH}*" | head -n 1)
if [ -z "$SEMVER_DIR" ]; then
    echo "❌ Could not find any build output directory for semver-git_${PLATFORM}_${ARCH}* in ./dist/"
    exit 1
fi

BINARY_PATH="$SEMVER_DIR/semver-git${BINARY_EXTENSION}"

if [ ! -f "$BINARY_PATH" ]; then
    echo "❌ Binary not found at: $BINARY_PATH"
    exit 1
fi

chmod +x "$BINARY_PATH"
echo "✅ Found binary: $BINARY_PATH"

# -----------------------------------------------------------------------------
# Helper functions to validate JSON output using jq
# -----------------------------------------------------------------------------
assert_json_valid() {
    local json="$1"
    if ! echo "$json" | jq . >/dev/null 2>&1; then
        echo "Output is not valid JSON: $json"
        return 1
    fi
    return 0
}

assert_json_field() {
    local json="$1"
    local field="$2"
    local expected="$3"
    local actual
    actual=$(echo "$json" | jq -r ".${field}" 2>/dev/null)
    if [ "$actual" != "$expected" ]; then
        echo "Expected JSON field '$field' to be '$expected', got '$actual'"
        return 1
    fi
    return 0
}

# -----------------------------------------------------------------------------
# Test framework functions
# -----------------------------------------------------------------------------
TESTS_RUN=0
TESTS_PASSED=0

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# run_test expects a test name and a function to call.
run_test() {
    local name="$1"
    shift
    ((TESTS_RUN++))
    echo "Running test: $name"
    "$@"
    local exit_code=$?
    if [ $exit_code -eq 0 ]; then
        echo -e "${GREEN}✅ [PASS]${NC} $name"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}❌ [FAIL]${NC} $name"
        return 1
    fi
}

# -----------------------------------------------------------------------------
# Helper functions for Git repo setup
# -----------------------------------------------------------------------------
setup_repo() {
    local repo_dir
    repo_dir=$(mktemp -d)
    git -C "$repo_dir" init -q
    git -C "$repo_dir" config user.name "Test User"
    git -C "$repo_dir" config user.email "test@example.com"
    echo "initial commit" > "$repo_dir/file.txt"
    git -C "$repo_dir" add file.txt
    git -C "$repo_dir" commit -q -m "Initial commit"
    echo "$repo_dir"
}

create_commit() {
    local repo_dir="$1"
    local content="$2"
    local message="$3"
    echo "$content" > "$repo_dir/file.txt"
    git -C "$repo_dir" add file.txt
    git -C "$repo_dir" commit -q -m "$message"
}

# -----------------------------------------------------------------------------
# Tests for fetch-tag command
# -----------------------------------------------------------------------------
echo "==> Testing fetch-tag command"

test_fetch_tag_basic_no_prefix() {
    local repo
    repo=$(setup_repo)
    local commit_hash
    commit_hash=$(git -C "$repo" rev-parse HEAD)
    git -C "$repo" tag "1.2.3"
    output=$("$BINARY_PATH" fetch-tag --repo "$repo" --commit HEAD)
    # Validate JSON output and fields
    assert_json_valid "$output" || return 1
    assert_json_field "$output" "tag" "1.2.3" || return 1
    assert_json_field "$output" "version" "1.2.3" || return 1
    assert_json_field "$output" "commit" "$commit_hash" || return 1
    return 0
}
run_test "fetch-tag basic (no prefix)" test_fetch_tag_basic_no_prefix

test_fetch_tag_with_prefix() {
    local repo
    repo=$(setup_repo)
    local commit_hash
    commit_hash=$(git -C "$repo" rev-parse HEAD)
    # Create tag using the new <prefix>/v<semver> format.
    git -C "$repo" tag "release/v1.2.3"
    output=$("$BINARY_PATH" fetch-tag --repo "$repo" --commit HEAD --prefix release)
    assert_json_valid "$output" || return 1
    assert_json_field "$output" "tag" "release/v1.2.3" || return 1
    assert_json_field "$output" "version" "1.2.3" || return 1
    assert_json_field "$output" "commit" "$commit_hash" || return 1
    return 0
}
run_test "fetch-tag with prefix" test_fetch_tag_with_prefix

test_fetch_tag_exact_no_match() {
    local repo
    repo=$(setup_repo)
    create_commit "$repo" "changed content" "Second commit"
    local first_commit
    first_commit=$(git -C "$repo" rev-list --max-parents=0 HEAD)
    git -C "$repo" tag "1.2.3" "$first_commit"
    output=$("$BINARY_PATH" fetch-tag --repo "$repo" --commit HEAD --exact)
    assert_json_valid "$output" || return 1
    assert_json_field "$output" "error" "No matching version tag found." || return 1
    return 0
}
run_test "fetch-tag with exact flag (no match)" test_fetch_tag_exact_no_match

# -----------------------------------------------------------------------------
# Tests for create-tag command
# -----------------------------------------------------------------------------
echo "==> Testing create-tag command"

test_create_tag_basic_patch() {
    local repo
    repo=$(setup_repo)
    git -C "$repo" tag "1.2.3"
    output=$("$BINARY_PATH" create-tag --repo "$repo" --commit HEAD --increment-type patch)
    assert_json_valid "$output" || return 1
    # Updated expected tag with "v" prefix.
    assert_json_field "$output" "tag" "v1.2.4" || return 1
    assert_json_field "$output" "version" "1.2.4" || return 1
    if ! git -C "$repo" tag | grep -qx "v1.2.4"; then
        echo "Tag v1.2.4 not found in repository."
        return 1
    fi
    return 0
}
run_test "create-tag basic (patch increment)" test_create_tag_basic_patch

test_create_tag_with_prefix_minor() {
    local repo
    repo=$(setup_repo)
    # Create initial tag using prefix format.
    git -C "$repo" tag "release/v1.2.3"
    output=$("$BINARY_PATH" create-tag --repo "$repo" --commit HEAD --prefix release --increment-type minor)
    assert_json_valid "$output" || return 1
    # Expect new tag to be formatted with prefix and a "v" (e.g., release/v1.3.0)
    assert_json_field "$output" "tag" "release/v1.3.0" || return 1
    assert_json_field "$output" "version" "1.3.0" || return 1
    if ! git -C "$repo" tag | grep -qx "release/v1.3.0"; then
        echo "Tag release/v1.3.0 not found in repository."
        return 1
    fi
    return 0
}
run_test "create-tag with prefix (minor increment)" test_create_tag_with_prefix_minor

test_create_tag_annotated_prerelease_build() {
    local repo
    repo=$(setup_repo)
    git -C "$repo" tag "1.2.3"
    output=$("$BINARY_PATH" create-tag --repo "$repo" --commit HEAD --increment-type major --annotated --prerelease rc.1 --build-metadata build123)
    assert_json_valid "$output" || return 1
    # Updated expected tag with "v" prefix.
    assert_json_field "$output" "tag" "v2.0.0-rc.1+build123" || return 1
    assert_json_field "$output" "version" "2.0.0-rc.1+build123" || return 1
    if ! git -C "$repo" tag | grep -qx "v2.0.0-rc.1+build123"; then
        echo "Tag v2.0.0-rc.1+build123 not found in repository."
        return 1
    fi
    return 0
}
run_test "create-tag annotated with prerelease & build metadata" test_create_tag_annotated_prerelease_build

# New test case for incrementing the major version
test_create_tag_basic_major() {
    local repo
    repo=$(setup_repo)
    git -C "$repo" tag "1.2.3"
    output=$("$BINARY_PATH" create-tag --repo "$repo" --commit HEAD --increment-type major)
    assert_json_valid "$output" || return 1
    # Updated expected tag with "v" prefix.
    assert_json_field "$output" "tag" "v2.0.0" || return 1
    assert_json_field "$output" "version" "2.0.0" || return 1
    if ! git -C "$repo" tag | grep -qx "v2.0.0"; then
        echo "Tag v2.0.0 not found in repository."
        return 1
    fi
    return 0
}
run_test "create-tag basic (major increment)" test_create_tag_basic_major

test_create_tag_with_push() {
    local remote_repo
    remote_repo=$(mktemp -d)
    git init --bare -q "$remote_repo"

    local repo
    repo=$(setup_repo)
    git -C "$repo" remote add origin "$remote_repo"
    git -C "$repo" tag "1.2.3"
    output=$("$BINARY_PATH" create-tag --repo "$repo" --commit HEAD --increment-type patch --push --upstream origin)
    assert_json_valid "$output" || return 1
    # Updated expected tag with "v" prefix.
    assert_json_field "$output" "tag" "v1.2.4" || return 1
    assert_json_field "$output" "version" "1.2.4" || return 1
    # Validate pushed flag and upstream
    pushed=$(echo "$output" | jq -r ".pushed")
    if [ "$pushed" != "true" ]; then
        echo "Expected pushed to be true, got $pushed"
        return 1
    fi
    assert_json_field "$output" "upstream" "origin" || return 1

    remote_tags=$(git ls-remote --tags "$remote_repo")
    if ! echo "$remote_tags" | grep -q "refs/tags/v1.2.4"; then
        echo "Tag v1.2.4 not found in remote repository."
        return 1
    fi
    return 0
}
run_test "create-tag with push to remote" test_create_tag_with_push

# -----------------------------------------------------------------------------
# Additional tests for new command parameters
# -----------------------------------------------------------------------------

test_create_tag_initial_version() {
    local repo
    repo=$(setup_repo)
    # No previous tag exists. Use --create-initial-version with --initial-version.
    output=$("$BINARY_PATH" create-tag --repo "$repo" --commit HEAD --increment-type patch --create-initial-version --initial-version 1.0.0)
    assert_json_valid "$output" || return 1
    # Updated expected tag with "v" prefix.
    assert_json_field "$output" "tag" "v1.0.0" || return 1
    assert_json_field "$output" "version" "1.0.0" || return 1
    if ! git -C "$repo" tag | grep -qx "v1.0.0"; then
        echo "Tag v1.0.0 not found in repository."
        return 1
    fi
    return 0
}
run_test "create-tag with create-initial-version" test_create_tag_initial_version

test_version_command() {
    local output
    output=$("$BINARY_PATH" version)
    # The version command prints plain text version info.
    if [[ "$output" != semver-git* ]]; then
         echo "Version output does not start with 'semver-git'"
         return 1
    fi
    return 0
}
run_test "version command" test_version_command

# -----------------------------------------------------------------------------
# Summary
# -----------------------------------------------------------------------------
echo
echo "Tests run: $TESTS_RUN"
echo "Tests passed: $TESTS_PASSED"
if [ "$TESTS_RUN" -eq "$TESTS_PASSED" ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed.${NC}"
    exit 1
fi
