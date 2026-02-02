# Homebrew Distribution Guide

This guide explains how to distribute DevOpsMaestro tools via Homebrew.

## Current Status ✅

**Our Homebrew tap is live!** Both tools are available:

```bash
# Add the tap
brew tap rmkohlman/tap

# Install NvimOps (standalone Neovim plugin/theme manager)
brew install rmkohlman/tap/nvimops
nvp version  # Should show v0.5.0

# Install DevOpsMaestro (workspace/container management)
# Note: dvm requires CGO and must be built locally
# See "Building dvm Locally" below
```

**Tap repository:** https://github.com/rmkohlman/homebrew-tap

---

## Available Formulas

| Formula | Binary | Description | CGO Required |
|---------|--------|-------------|--------------|
| `nvimops` | `nvp` | Neovim plugin & theme manager | No |
| `devopsmaestro` | `dvm` | Workspace/container management | Yes (SQLite) |

### Installing NvimOps (nvp)

```bash
brew tap rmkohlman/tap
brew install rmkohlman/tap/nvimops

# Verify
nvp version
nvp --help
```

### Building dvm Locally

dvm requires CGO for SQLite and cannot be distributed as a pre-built Homebrew bottle on macOS. Build from source:

```bash
git clone https://github.com/rmkohlman/devopsmaestro.git
cd devopsmaestro
go build -o dvm .
sudo mv dvm /usr/local/bin/
```

---

## Overview

Homebrew installation works through **formulas** - Ruby files that tell Homebrew how to:
- Download your source code
- Build the binary
- Install it to the right location
- Handle dependencies

## Two Distribution Options

### Option 1: Personal Homebrew Tap ✅ (Current)

**Pros:**
- Full control over releases
- Easier to publish (no approval needed)
- Good for beta/testing
- Can iterate quickly

**Cons:**
- Less discoverability
- Users need to add your tap first

**Setup:**

1. **Create a tap repository on GitHub:**
   ```bash
   # Repository name MUST be: homebrew-tap (or homebrew-<name>)
   # URL: https://github.com/yourusername/homebrew-tap
   ```

2. **Add the formula:**
   ```bash
   # Copy homebrew/dvm.rb to your tap repo
   cp homebrew/dvm.rb ~/homebrew-tap/Formula/dvm.rb
   git add Formula/dvm.rb
   git commit -m "Add dvm formula"
   git push
   ```

3. **Users install with:**
   ```bash
   brew tap yourusername/tap
   brew install dvm
   ```

### Option 2: Official Homebrew Core

**Pros:**
- Maximum discoverability
- Users just run `brew install dvm`
- Automatic updates via Homebrew
- Vetted by Homebrew maintainers

**Cons:**
- Strict requirements (must be "notable" software)
- Review process can take time
- Must follow all Homebrew best practices
- Need stable releases first

**Requirements:**
1. 75+ GitHub stars (shows community interest)
2. Stable versioned releases
3. Good documentation
4. CI/CD with proper builds
5. No security issues

## Step-by-Step: Publishing Your Formula

### Prerequisites

1. **Create Versioned Releases**
   ```bash
   # Tag a version
   git tag -a v0.1.0 -m "Release v0.1.0"
   git push origin v0.1.0
   
   # GitHub will create a tarball at:
   # https://github.com/yourusername/devopsmaestro/archive/refs/tags/v0.1.0.tar.gz
   ```

2. **Calculate SHA256 Checksum**
   ```bash
   # Download the release tarball
   curl -L -o v0.1.0.tar.gz https://github.com/yourusername/devopsmaestro/archive/refs/tags/v0.1.0.tar.gz
   
   # Calculate checksum
   shasum -a 256 v0.1.0.tar.gz
   # Output: abc123def456... v0.1.0.tar.gz
   ```

3. **Update Formula**
   ```ruby
   class Dvm < Formula
     desc "DevOpsMaestro - Kubernetes-style development environment orchestration"
     homepage "https://github.com/yourusername/devopsmaestro"
     url "https://github.com/yourusername/devopsmaestro/archive/refs/tags/v0.1.0.tar.gz"
     sha256 "abc123def456..." # Your actual checksum here
     license "MIT"
     # ... rest of formula
   end
   ```

### Publishing to Your Own Tap

```bash
# 1. Create tap repository
mkdir homebrew-tap
cd homebrew-tap
git init
mkdir Formula

# 2. Copy formula
cp /path/to/devopsmaestro/homebrew/dvm.rb Formula/

# 3. Push to GitHub
git add .
git commit -m "Initial formula for dvm"
git remote add origin git@github.com:yourusername/homebrew-tap.git
git push -u origin main
```

