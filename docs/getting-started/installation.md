# Installation

## Requirements

- **Docker** (for dvm) - One of the following:
    - [OrbStack](https://orbstack.dev/) (Recommended for macOS)
    - [Docker Desktop](https://www.docker.com/products/docker-desktop/)
    - [Podman](https://podman.io/)
    - [Colima](https://github.com/abiosoft/colima)
- **Go 1.25+** (only if building from source)

---

## Homebrew (Recommended)

The easiest way to install DevOpsMaestro on macOS or Linux:

```bash
# Add the tap
brew tap rmkohlman/tap

# Install DevOpsMaestro (includes dvm binary)
brew install devopsmaestro

# Or install NvimOps only (no containers needed)
brew install nvimops

# Verify installation
dvm version
nvp version
```

---

## From Source

Build from source if you want the latest development version:

```bash
# Clone the repository
git clone https://github.com/rmkohlman/devopsmaestro.git
cd devopsmaestro

# Build both binaries
go build -o dvm .
go build -o nvp ./cmd/nvp/

# Install to your PATH
sudo mv dvm nvp /usr/local/bin/

# Verify installation
dvm version
nvp version
```

---

## From GitHub Releases

Download pre-built binaries from the [Releases page](https://github.com/rmkohlman/devopsmaestro/releases):

1. Download the archive for your platform (e.g., `devopsmaestro_0.7.1_darwin_arm64.tar.gz`)
2. Extract the archive
3. Move binaries to your PATH

```bash
# Example for macOS ARM64
curl -LO https://github.com/rmkohlman/devopsmaestro/releases/latest/download/devopsmaestro_0.7.1_darwin_arm64.tar.gz
tar xzf devopsmaestro_0.7.1_darwin_arm64.tar.gz
sudo mv dvm nvp /usr/local/bin/
```

---

## Shell Completion

Enable tab completion for your shell:

=== "Bash"

    ```bash
    # Add to ~/.bashrc
    eval "$(dvm completion bash)"
    eval "$(nvp completion bash)"
    ```

=== "Zsh"

    ```bash
    # Add to ~/.zshrc
    eval "$(dvm completion zsh)"
    eval "$(nvp completion zsh)"
    ```

=== "Fish"

    ```bash
    # Add to ~/.config/fish/config.fish
    dvm completion fish | source
    nvp completion fish | source
    ```

For more details, see [Shell Completion](../configuration/shell-completion.md).

---

## Verify Installation

After installation, verify everything is working:

```bash
# Check versions
dvm version
nvp version

# Check detected container platforms
dvm get platforms

# Initialize (creates database)
dvm init
nvp init
```

---

## Next Steps

- [Quick Start Guide](quickstart.md) - Get up and running in 5 minutes
- [Add an Existing App](existing-projects.md) - Add your current repos to dvm
- [Start a New App](new-projects.md) - Create a new app from scratch
