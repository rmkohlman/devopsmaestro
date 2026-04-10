# Release Process

This document describes the workflow for creating DevOpsMaestro releases.

## Version Numbering

DevOpsMaestro follows [Semantic Versioning](https://semver.org/):

- **MAJOR** (X.0.0) — Incompatible or breaking changes
- **MINOR** (0.X.0) — New features, backward compatible
- **PATCH** (0.0.X) — Bug fixes, backward compatible

**Examples:**
- `v0.2.0` — Added theme system (new feature)
- `v0.2.1` — Fixed theme rendering bug
- `v1.0.0` — First stable release

---

## Automated Releases

Releases are automated via GitHub Actions and GoReleaser. When a version tag is pushed, the release workflow:

1. Builds `nvp` binaries for all supported platforms
2. Generates shell completions (bash, zsh, fish)
3. Creates archives with binaries, docs, and completions
4. Generates SHA256 checksums
5. Publishes a GitHub Release with all assets

### Supported Platforms

| Platform | Architecture |
|----------|--------------|
| macOS | Apple Silicon (M1/M2/M3) |
| macOS | Intel |
| Linux | x86_64 |
| Linux | ARM64 |

> **Note:** `dvm` requires CGO (SQLite) and must be built manually on macOS. `nvp` is fully cross-compiled.

---

## Release Steps

### 1. Prepare

- Update `CHANGELOG.md` with the new version entry and release date
- Update installation URLs in `README.md` to point to the new version
- Commit and push: `git commit -m "chore: prepare vX.Y.Z release"`

### 2. Tag and Push

```bash
git tag vX.Y.Z
git push origin vX.Y.Z
```

Pushing the tag triggers the automated release workflow. Monitor progress at:
`https://github.com/rmkohlman/devopsmaestro/actions`

Typical build time: ~1–2 minutes.

### 3. Update Homebrew Tap

After the release is published, update the Homebrew formula with the new version checksums:

```bash
gh release download vX.Y.Z --pattern checksums.txt --output -
```

Edit the formula in the `homebrew-tap` repository with the new version and SHA256 values, then commit and push.

### 4. Verify

```bash
# Test Homebrew install
brew update && brew upgrade rmkohlman/tap/nvimops
nvp version  # Should show new version

# Test direct download
gh release download vX.Y.Z --pattern "nvp_*_darwin_arm64.tar.gz"
tar xzf nvp_*.tar.gz
./nvp version
```

---

## CHANGELOG Format

New version entries follow [Keep a Changelog](https://keepachangelog.com/) format:

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Added
- New features

### Changed
- Changes to existing functionality

### Fixed
- Bug fixes
```

**Categories:** Added · Changed · Deprecated · Removed · Fixed · Security

---

## Troubleshooting

### Tag Already Exists

```bash
# Delete local tag
git tag -d vX.Y.Z

# Delete remote tag
git push --delete origin vX.Y.Z

# Re-create and push
git tag vX.Y.Z
git push origin vX.Y.Z
```

### Binary Won't Run After Download

```bash
chmod +x dvm-darwin-arm64
./dvm-darwin-arm64 version
```

---

**Last Updated:** 2026-04-09 (v0.57.1)
