# Release Process

This document describes the complete workflow for creating DevOpsMaestro releases.

---

## 📋 Version Numbering

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR** version (X.0.0) - Incompatible API changes, breaking changes
- **MINOR** version (0.X.0) - New features, backward compatible
- **PATCH** version (0.0.X) - Bug fixes, backward compatible

### Examples
- `v0.2.0` - Added theme system (new feature, no breaking changes)
- `v0.2.1` - Fixed theme rendering bug (bug fix)
- `v1.0.0` - First stable release (major milestone)

---

## ✅ Pre-Release Checklist

Before starting the release process:

- [ ] All planned features completed
- [ ] All tests passing: `go test ./...`
- [ ] Code formatted: `go fmt ./...`
- [ ] No lint errors: `go vet ./...`
- [ ] Binary builds successfully: `go build -o dvm`
- [ ] Manual testing completed
- [ ] Documentation updated

---

## 🚀 Release Steps

### Step 1: Prepare Repository

#### 1.1 Update CHANGELOG.md

Add new version entry with release date:

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Added
- New features

### Changed
- Changes to existing features

### Fixed
- Bug fixes
```

**Categories:**
- `Added` - New features
- `Changed` - Changes to existing functionality
- `Deprecated` - Soon-to-be removed features
- `Removed` - Removed features
- `Fixed` - Bug fixes
- `Security` - Security fixes

#### 1.2 Update Version References

Update README.md installation URLs to point to new version:

```markdown
curl -L https://github.com/rmkohlman/devopsmaestro/releases/download/vX.Y.Z/dvm-darwin-arm64 -o dvm
```

#### 1.3 Commit Changes

```bash
git add CHANGELOG.md README.md
git commit -m "chore: prepare vX.Y.Z release"
git push origin main
```

---

### Step 2: Create Git Tag

```bash
# Create annotated tag
git tag -a vX.Y.Z -m "Release vX.Y.Z: <short description>"

# Push tag to GitHub
git push origin vX.Y.Z
```

**Example:**
```bash
git tag -a v0.2.0 -m "Release v0.2.0: Theme system + YAML colorization"
git push origin v0.2.0
```

---

### Step 3: Build Cross-Platform Binaries

#### 3.1 Set Build Variables

```bash
VERSION="X.Y.Z"
BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)
COMMIT=$(git rev-parse --short HEAD)
LDFLAGS="-X 'devopsmaestro/cmd.Version=${VERSION}' -X 'devopsmaestro/cmd.BuildTime=${BUILD_TIME}' -X 'devopsmaestro/cmd.Commit=${COMMIT}'"
```

#### 3.2 Build for All Platforms

```bash
# macOS Apple Silicon (M1/M2/M3)
GOOS=darwin GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o dvm-darwin-arm64

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o dvm-darwin-amd64

# Linux x86_64
GOOS=linux GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o dvm-linux-amd64

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o dvm-linux-arm64
```

#### 3.3 Generate Checksums

```bash
shasum -a 256 dvm-darwin-arm64 dvm-darwin-amd64 dvm-linux-amd64 dvm-linux-arm64 > checksums.txt
```

#### 3.4 Verify Builds

Test at least one binary:

```bash
./dvm-darwin-arm64 version
# Should show vX.Y.Z, correct build time, and commit hash
```

---

### Step 4: Create GitHub Release

#### 4.1 Navigate to Releases

Go to: `https://github.com/rmkohlman/devopsmaestro/releases/new`

#### 4.2 Fill Release Form

**Choose a tag:** Select `vX.Y.Z` (existing tag)

**Release title:** `vX.Y.Z - Feature Name`

**Description:** Copy from CHANGELOG.md and enhance with:
- Feature overview
- Installation instructions
- Theme table (if applicable)
- Usage examples
- Links to documentation

**Example structure:**
```markdown
## 🎨 Feature Name

Brief description of what's new.

---

### ✨ Added
- Feature 1
- Feature 2

### 🔧 Changed
- Change 1

---

## 🚀 Installation

### Download Pre-Built Binary

**macOS (Apple Silicon):**
```bash
curl -L https://github.com/rmkohlman/devopsmaestro/releases/download/vX.Y.Z/dvm-darwin-arm64 -o dvm
chmod +x dvm
sudo mv dvm /usr/local/bin/
```

[... other platforms ...]

---

## 📚 Documentation

- [README](...)
- [CHANGELOG](...)

---

**Full Changelog:** https://github.com/rmkohlman/devopsmaestro/blob/main/CHANGELOG.md
```

#### 4.3 Upload Assets

Drag and drop these 5 files:
1. `dvm-darwin-arm64`
2. `dvm-darwin-amd64`
3. `dvm-linux-amd64`
4. `dvm-linux-arm64`
5. `checksums.txt`

#### 4.4 Publish

