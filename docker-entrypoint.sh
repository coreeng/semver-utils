#!/bin/sh
set -e

# Router script to call the appropriate binary
# Usage: docker run ghcr.io/coreeng/semver-utils:latest <tool> [args...]
# Where <tool> is either "semver" or "semver-git"

TOOL="${1:-}"

case "$TOOL" in
  semver)
    shift
    exec /usr/local/bin/semver "$@"
    ;;
  semver-git)
    shift
    exec /usr/local/bin/semver-git "$@"
    ;;
  *)
    # If no recognized tool specified, default to semver for backward compatibility
    # This allows "docker run image --help" to work
    exec /usr/local/bin/semver "$@"
    ;;
esac
