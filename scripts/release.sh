#!/usr/bin/env bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored messages
info() { echo -e "${BLUE}ℹ ${NC}$1"; }
success() { echo -e "${GREEN}✓${NC} $1"; }
warning() { echo -e "${YELLOW}⚠${NC} $1"; }
error() { echo -e "${RED}✗${NC} $1"; exit 1; }

# Function to validate semantic version format
validate_version() {
    local version=$1
    if [[ ! $version =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+(\.[a-zA-Z0-9]+)*)?$ ]]; then
        error "Invalid version format: $version\nExpected format: v1.2.3 or v1.2.3-beta.1"
    fi
}

# Check if version argument is provided
if [ -z "$1" ]; then
    error "Usage: $0 <version>\nExample: $0 v1.0.0"
fi

VERSION=$1
validate_version "$VERSION"

info "Creating release $VERSION"

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    error "Not in a git repository"
fi

# Check if tag already exists
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    error "Tag $VERSION already exists"
fi

# Check current branch
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" != "main" ] && [ "$CURRENT_BRANCH" != "master" ]; then
    warning "You are on branch '$CURRENT_BRANCH', not 'main' or 'master'"
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        error "Aborted by user"
    fi
fi

# Check for uncommitted changes
if [ -n "$(git status --porcelain)" ]; then
    warning "You have uncommitted changes:"
    git status --short
    echo
    read -p "Commit these changes? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        read -p "Commit message: " COMMIT_MSG
        if [ -z "$COMMIT_MSG" ]; then
            COMMIT_MSG="Prepare release $VERSION"
        fi
        git add .
        git commit -m "$COMMIT_MSG"
        success "Changes committed"
    else
        error "Please commit or stash your changes before tagging"
    fi
fi

# Verify we're up to date with remote
info "Checking remote status..."
git fetch origin "$CURRENT_BRANCH" --quiet || true

LOCAL=$(git rev-parse @)
REMOTE=$(git rev-parse @{u} 2>/dev/null || echo "")

if [ -n "$REMOTE" ] && [ "$LOCAL" != "$REMOTE" ]; then
    warning "Your branch is not up to date with remote"
    read -p "Pull latest changes? (Y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Nn]$ ]]; then
        git pull origin "$CURRENT_BRANCH"
        success "Pulled latest changes"
    fi
fi

# Show summary
echo
echo "═══════════════════════════════════════"
echo "  Release Summary"
echo "═══════════════════════════════════════"
echo "  Version:        $VERSION"
echo "  Branch:         $CURRENT_BRANCH"
echo "  Latest commit:  $(git log -1 --oneline)"
echo "═══════════════════════════════════════"
echo

# Final confirmation
read -p "Create and push tag $VERSION? (Y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Nn]$ ]]; then
    error "Aborted by user"
fi

# Create the tag
info "Creating tag $VERSION..."
git tag -a "$VERSION" -m "Release $VERSION"
success "Tag created"

# Push the commit and tag
info "Pushing to remote..."
git push origin "$CURRENT_BRANCH"
git push origin "$VERSION"
success "Tag pushed to remote"

echo
success "Release $VERSION created successfully!"
echo
info "GitHub Actions will now build binaries for all platforms"
info "Check progress at: https://github.com/$(git remote get-url origin | sed 's/.*github.com[:/]\(.*\)\.git/\1/')/actions"
echo
info "Once the build completes, the release will be available at:"
info "https://github.com/$(git remote get-url origin | sed 's/.*github.com[:/]\(.*\)\.git/\1/')/releases/tag/$VERSION"
