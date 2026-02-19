# Installation

## Requirements

- **Container Runtime** (for dvm workspaces) - One of the following:
    - [OrbStack](https://orbstack.dev/) (⭐ Recommended for macOS)
    - [Docker Desktop](https://www.docker.com/products/docker-desktop/)
    - [Podman](https://podman.io/)
    - [Colima](https://github.com/abiosoft/colima)
- **Go 1.25+** (only if building from source)

!!! note "NvimOps works without containers"
    If you only want to use `nvp` (Neovim plugin manager), you don't need Docker. Install just `nvimops` via Homebrew.

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
dvm version  # Should show v0.12.0+
nvp version  # Should show v0.12.0+
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

### macOS/Linux (Quick Install)

```bash
# Download latest release (replace with actual version)
curl -LO https://github.com/rmkohlman/devopsmaestro/releases/latest/download/devopsmaestro_0.12.0_$(uname -s | tr '[:upper:]' '[:lower:]')_$(uname -m | sed 's/x86_64/amd64/').tar.gz

# Extract and install
tar xzf devopsmaestro_*.tar.gz
sudo mv dvm nvp /usr/local/bin/

# Clean up
rm devopsmaestro_*.tar.gz

# Verify
dvm version
nvp version
```

### Manual Download

1. Go to [Releases page](https://github.com/rmkohlman/devopsmaestro/releases/latest)
2. Download the archive for your platform:
   - `devopsmaestro_0.12.0_darwin_amd64.tar.gz` (macOS Intel)
   - `devopsmaestro_0.12.0_darwin_arm64.tar.gz` (macOS Apple Silicon)
   - `devopsmaestro_0.12.0_linux_amd64.tar.gz` (Linux x64)
   - `devopsmaestro_0.12.0_linux_arm64.tar.gz` (Linux ARM64)
3. Extract and move binaries to your PATH

```bash
# Example for macOS Apple Silicon
tar xzf devopsmaestro_0.12.0_darwin_arm64.tar.gz
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
dvm version  # Should show v0.12.0+
nvp version  # Should show v0.12.0+

# Check detected container platforms
dvm get platforms

# Expected output:
# PLATFORM    STATUS    VERSION        CONTEXT
# orbstack    ready     v1.5.1         orbstack-vm  
# docker      ready     v25.0.3        default
# colima      stopped   v0.6.8         colima

# Initialize (creates database and config)
dvm init
nvp init
```

!!! success "Installation Complete!"
    If all commands run without errors, you're ready to start using DevOpsMaestro!

### Troubleshooting

#### Container Platform Issues

If `dvm get platforms` shows all platforms as "unavailable":

```bash
# Check if Docker/OrbStack is running
docker ps
# or
orbctl status

# Start your preferred platform
# OrbStack: Open OrbStack app
# Docker Desktop: Open Docker Desktop app  
# Colima: colima start
```

#### Permission Issues (Linux/macOS)

If you get permission errors:

```bash
# Option 1: Install to user directory
mkdir -p ~/bin
mv dvm nvp ~/bin/
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# Option 2: Use sudo for /usr/local/bin (shown above)
sudo mv dvm nvp /usr/local/bin/
```

#### Binary Not Found

If `dvm: command not found`:

```bash
# Check if binary is in PATH
which dvm
echo $PATH

# Add to PATH if needed (replace with your install location)
echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

---

## Next Steps

- [Quick Start Guide](quickstart.md) - Get up and running in 5 minutes
- [Add an Existing App](existing-projects.md) - Add your current repos to dvm
- [Start a New App](new-projects.md) - Create a new app from scratch
