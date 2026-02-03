# Release Process

This document describes the complete workflow for creating DevOpsMaestro releases.

## Overview

This repository produces **two binaries** from a single release:

| Binary | Tool | CGO Required | GoReleaser Build |
|--------|------|--------------|------------------|
| `dvm` | DevOpsMaestro | Yes (SQLite) | Cross-compiled (4 platforms) |
| `nvp` | NvimOps | No | Cross-compiled (4 platforms) |

**Note:** `dvm` requires CGO for SQLite, but GoReleaser can cross-compile it. Users may need to build locally on macOS due to CGO signing issues.

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

## 🤖 Automated Releases with GoReleaser

**GoReleaser is configured and active via GitHub Actions!**

### How It Works

When you push a tag matching `v*`, GitHub Actions automatically:

1. **Triggers** the Release workflow (`.github/workflows/release.yml`)
2. **Runs GoReleaser** which handles everything else

### What GitHub Actions Does

```
┌─────────────────────────────────────────────────────────────────┐
│                    GitHub Actions Workflow                       │
│                    (.github/workflows/release.yml)               │
├─────────────────────────────────────────────────────────────────┤
│  1. Checkout repository (full history for changelog)            │
│  2. Setup Go 1.23                                               │
│  3. Run GoReleaser with GITHUB_TOKEN                            │
└─────────────────────────────────────────────────────────────────┘
```

### What GoReleaser Does

GoReleaser (`.goreleaser.yaml`) performs these steps:

| Step | Description |
|------|-------------|
| **1. go mod tidy** | Ensures dependencies are clean |
| **2. Build nvp** | Cross-compiles for 4 platforms (CGO_ENABLED=0) |
| **3. Inject version** | Embeds Version, BuildTime, Commit via ldflags |
| **4. Shell completions** | Generates bash/zsh/fish for linux_amd64, bundles in all archives |
| **5. Create archives** | Produces `nvp_X.Y.Z_OS_ARCH.tar.gz` with binary + docs + completions |
| **6. Generate checksums** | Creates `checksums.txt` with SHA256 hashes |
| **7. Build changelog** | Extracts commits since last tag, groups by type |
| **8. Create GitHub Release** | Uploads assets, sets release notes |

### What GoReleaser Does NOT Do

| Item | Status | Reason |
|------|--------|--------|
| **Build dvm** | ❌ Manual | Requires CGO (SQLite) - can't cross-compile from Linux |
| **Update Homebrew tap** | ❌ Manual | `skip_upload: true` in config - we maintain it manually |
| **Announce release** | ❌ Disabled | `announce.skip: true` in config |
| **Sign artifacts** | ❌ Disabled | Requires GPG key setup |

### Build Targets

GoReleaser builds `nvp` only (not `dvm`):

| Platform | Architecture | File |
|----------|--------------|------|
| macOS | Apple Silicon (M1/M2/M3) | `nvp_X.Y.Z_darwin_arm64.tar.gz` |
| macOS | Intel | `nvp_X.Y.Z_darwin_amd64.tar.gz` |
| Linux | x86_64 | `nvp_X.Y.Z_linux_amd64.tar.gz` |
| Linux | ARM64 | `nvp_X.Y.Z_linux_arm64.tar.gz` |

### Version Injection

GoReleaser injects these values at build time:

```go
// In cmd/nvp/main.go
var (
    Version   = "dev"      // → Set to tag (e.g., "0.5.1")
    BuildTime = "unknown"  // → Set to build timestamp
    Commit    = "unknown"  // → Set to short commit hash
)
```

### Archive Contents

Each `.tar.gz` archive contains:

```
nvp_0.5.1_darwin_arm64/
├── nvp                    # Binary
├── README.md              # Main README
├── LICENSE                # GPL-3.0
├── NVIMOPS_TEST_PLAN.md   # Test documentation
└── completions/           # Shell completions
    ├── nvp.bash           # Bash completion
    ├── _nvp               # Zsh completion
    └── nvp.fish           # Fish completion
```

### Changelog Generation

GoReleaser auto-generates changelog from commits:

| Prefix | Category | Example |
|--------|----------|---------|
| `feat:` | 🚀 Features | `feat: Add theme system` |
| `fix:` | 🐛 Bug Fixes | `fix: Socket validation` |
| `docs:` | 📝 Documentation | (excluded from changelog) |
| `test:` | - | (excluded from changelog) |
| `chore:` | - | (excluded from changelog) |

### Configuration Files

| File | Purpose |
|------|---------|
| `.github/workflows/release.yml` | GitHub Actions workflow - triggers on `v*` tags |
| `.goreleaser.yaml` | GoReleaser config - defines builds, archives, release |

