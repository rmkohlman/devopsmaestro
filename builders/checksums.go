package builders

// Pinned tool versions and SHA256 checksums for builder stage downloads.
//
// Every binary downloaded during image build is verified against a known checksum
// to prevent supply-chain attacks. Checksums are stored as Go constants so that
// a version bump is a single, auditable code change.
//
// To update: bump the version constant, download the new release assets, compute
// sha256sum for each arch, and update the corresponding checksum constant.

// --- Neovim ---
// https://github.com/neovim/neovim/releases
const neovimVersion = "0.11.6"
const neovimChecksumArm64 = "8ddc0c101846145e830b17bbca50782ca9307eee4fab539d9e2ddaf8793c06f1"
const neovimChecksumX86_64 = "2fc90b962327f73a78afbfb8203fd19db8db9cdf4ee5e2bef84704339add89cc"

// --- Lazygit ---
// https://github.com/jesseduffield/lazygit/releases
const lazygitVersion = "0.60.0"
const lazygitChecksumArm64 = "2c699579165416eede4d2cfaf7d76ccd8f3b20f20f2e8b4abff6b5a6350fcdd7"
const lazygitChecksumX86_64 = "6252ca6cf98bc4fd3e0d927b54225910cfa57b065d0ad88263f14592f7f9ab15"

// --- Starship ---
// https://github.com/starship/starship/releases
// Uses musl-linked static binaries (works on both Alpine and Debian)
const starshipVersion = "1.24.2"
const starshipChecksumArm64 = "56b9ff412bbf374d29b99e5ac09a849124cb37a0a13121e8470df32de53c1ea6"
const starshipChecksumX86_64 = "00ff3c1f8ffb59b5c15d4b44c076bcca04d92cf0055c86b916248c14f3ae714a"

// --- Tree-sitter CLI ---
// Built from source via Cargo to avoid GLIBC version mismatches (see #334).
// Only the version is needed; no binary checksums since we compile from crates.io.
const treeSitterVersion = "0.24.7"

// --- golangci-lint ---
// https://github.com/golangci/golangci-lint/releases
const golangciLintVersion = "2.11.3"
const golangciLintChecksumAmd64 = "87bb8cddbcc825d5778b64e8a91b46c0526b247f4e2f2904dea74ec7450475d1"
const golangciLintChecksumArm64 = "ee3d95f301359e7d578e6d99c8ad5aeadbabc5a13009a30b2b0df11c8058afe9"

// --- opencode ---
// https://github.com/anomalyco/opencode/releases
// Uses musl-linked static binaries (works on both Alpine and Debian)
const opencodeVersion = "1.2.27"
const opencodeChecksumArm64 = "7da2618b210f9e29b746e6b863716d9d77d3484a343846b16828686babdf1dd1"
const opencodeChecksumAmd64 = "660f7319f748a66bda1748c1e7ae74dade1ba3837e6c181263506d88e7b5a4b6"
