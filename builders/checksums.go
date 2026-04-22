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
// Uses pre-built tarball from neovim/neovim releases. These tarballs target GLIBC 2.17+,
// making them compatible with virtually all Linux base images (see #356).
// Previous AppImage approach (#342) failed because AppImages do NOT bundle glibc —
// the extracted nvim binary still dynamically links to the system glibc.
const neovimVersion = "0.11.6"
const neovimTarballChecksumArm64 = "8ddc0c101846145e830b17bbca50782ca9307eee4fab539d9e2ddaf8793c06f1"
const neovimTarballChecksumX86_64 = "2fc90b962327f73a78afbfb8203fd19db8db9cdf4ee5e2bef84704339add89cc"

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
// https://github.com/tree-sitter/tree-sitter/releases
// Debian: uses pre-built binary from GitHub releases (faster than cargo install).
// Alpine: still built from source via Cargo (pre-built binary requires glibc).
const treeSitterVersion = "0.24.7"
const treeSitterChecksumArm64 = "bad9cd53adcbd18df33084bb811b8cf7868fffd79437acfc83ac1025e7574c78"
const treeSitterChecksumX86_64 = "628fa0e1c4d78b5d4f7de64b6ab42fc050e3bee14cb92a076beb82d762d76d69"

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

// --- kubectl (KindCICD app builder) ---
// https://kubernetes.io/releases/
// Pinned upstream release; checksum from https://dl.k8s.io/release/vX.Y.Z/bin/linux/<arch>/kubectl.sha256
// To refresh: bump kubectlVersion, fetch the two .sha256 files for amd64/arm64.
// Compatibility: CLI must be within ±1 minor of cluster control-plane version.
const kubectlVersion = "1.31.4"
const kubectlChecksumAmd64 = "298e19e9c6c17199011404278f0ff8168a7eca4217edad9097af577023a5620f"
const kubectlChecksumArm64 = "b97e93c20e3be4b8c8fa1235a41b4d77d4f2022ed3d899230dbbbbd43d26f872"

// --- helm (KindCICD app builder) ---
// https://github.com/helm/helm/releases
// Checksums from https://get.helm.sh/helm-vX.Y.Z-linux-<arch>.tar.gz.sha256sum
const helmVersion = "3.16.3"
const helmChecksumAmd64 = "f5355c79190951eed23c5432a3b920e071f4c00a64f75e077de0dd4cb7b294ea"
const helmChecksumArm64 = "5bd34ed774df6914b323ff84a0a156ea6ff2ba1eaf0113962fa773f3f9def798"

// --- kustomize (KindCICD app builder) ---
// https://github.com/kubernetes-sigs/kustomize/releases
// Checksums from the release's checksums.txt asset.
const kustomizeVersion = "5.5.0"
const kustomizeChecksumAmd64 = "6703a3a70a0c47cf0b37694030b54f1175a9dfeb17b3818b623ed58b9dbc2a77"
const kustomizeChecksumArm64 = "b4170d1acb8cfacace9f72884bef957ff56efdcd4813b66e7604aabc8b57e93d"

// --- argocd CLI (KindCICD app builder, conditional on .argocd/) ---
// https://github.com/argoproj/argo-cd/releases
// Checksums from the release's cli_checksums.txt asset.
// Compatibility: CLI must be ≤ server +1 minor — bump cautiously.
const argocdVersion = "2.13.1"
const argocdChecksumAmd64 = "8e436f0429d2a88b3181d2cfc460c034070e0ee1c665467271e5d75eb4d55f7f"
const argocdChecksumArm64 = "76cbc9044c6c8f989302e0354516a95b485e1c9c5eba431fef6a669b2fbd3be4"