**Users then install:**
```bash
brew tap yourusername/tap
brew install dvm
```

### Submitting to Homebrew Core

1. **Test Your Formula Locally**
   ```bash
   brew install --build-from-source homebrew/dvm.rb
   brew test dvm
   brew audit --strict dvm
   ```

2. **Fork homebrew-core**
   ```bash
   # Go to https://github.com/Homebrew/homebrew-core
   # Click "Fork"
   ```

3. **Add Your Formula**
   ```bash
   # Clone your fork
   git clone git@github.com:yourusername/homebrew-core.git
   cd homebrew-core
   
   # Create formula (first letter directory)
   cp /path/to/dvm.rb Formula/d/dvm.rb
   
   # Create branch
   git checkout -b dvm
   git add Formula/d/dvm.rb
   git commit -m "dvm 0.1.0 (new formula)"
   git push origin dvm
   ```

4. **Submit Pull Request**
   - Go to https://github.com/Homebrew/homebrew-core
   - Click "New Pull Request"
   - Follow the PR template

## Automation with GitHub Actions

Create `.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Build Release Binaries
        run: make release
      
      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: dist/*
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      
      - name: Update Homebrew Formula
        # Auto-update your tap formula with new version
        # See: https://github.com/dawidd6/action-homebrew-bump-formula
        uses: dawidd6/action-homebrew-bump-formula@v3
        with:
          tap: yourusername/homebrew-tap
          formula: dvm
          token: ${{ secrets.HOMEBREW_TAP_TOKEN }}
```

## Testing Your Formula

Before publishing:

```bash
# Install from local file
brew install --build-from-source homebrew/dvm.rb

# Test it
dvm version

# Run formula tests
brew test dvm

# Audit for issues
brew audit --strict dvm

# Uninstall
brew uninstall dvm
```

## Updating Your Formula

When you release a new version:

```bash
# 1. Tag new version
git tag v0.2.0
git push origin v0.2.0

# 2. Calculate new checksum
curl -L -o v0.2.0.tar.gz https://github.com/yourusername/devopsmaestro/archive/refs/tags/v0.2.0.tar.gz
shasum -a 256 v0.2.0.tar.gz

# 3. Update formula (in your tap)
# Change url and sha256
vim Formula/dvm.rb

# 4. Test
brew reinstall dvm --build-from-source

# 5. Commit
git add Formula/dvm.rb
git commit -m "dvm 0.2.0"
git push
```

Users then upgrade with:
```bash
brew update
brew upgrade dvm
```

## Complete Homebrew Experience

Once published, users get:

```bash
# Installation
brew install dvm                    # Install latest version
brew install dvm@0.1.0             # Install specific version

# Management  
brew upgrade dvm                   # Upgrade to latest
brew uninstall dvm                 # Remove
brew reinstall dvm                 # Reinstall

# Information
brew info dvm                      # Show details
brew list --versions dvm           # Show installed versions
brew deps dvm                      # Show dependencies
brew uses --installed dvm          # What depends on dvm

# Maintenance
brew cleanup dvm                   # Remove old versions
brew pin dvm                       # Prevent upgrades
brew unpin dvm                     # Allow upgrades again
```

## Best Practices

1. **Versioning:** Use semantic versioning (v1.2.3)
2. **Changelog:** Maintain CHANGELOG.md with release notes
3. **Testing:** Always test formula before publishing
4. **Documentation:** Keep README.md up to date
5. **CI/CD:** Automate builds and releases
6. **Bottles:** Pre-compile binaries for faster installation (advanced)

## Resources

- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [Homebrew Python Guide](https://docs.brew.sh/Python-for-Formula-Authors)
- [Creating Taps](https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap)
- [Acceptable Formulae](https://docs.brew.sh/Acceptable-Formulae)
- [Formula PR Guidelines](https://docs.brew.sh/How-To-Open-a-Homebrew-Pull-Request)

## Current Status

✅ Homebrew tap live at https://github.com/rmkohlman/homebrew-tap
✅ NvimOps (nvp) formula available - v0.5.0
✅ DevOpsMaestro formula available (requires local CGO build)
✅ GitHub releases with pre-built nvp binaries (4 platforms)
✅ GoReleaser configured for automated releases

**Formulas:**
- `Formula/nvimops.rb` - NvimOps (nvp binary)
- `Formula/devopsmaestro.rb` - DevOpsMaestro (dvm binary)
