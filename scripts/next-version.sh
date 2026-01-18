#!/bin/bash

# Script to parse and increment version
# Gets current version from git tags or defaults to v0.1.0
# Increments the patch version and returns the new version

CURRENT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.1.0")

# Parse version using parameter expansion (no sed/awk needed)
# Format: v0.1.0 -> MAJOR=0, MINOR=1, PATCH=0
VERSION_STRING="${CURRENT_VERSION#v}"  # Remove 'v' prefix
IFS='.' read -r MAJOR MINOR PATCH <<< "$VERSION_STRING"

# Calculate new patch version
NEW_PATCH=$((PATCH + 1))
NEW_VERSION="v${MAJOR}.${MINOR}.${NEW_PATCH}"

echo "$NEW_VERSION"