- ✅ Check "Set as the latest release"
- ⬜ Leave "Set as a pre-release" unchecked (unless it's a beta)
- Click **"Publish release"**

---

### Step 5: Verify Release

#### 5.1 Check Release Page

Visit: `https://github.com/rmkohlman/devopsmaestro/releases/latest`

Verify:
- [ ] Release is visible
- [ ] Version number correct
- [ ] Description formatted properly
- [ ] All 5 assets present
- [ ] "Latest" badge visible

#### 5.2 Test Binary Download

```bash
# Download binary
curl -L https://github.com/rmkohlman/devopsmaestro/releases/download/vX.Y.Z/dvm-darwin-arm64 -o dvm-test
chmod +x dvm-test

# Test it works
./dvm-test version
# Should show correct version

# Verify checksum
curl -L https://github.com/rmkohlman/devopsmaestro/releases/download/vX.Y.Z/checksums.txt -o checksums-verify.txt
shasum -a 256 dvm-test | grep -f checksums-verify.txt
# Should match
```

#### 5.3 Update README Badges (Optional)

If not done earlier, add/update release badge:

```markdown
[![Release](https://img.shields.io/github/v/release/rmkohlman/devopsmaestro)](https://github.com/rmkohlman/devopsmaestro/releases/latest)
```

Commit and push:
```bash
git add README.md
git commit -m "docs: update release badge to vX.Y.Z"
git push origin main
```

---

### Step 6: Announce (Optional)

#### 6.1 Create GitHub Discussion

Navigate to: `https://github.com/rmkohlman/devopsmaestro/discussions`

**Category:** Announcements

**Title:** `🎉 DevOpsMaestro vX.Y.Z Released - Feature Name`

**Content:**
```markdown
Excited to announce **DevOpsMaestro vX.Y.Z** is now available! 🚀

## 🎨 What's New

Brief feature overview with highlights.

### Try It Now

```bash
curl -L https://github.com/rmkohlman/devopsmaestro/releases/download/vX.Y.Z/dvm-darwin-arm64 -o dvm
chmod +x dvm
./dvm version
```

### Full Details

See the [release notes](https://github.com/rmkohlman/devopsmaestro/releases/tag/vX.Y.Z) for complete details.

### Feedback Welcome

Try it out and let us know what you think!
```

---

## 🤖 Future Automation

### GoReleaser (Planned for v0.3.0+)

Automate the build process with [GoReleaser](https://goreleaser.com/):

**Benefits:**
- Automated cross-platform builds
- Automatic checksum generation
- GitHub Release creation
- Homebrew tap updates
- Docker image publishing

**Setup:**
```yaml
# .goreleaser.yml
builds:
  - binary: dvm
    goos: [darwin, linux]
    goarch: [amd64, arm64]
    ldflags:
      - -X devopsmaestro/cmd.Version={{.Version}}
      - -X devopsmaestro/cmd.BuildTime={{.Date}}
      - -X devopsmaestro/cmd.Commit={{.ShortCommit}}
```

**Usage:**
```bash
goreleaser release --clean
```

---

## 📝 Release Checklist Template

Copy this checklist for each release:

```markdown
## Release vX.Y.Z Checklist

### Pre-Release
- [ ] All features completed
- [ ] All tests passing
- [ ] Code formatted and linted
- [ ] Binary builds successfully
- [ ] Manual testing completed

### Preparation
- [ ] CHANGELOG.md updated
- [ ] README.md version URLs updated
- [ ] Changes committed and pushed
- [ ] Git tag created and pushed

### Build
- [ ] Build variables set
- [ ] macOS arm64 built
- [ ] macOS amd64 built
- [ ] Linux amd64 built
- [ ] Linux arm64 built
- [ ] Checksums generated
- [ ] At least one binary tested

### GitHub Release
- [ ] Release created on GitHub
- [ ] Title and description filled
- [ ] All 5 assets uploaded
- [ ] Published as latest

### Verification
- [ ] Release page visible
- [ ] Binary download works
- [ ] Checksum verifies
- [ ] Version command shows correct info

### Post-Release
- [ ] README badges updated
- [ ] Announcement discussion created
- [ ] Social media updated (if applicable)
```

---

## 🐛 Troubleshooting

### Build Fails

**Problem:** Cross-compilation errors

**Solution:**
```bash
# Ensure GOOS/GOARCH are correct
echo $GOOS $GOARCH

# Clear build cache
go clean -cache

# Try again
GOOS=linux GOARCH=amd64 go build -o dvm-linux-amd64
```

### Tag Already Exists

**Problem:** `fatal: tag 'vX.Y.Z' already exists`

**Solution:**
```bash
# Delete local tag
git tag -d vX.Y.Z

# Delete remote tag (careful!)
git push --delete origin vX.Y.Z

# Create new tag
git tag -a vX.Y.Z -m "Release vX.Y.Z: ..."
git push origin vX.Y.Z
```

### Binary Won't Run

**Problem:** "Permission denied" or "Command not found"

**Solution:**
```bash
# Make executable
chmod +x dvm-darwin-arm64

# Verify it's a valid binary
file dvm-darwin-arm64
# Should show: Mach-O 64-bit executable arm64

# Test
./dvm-darwin-arm64 version
```

---

**Last Updated:** 2026-01-24 (v0.2.0)
