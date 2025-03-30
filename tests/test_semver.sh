#!/usr/bin/env bash
set -uo pipefail

# Always operate from the repository root:
# This allows your relative paths (like ./dist) to remain valid
cd "$(dirname "${BASH_SOURCE[0]}")/.."

echo "==> Building project with GoReleaser (snapshot mode)..."
make build

# -----------------------------------------------------------------------------
# Detect current platform/architecture
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

# -----------------------------------------------------------------------------
# Locate the "semver" binary in ./dist/ - flexible approach
# -----------------------------------------------------------------------------
SEMVER_DIR=$(find ./dist -type d -name "semver_${PLATFORM}_${ARCH}*" | head -n 1)

if [ -z "$SEMVER_DIR" ]; then
    echo "❌ Could not find any build output directory for semver_${PLATFORM}_${ARCH}* in ./dist/"
    exit 1
fi

BINARY_PATH="$SEMVER_DIR/semver${BINARY_EXTENSION}"

if [ ! -f "$BINARY_PATH" ]; then
    echo "❌ Binary not found at: $BINARY_PATH"
    exit 1
fi

chmod +x "$BINARY_PATH"
echo "✅ Found binary: $BINARY_PATH"

echo "==> Proceeding to run tests..."

# -----------------------------------------------------------------------------
# Test harness
# -----------------------------------------------------------------------------
TESTS_RUN=0
TESTS_PASSED=0

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No color

run_test() {
    local name="$1"
    local cmd="$2"
    local expected_output="$3"
    local expected_exit_code="$4"

    ((TESTS_RUN++))
    echo "Running test: $name"

    # Capture command output and exit code
    output=$($cmd 2>&1)
    exit_code=$?

    # Compare exit code
    if [ "$exit_code" -ne "$expected_exit_code" ]; then
        echo -e "${RED}❌ [FAIL]${NC} $name"
        echo "  Expected exit code: $expected_exit_code"
        echo "  Got exit code:      $exit_code"
        return
    fi

    # Compare output exactly
    if [ "$output" != "$expected_output" ]; then
        echo -e "${RED}❌ [FAIL]${NC} $name"
        echo "  Expected output: $expected_output"
        echo "  Got output:      $output"
        return
    fi

    echo -e "${GREEN}✅ [PASS]${NC} $name"
    ((TESTS_PASSED++))
}

run_test_contains() {
  local name="$1"
  local cmd="$2"
  local substring="$3"
  local expected_exit_code="$4"

  ((TESTS_RUN++))
  echo "Running test: $name"

  set +e
  output=$($cmd 2>&1)
  exit_code=$?
  set -e

  if [ "$exit_code" -ne "$expected_exit_code" ]; then
      echo -e "${RED}❌ [FAIL]${NC} $name"
      echo "  Expected exit code: $expected_exit_code"
      echo "  Got exit code:      $exit_code"
      return
  fi

  if [[ "$output" != *"$substring"* ]]; then
      echo -e "${RED}❌ [FAIL]${NC} $name"
      echo "  Substring not found: $substring"
      echo "  Full output:         $output"
      return
  fi

  echo -e "${GREEN}✅ [PASS]${NC} $name"
  ((TESTS_PASSED++))
}

# -----------------------------------------------------------------------------
# 1. increment command tests
# -----------------------------------------------------------------------------
run_test "Increment major" \
    "$BINARY_PATH increment major 1.2.3" \
    "2.0.0" \
    0

run_test "Increment minor" \
    "$BINARY_PATH increment minor 1.2.3" \
    "1.3.0" \
    0

run_test "Increment patch" \
    "$BINARY_PATH increment patch 1.2.3" \
    "1.2.4" \
    0

# -----------------------------------------------------------------------------
# 2. compare command tests
# -----------------------------------------------------------------------------
# eq tests
run_test "Compare eq (1.2.3 == 1.2.3 => exit 0)" \
    "$BINARY_PATH compare eq 1.2.3 1.2.3" \
    "" \
    0

run_test "Compare eq (1.2.3 == 1.2.4 => exit 1)" \
    "$BINARY_PATH compare eq 1.2.3 1.2.4" \
    "" \
    1

# gt tests
run_test "Compare gt (1.2.4 > 1.2.3 => exit 0)" \
    "$BINARY_PATH compare gt 1.2.4 1.2.3" \
    "" \
    0