---

## 📁 GitHub Actions Workflows

### Current Workflows

We have **1 workflow** configured:

#### Release Workflow (`.github/workflows/release.yml`)

```yaml
name: Release
on:
  push:
    tags:
      - 'v*'        # Triggers on any tag starting with 'v'
```

| Setting | Value | Description |
|---------|-------|-------------|
| **Trigger** | `push.tags: v*` | Runs when a `v*` tag is pushed |
| **Runner** | `ubuntu-latest` | Uses GitHub's Linux runner |
| **Go Version** | `1.23` | Set via `actions/setup-go@v5` |
| **Permissions** | `contents: write` | Allows creating releases |

**Steps:**
1. **Checkout** - Full clone with `fetch-depth: 0` (needed for changelog)
2. **Setup Go** - Installs Go 1.23
3. **Run GoReleaser** - Executes `goreleaser release --clean`

**Environment Variables:**
| Variable | Source | Purpose |
|----------|--------|---------|
| `GITHUB_TOKEN` | Auto-provided | Create releases, upload assets |
| `HOMEBREW_TAP_GITHUB_TOKEN` | Secret (optional) | Update Homebrew tap (currently unused) |

### Workflows We DON'T Have (Potential Additions)

| Workflow | Trigger | Purpose | Priority |
|----------|---------|---------|----------|
| **CI/Test** | `push`, `pull_request` | Run `go test ./...` on every commit | High |
| **Lint** | `push`, `pull_request` | Run `golangci-lint` | Medium |
| **Build Check** | `push`, `pull_request` | Verify code compiles | Medium |
| **CodeQL** | `schedule`, `push` | Security scanning | Low |
| **Dependabot** | `schedule` | Dependency updates | Low |

### Recommended: Add CI Workflow

Create `.github/workflows/ci.yml`:

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      
      - name: Run tests
        run: go test -v ./pkg/nvimops/... ./cmd/nvp/...
      
      - name: Build nvp
        run: go build -o nvp ./cmd/nvp/
      
      - name: Verify binary
        run: ./nvp version
```

This would run tests automatically on every push and PR.

---

## 🚀 Release Steps (Current Process)

### Step 1: Prepare Repository

1. Update CHANGELOG.md with new version entry
2. Update version references in docs (README, INSTALL, etc.)
3. Ensure all tests pass: `go test ./...`
4. Commit changes: `git commit -am "docs: Update CHANGELOG for vX.Y.Z release"`
5. Push: `git push`

### Step 2: Create and Push Tag

```bash
git tag vX.Y.Z
git push origin vX.Y.Z
```

### Step 3: Wait for GitHub Actions

Monitor at: https://github.com/rmkohlman/devopsmaestro/actions

GitHub Actions will automatically:
- Build nvp for all 4 platforms
- Generate shell completions
- Create GitHub release with all artifacts
- Generate changelog from commits

Typical build time: ~1-2 minutes

### Step 4: Update Homebrew Tap (Manual)

After release is published, update the Homebrew formula:

```bash
# Get new checksums
gh release download vX.Y.Z --pattern checksums.txt --output -

# Update homebrew-tap
cd ~/Developer/tools/homebrew-tap
# Edit Formula/nvimops.rb with new version and SHA256 values
git add . && git commit -m "nvimops X.Y.Z" && git push
```

### Step 5: Verify

```bash
# Test Homebrew
brew update && brew upgrade rmkohlman/tap/nvimops
nvp version  # Should show new version

# Test direct download
gh release download vX.Y.Z --pattern "nvp_*_darwin_arm64.tar.gz"
tar xzf nvp_*.tar.gz
./nvp version
```

---

## 🔧 Manual GoReleaser Commands

For testing locally (not usually needed):

```bash
# Validate configuration
goreleaser check

# Dry run (builds but doesn't publish)
goreleaser release --snapshot --skip=publish --clean

# Full release (usually done by CI)
goreleaser release --clean
```

---

## 🤖 Future Automation Opportunities

### Auto-update Homebrew Tap

Could enable by:
1. Setting `skip_upload: false` in `.goreleaser.yaml`
2. Ensuring `HOMEBREW_TAP_GITHUB_TOKEN` secret has write access to homebrew-tap repo

### Build dvm in CI

Would require:
1. Using a macOS runner (expensive)
2. Or setting up cross-compilation toolchain for CGO
3. Or using Docker with CGO cross-compile tools

### Artifact Signing

Could enable by:
1. Setting up GPG key
2. Adding key to GitHub Secrets
3. Uncommenting `signs:` section in `.goreleaser.yaml`

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

**Last Updated:** 2026-02-02 (v0.5.1)
