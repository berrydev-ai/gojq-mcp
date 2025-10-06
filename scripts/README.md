# Release Scripts

Scripts to help automate the release process for gojq-mcp.

## Scripts

### `release.sh` (Interactive)

Interactive release script with safety checks and confirmations.

**Usage:**
```bash
./scripts/release.sh v1.0.0
```

**Features:**
- ✅ Validates semantic version format
- ✅ Checks for existing tags
- ✅ Warns if not on main/master branch
- ✅ Detects uncommitted changes and offers to commit
- ✅ Checks if local branch is up to date with remote
- ✅ Shows release summary before tagging
- ✅ Requires confirmation before pushing
- ✅ Provides links to GitHub Actions and Release pages
- ✅ Color-coded output

**Example:**
```bash
# Create a new release
./scripts/release.sh v1.2.3

# Create a pre-release
./scripts/release.sh v2.0.0-beta.1
```

**What it does:**
1. Validates version format (v1.2.3, v1.2.3-beta, etc.)
2. Checks if tag already exists
3. Verifies you're on main/master (with option to override)
4. Commits any uncommitted changes (with prompt)
5. Checks if branch is up to date with remote
6. Shows a summary and asks for confirmation
7. Creates annotated git tag
8. Pushes commit and tag to remote
9. Displays links to track build progress

### `release-ci.sh` (Non-Interactive)

Non-interactive version for CI/automation environments.

**Usage:**
```bash
./scripts/release-ci.sh v1.0.0
```

**Features:**
- ✅ Validates semantic version format
- ✅ Checks for existing tags
- ✅ No prompts or confirmations
- ✅ Suitable for automation

**Use cases:**
- CI/CD pipelines
- Automated release workflows
- Scripts that need to run without user interaction

## Version Format

Both scripts follow [Semantic Versioning](https://semver.org/) with a `v` prefix:

| Type | Format | Example |
|------|--------|---------|
| **Stable** | `vMAJOR.MINOR.PATCH` | `v1.0.0` |
| **Pre-release** | `vMAJOR.MINOR.PATCH-PRERELEASE` | `v1.0.0-alpha.1` |
| | | `v2.0.0-beta` |
| | | `v1.0.0-rc.1` |

**Version components:**
- **MAJOR**: Breaking changes
- **MINOR**: New features (backwards compatible)
- **PATCH**: Bug fixes (backwards compatible)
- **PRERELEASE**: Pre-release identifiers (alpha, beta, rc)

## Workflow

The typical release workflow:

```bash
# 1. Make your changes
git add .
git commit -m "Add new feature"

# 2. Update CHANGELOG.md
#    - Add changes to the [Unreleased] section as you work
#    - Before release, create a new version section
#    - Move items from [Unreleased] to the new version section
#    Example:
#      ## [1.1.0] - 2024-10-06
#      ### Added
#      - New feature XYZ

# 3. Run the release script
./scripts/release.sh v1.1.0

# 4. GitHub Actions automatically:
#    - Runs tests
#    - Builds binaries for all platforms
#    - Creates GitHub Release with changelog content
#    - Attaches binaries to release
```

## Changelog Maintenance

This project uses [Keep a Changelog](https://keepachangelog.com/) format. Update `CHANGELOG.md` as you make changes:

**While developing:**
```markdown
## [Unreleased]

### Added
- New feature description

### Fixed
- Bug fix description
```

**Before releasing version 1.1.0:**
```markdown
## [Unreleased]

## [1.1.0] - 2024-10-06

### Added
- New feature description

### Fixed
- Bug fix description

[Unreleased]: https://github.com/berrydev-ai/gojq-mcp/compare/v1.1.0...HEAD
[1.1.0]: https://github.com/berrydev-ai/gojq-mcp/compare/v1.0.0...v1.1.0
```

**Changelog categories:**
- `Added` - New features
- `Changed` - Changes to existing functionality
- `Deprecated` - Soon-to-be removed features
- `Removed` - Removed features
- `Fixed` - Bug fixes
- `Security` - Security fixes

The release script will check for a changelog entry and warn if missing. The release workflow automatically extracts the changelog content for the version being released and includes it in the GitHub release notes.

## Troubleshooting

### "Tag already exists"
```bash
# List all tags
git tag

# Delete local tag
git tag -d v1.0.0

# Delete remote tag
git push origin :refs/tags/v1.0.0
```

### "Not on main/master branch"
```bash
# Switch to main
git checkout main

# Or force release from current branch
# (script will prompt for confirmation)
```

### "Uncommitted changes"
```bash
# Option 1: Let the script commit them
# (it will prompt you)

# Option 2: Commit manually
git add .
git commit -m "Your message"

# Option 3: Stash changes
git stash
```

### "Branch not up to date"
```bash
# Pull latest changes
git pull origin main

# Or let the script do it
# (it will prompt you)
```

## Manual Release (Without Scripts)

If you prefer to create releases manually:

```bash
# 1. Commit all changes
git add .
git commit -m "Prepare release v1.0.0"
git push origin main

# 2. Create and push tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# 3. GitHub Actions will handle the rest
```

## See Also

- [GitHub Actions Workflows](../.github/workflows/)
- [Semantic Versioning](https://semver.org/)
- [Git Tagging Documentation](https://git-scm.com/book/en/v2/Git-Basics-Tagging)
