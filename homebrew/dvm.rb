# Documentation: https://docs.brew.sh/Formula-Cookbook
#                https://rubydoc.brew.sh/Formula
# HOWTO: How to submit a formula to Homebrew
#   https://docs.brew.sh/How-To-Open-a-Homebrew-Pull-Request

class Dvm < Formula
  desc "DevOpsMaestro - Kubernetes-style development environment orchestration"
  homepage "https://github.com/rmkohlman/devopsmaestro"
  url "https://github.com/rmkohlman/devopsmaestro/archive/refs/tags/v0.1.0-dev.tar.gz"
  sha256 "e165863da100e52e3de0f330f3d34e38f06847141ff61ae8f7006f7f68513fc1"
  license "GPL-3.0"
  head "https://github.com/rmkohlman/devopsmaestro.git", branch: "main"

  # Uncomment when you want automatic version checking
  # livecheck do
  #   url :stable
  #   strategy :github_latest
  # end

  # Pre-built binaries for different architectures (created by CI/CD)
  # bottle do
  #   sha256 cellar: :any_skip_relocation, arm64_sequoia: "..."
  #   sha256 cellar: :any_skip_relocation, arm64_sonoma:  "..."
  #   sha256 cellar: :any_skip_relocation, sonoma:        "..."
  #   sha256 cellar: :any_skip_relocation, x86_64_linux:  "..."
  # end

  depends_on "go" => :build

  def install
    # Determine version
    dvm_version = if build.stable?
      version.to_s
    else
      Utils.safe_popen_read("git", "describe", "--tags", "--dirty").chomp
    end

    # Get commit hash (for stable releases, use a placeholder since tarball has no .git)
    commit_hash = if build.stable?
      "release-#{version}"
    else
      Utils.git_short_head
    end

    # Build flags (same as in Makefile)
    ldflags = %W[
      -s -w
      -X main.Version=#{dvm_version}
      -X main.BuildTime=#{time.iso8601}
      -X main.Commit=#{commit_hash}
    ]

    # Build the binary
    system "go", "build", *std_go_args(ldflags:, output: bin/"dvm")

    # Install shell completions (when you add them)
    # generate_completions_from_executable(bin/"dvm", "completion")
  end

  def post_install
    # Create data directory on first install
    (var/"devopsmaestro").mkpath
    ohai "DevOpsMaestro (dvm) installed successfully!"
    ohai "Run 'dvm init' to initialize your environment"
  end

  test do
    # Test that the binary runs and returns version
    assert_match version.to_s, shell_output("#{bin}/dvm version")
    
    # Test basic functionality
    system bin/"dvm", "--help"
  end
end