run_test "Compare gt (1.2.3 !> 1.2.3 => exit 1)" \
    "$BINARY_PATH compare gt 1.2.3 1.2.3" \
    "" \
    1

run_test "Compare gt (1.2.2 < 1.2.3 => exit 1)" \
    "$BINARY_PATH compare gt 1.2.2 1.2.3" \
    "" \
    1

# gte tests
run_test "Compare gte (1.2.3 >= 1.2.3 => exit 0)" \
    "$BINARY_PATH compare gte 1.2.3 1.2.3" \
    "" \
    0

run_test "Compare gte (1.2.3 >= 1.2.2 => exit 0)" \
    "$BINARY_PATH compare gte 1.2.3 1.2.2" \
    "" \
    0

run_test "Compare gte (1.2.3 < 1.2.4 => exit 1)" \
    "$BINARY_PATH compare gte 1.2.3 1.2.4" \
    "" \
    1

# lt tests
run_test "Compare lt (1.2.3 < 1.2.4 => exit 0)" \
    "$BINARY_PATH compare lt 1.2.3 1.2.4" \
    "" \
    0

run_test "Compare lt (1.2.3 !< 1.2.3 => exit 1)" \
    "$BINARY_PATH compare lt 1.2.3 1.2.3" \
    "" \
    1

run_test "Compare lt (1.2.4 > 1.2.3 => exit 1)" \
    "$BINARY_PATH compare lt 1.2.4 1.2.3" \
    "" \
    1

# lte tests
run_test "Compare lte (1.2.3 <= 1.2.4 => exit 0)" \
    "$BINARY_PATH compare lte 1.2.3 1.2.4" \
    "" \
    0

run_test "Compare lte (1.2.3 <= 1.2.3 => exit 0)" \
    "$BINARY_PATH compare lte 1.2.3 1.2.3" \
    "" \
    0

run_test "Compare lte (1.2.4 !<= 1.2.3 => exit 1)" \
    "$BINARY_PATH compare lte 1.2.4 1.2.3" \
    "" \
    1

# -----------------------------------------------------------------------------
# 3. get command tests
# -----------------------------------------------------------------------------
run_test "Get major (1.2.3-alpha+build => 1)" \
    "$BINARY_PATH get major 1.2.3-alpha+build" \
    "1" \
    0

run_test "Get minor (1.2.3-alpha+build => 2)" \
    "$BINARY_PATH get minor 1.2.3-alpha+build" \
    "2" \
    0

run_test "Get patch (1.2.3-alpha+build => 3)" \
    "$BINARY_PATH get patch 1.2.3-alpha+build" \
    "3" \
    0

run_test "Get prerelease (1.2.3-alpha+build => alpha)" \
    "$BINARY_PATH get prerelease 1.2.3-alpha+build" \
    "alpha" \
    0

run_test "Get buildmetadata (1.2.3-alpha+build => build)" \
    "$BINARY_PATH get buildmetadata 1.2.3-alpha+build" \
    "build" \
    0

# -----------------------------------------------------------------------------
# 4. set command tests
# -----------------------------------------------------------------------------
run_test "Set major (1.2.3 => 5.2.3)" \
    "$BINARY_PATH set major 1.2.3 5" \
    "5.2.3" \
    0

run_test "Set minor (1.2.3 => 1.9.3)" \
    "$BINARY_PATH set minor 1.2.3 9" \
    "1.9.3" \
    0

run_test "Set patch (1.2.3 => 1.2.99)" \
    "$BINARY_PATH set patch 1.2.3 99" \
    "1.2.99" \
    0

run_test "Set prerelease (1.2.3 => rc.1)" \
    "$BINARY_PATH set prerelease 1.2.3 rc.1" \
    "1.2.3-rc.1" \
    0

run_test "Set buildmetadata (1.2.3 => 1.2.3-alpha+build123)" \
    "$BINARY_PATH set buildmetadata 1.2.3-alpha build123" \
    "1.2.3-alpha+build123" \
    0

# -----------------------------------------------------------------------------
# 5. version command test
# -----------------------------------------------------------------------------
run_test_contains "Version command" "$BINARY_PATH version" "semver version:" 0

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
