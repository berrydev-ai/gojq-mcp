#!/usr/bin/env bash
set -e

# Non-interactive version for CI/automation
# Usage: ./scripts/release-ci.sh v1.0.0

VERSION=$1

if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 v1.0.0"
    exit 1
fi

# Validate version format
if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+(\.[a-zA-Z0-9]+)*)?$ ]]; then
    echo "Invalid version format: $VERSION"
    echo "Expected format: v1.2.3 or v1.2.3-beta.1"
    exit 1
fi

# Check if tag already exists
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    echo "Tag $VERSION already exists"
    exit 1
fi

# Create and push tag
echo "Creating and pushing tag $VERSION..."
git tag -a "$VERSION" -m "Release $VERSION"
git push origin "$VERSION"

echo "Release $VERSION created successfully!"
