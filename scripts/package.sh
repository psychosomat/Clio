#!/usr/bin/env bash
set -euo pipefail

RAW_VERSION="${1:-${GITHUB_REF_NAME:-dev}}"
VERSION="${RAW_VERSION#v}"
PACKAGE_NAME="clio"
ARCH="amd64"
PACMAN_ARCH="x86_64"
RELEASE="1"
BUILD_LDFLAGS="-s -w -X main.version=${VERSION}"

# Create build directories
BUILD_DIR="build"
DIST_DIR="dist"
BIN_PATH="$BUILD_DIR/clio"
rm -rf "$BUILD_DIR" "$DIST_DIR"
mkdir -p "$BUILD_DIR" "$DIST_DIR"

echo "=== 1. Building Clio Binary ==="
go build -a -trimpath -ldflags "$BUILD_LDFLAGS" -o "$BIN_PATH" ./cmd/clio
echo "Binary built successfully at $BIN_PATH"

echo "=== 2. Creating Debian Package (.deb) ==="
DEB_ROOT="$BUILD_DIR/deb/${PACKAGE_NAME}_${VERSION}_${ARCH}"
mkdir -p "$DEB_ROOT/DEBIAN"
mkdir -p "$DEB_ROOT/usr/bin"

# Copy binary
cp "$BIN_PATH" "$DEB_ROOT/usr/bin/clio"

# Create control file
cat <<EOF > "$DEB_ROOT/DEBIAN/control"
Package: ${PACKAGE_NAME}
Version: ${VERSION}
Section: utils
Priority: optional
Architecture: ${ARCH}
Maintainer: Clio Developer <clio@example.com>
Description: Clio is a beautiful, keyboard-first Terminal TUI notes utility written in Go.
EOF

# Build using dpkg-deb if available
if command -v dpkg-deb &> /dev/null; then
    dpkg-deb --build "$DEB_ROOT" "$DIST_DIR/${PACKAGE_NAME}_${VERSION}_${ARCH}.deb"
    echo "✓ Debian package created: dist/${PACKAGE_NAME}_${VERSION}_${ARCH}.deb"
else
    echo "⚠ dpkg-deb command not found. Skipping .deb creation."
fi

echo "=== 3. Creating Arch Linux Package (.pkg.tar.zst) ==="
PACMAN_ROOT="$BUILD_DIR/pacman"
mkdir -p "$PACMAN_ROOT/usr/bin"
cp "$BIN_PATH" "$PACMAN_ROOT/usr/bin/clio"

# Generate .PKGINFO metadata
SIZE=$(du -sb "$BIN_PATH" | cut -f1)
cat <<EOF > "$PACMAN_ROOT/.PKGINFO"
pkgname = ${PACKAGE_NAME}
pkgver = ${VERSION}-${RELEASE}
pkgdesc = Clio - Terminal Quick Notes Utility
url = https://github.com/clio/clio
builddate = $(date +%s)
packager = Clio Developer <clio@example.com>
size = ${SIZE}
arch = ${PACMAN_ARCH}
license = MIT
EOF

if command -v tar &> /dev/null && command -v zstd &> /dev/null; then
    cd "$PACMAN_ROOT"
    tar --owner=0 --group=0 -cf - .PKGINFO usr | zstd -z -q - -o "../../dist/${PACKAGE_NAME}-${VERSION}-${RELEASE}-${PACMAN_ARCH}.pkg.tar.zst"
    cd ../..
    echo "✓ Pacman package created: dist/${PACKAGE_NAME}-${VERSION}-${RELEASE}-${PACMAN_ARCH}.pkg.tar.zst"
else
    if command -v makepkg &> /dev/null; then
        mkdir -p "$BUILD_DIR/makepkg"
        mkdir -p "$BUILD_DIR/makepkg/build"
        cp "$BIN_PATH" "$BUILD_DIR/makepkg/build/clio"
        cat <<EOF > "$BUILD_DIR/makepkg/PKGBUILD"
pkgname=${PACKAGE_NAME}
pkgver=${VERSION}
pkgrel=${RELEASE}
pkgdesc="Terminal Quick Notes Utility"
arch=('x86_64')
license=('MIT')
package() {
    install -Dm755 "\${srcdir}/build/clio" "\${pkgdir}/usr/bin/clio"
}
EOF
        cd "$BUILD_DIR/makepkg"
        SRCDEST="." makepkg -f --nodeps
        cd ../..
        cp "$BUILD_DIR/makepkg"/*.pkg.tar.zst "$DIST_DIR/"
        echo "✓ Pacman package created using makepkg"
    else
        echo "⚠ Neither (tar + zstd) nor makepkg found. Skipping pacman package creation."
    fi
fi

echo "=== Packaging Complete ==="
if [ -d "$DIST_DIR" ]; then
    ls -l "$DIST_DIR"
fi
