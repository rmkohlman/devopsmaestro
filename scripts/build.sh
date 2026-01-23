#!/bin/bash
set -e

# DevOpsMaestro Build Script
# Builds the dvm binary and optionally installs it

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BINARY_NAME="dvm"

echo "=== Building DevOpsMaestro (dvm) ==="
echo

# Navigate to project directory
cd "$PROJECT_DIR"

# Download dependencies
echo "Downloading Go dependencies..."
go mod download
go mod tidy
echo "✓ Dependencies downloaded"
echo

# Run tests (optional, comment out for faster builds)
# echo "Running tests..."
# go test ./... -v
# echo "✓ Tests passed"
# echo

# Build the binary
echo "Building binary..."
go build -o "$BINARY_NAME" -ldflags="-s -w" main.go
echo "✓ Binary built: $PROJECT_DIR/$BINARY_NAME"
echo

# Check if we should install
if [ "$1" == "--install" ] || [ "$1" == "-i" ]; then
    echo "Installing to /usr/local/bin..."
    sudo mv "$BINARY_NAME" /usr/local/bin/
    echo "✓ Installed: /usr/local/bin/$BINARY_NAME"
    echo
    echo "Verify installation:"
    echo "  dvm --help"
else
    echo "Binary is ready at: $PROJECT_DIR/$BINARY_NAME"
    echo
    echo "To install system-wide, run:"
    echo "  sudo mv $BINARY_NAME /usr/local/bin/"
    echo "Or run with --install flag:"
    echo "  ./scripts/build.sh --install"
fi

echo
echo "=== Build Complete ==="
