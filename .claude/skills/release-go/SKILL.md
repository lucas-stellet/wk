# Go Release

Release a new version of a Go module with semantic versioning.

## When to Apply

Use this skill when the user says:
- "release", "launch", "publish"
- "new version", "bump version"
- "tag version", "create release"

## Process

### 1. Check Current State

```bash
# Get latest tag
git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"

# Check for uncommitted changes
git status --porcelain

# Check if on main/master branch
git branch --show-current
```

### 2. Determine Version Bump

Ask the user which type of release:
- **patch** (v0.1.0 → v0.1.1): Bug fixes, small changes
- **minor** (v0.1.0 → v0.2.0): New features, backward compatible
- **major** (v0.1.0 → v1.0.0): Breaking changes

### 3. Create Release

```bash
# Create annotated tag
git tag -a v{VERSION} -m "Release v{VERSION}"

# Push tag to remote
git push origin main --tags
```

### 4. Verify Release

```bash
# Confirm tag was pushed
git ls-remote --tags origin | grep v{VERSION}

# Show install command
echo "Install with: go install {MODULE_PATH}@v{VERSION}"
```

## Version Calculation

Given current version `vX.Y.Z`:
- **patch**: `vX.Y.(Z+1)`
- **minor**: `vX.(Y+1).0`
- **major**: `v(X+1).0.0`

## Pre-release Checklist

Before releasing, verify:
1. All changes are committed
2. Tests pass (if applicable)
3. Code builds successfully: `go build ./...`
4. On correct branch (main/master)

## Example Output

```
Current version: v0.1.0
Release type: minor
New version: v0.2.0

✓ Created tag v0.2.0
✓ Pushed to origin

Install with:
  go install github.com/user/project@v0.2.0
  go install github.com/user/project@latest
```
