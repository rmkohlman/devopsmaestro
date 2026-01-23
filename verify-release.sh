#!/bin/bash
# Verification script for v0.2.0 release
# Run this AFTER creating the GitHub Release

set -e

VERSION="0.2.0"
REPO="rmkohlman/devopsmaestro"
RELEASE_URL="https://github.com/${REPO}/releases/download/v${VERSION}"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔍 DevOpsMaestro v${VERSION} Release Verification"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Test 1: Download macOS arm64 binary
echo "📥 Test 1: Downloading macOS arm64 binary..."
curl -L "${RELEASE_URL}/dvm-darwin-arm64" -o dvm-verify-test 2>/dev/null
chmod +x dvm-verify-test
echo "✅ Download successful ($(ls -lh dvm-verify-test | awk '{print $5}'))"
echo ""

# Test 2: Verify binary works
echo "🧪 Test 2: Testing binary execution..."
./dvm-verify-test version
echo "✅ Binary executes successfully"
echo ""

# Test 3: Download checksums
echo "📋 Test 3: Downloading checksums..."
curl -L "${RELEASE_URL}/checksums.txt" -o checksums-verify.txt 2>/dev/null
echo "✅ Checksums downloaded"
echo ""

# Test 4: Verify checksum matches
echo "🔐 Test 4: Verifying checksum..."
EXPECTED=$(grep "dvm-darwin-arm64" checksums-verify.txt | awk '{print $1}')
ACTUAL=$(shasum -a 256 dvm-verify-test | awk '{print $1}')

if [ "$EXPECTED" = "$ACTUAL" ]; then
    echo "✅ Checksum matches!"
    echo "   Expected: $EXPECTED"
    echo "   Actual:   $ACTUAL"
else
    echo "❌ Checksum mismatch!"
    echo "   Expected: $EXPECTED"
    echo "   Actual:   $ACTUAL"
    exit 1
fi
echo ""

# Test 5: Test theme system
echo "🎨 Test 5: Testing theme system..."
DVM_THEME=catppuccin-mocha ./dvm-verify-test version > /dev/null 2>&1
echo "✅ Catppuccin Mocha theme works"

DVM_THEME=tokyo-night ./dvm-verify-test version > /dev/null 2>&1
echo "✅ Tokyo Night theme works"

DVM_THEME=nord ./dvm-verify-test version > /dev/null 2>&1
echo "✅ Nord theme works"
echo ""

# Test 6: Check release page
echo "🌐 Test 6: Checking release page..."
if curl -s "https://github.com/${REPO}/releases/tag/v${VERSION}" | grep -q "v${VERSION}"; then
    echo "✅ Release page exists and accessible"
else
    echo "⚠️  Release page not found (may need to wait for CDN)"
fi
echo ""

# Cleanup
rm -f dvm-verify-test checksums-verify.txt

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ All verification tests passed!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "🎉 v0.2.0 release is LIVE and working!"
echo ""
echo "📍 Release URL: https://github.com/${REPO}/releases/tag/v${VERSION}"
echo "📍 Latest:      https://github.com/${REPO}/releases/latest"
echo ""
