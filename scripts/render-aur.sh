#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 3 ]; then
    echo "usage: $0 <version-without-v> <source-url> <sha256>" >&2
    exit 1
fi

VERSION="$1"
SOURCE_URL="$2"
SOURCE_SHA="$3"
PACKAGE_NAME="${PACKAGE_NAME:-clio}"
OUT_DIR="${OUT_DIR:-dist/aur}"

mkdir -p "$OUT_DIR"

cat > "${OUT_DIR}/PKGBUILD" <<EOF
pkgname=${PACKAGE_NAME}
pkgver=${VERSION}
pkgrel=1
pkgdesc="A lightning-fast, keyboard-driven TUI for taking Markdown notes in the terminal. Powered by Go & Bubble Tea."
arch=('x86_64')
url="https://github.com/${GITHUB_REPOSITORY:-clio/clio}"
license=('MIT')
depends=()
makedepends=('go')
source=("${SOURCE_URL}")
sha256sums=('${SOURCE_SHA}')

build() {
  cd "\${srcdir}/clio-\${pkgver}"
  go build -trimpath -ldflags="-s -w -X main.version=\${pkgver}" -o clio ./cmd/clio
}

package() {
  cd "\${srcdir}/clio-\${pkgver}"
  install -Dm755 clio "\${pkgdir}/usr/bin/clio"
}
EOF

cat > "${OUT_DIR}/.SRCINFO" <<EOF
pkgbase = ${PACKAGE_NAME}
	pkgdesc = A lightning-fast, keyboard-driven TUI for taking Markdown notes in the terminal. Powered by Go & Bubble Tea.
	pkgver = ${VERSION}
	pkgrel = 1
	url = https://github.com/${GITHUB_REPOSITORY:-clio/clio}
	arch = x86_64
	license = MIT
	makedepends = go
	source = ${SOURCE_URL}
	sha256sums = ${SOURCE_SHA}

pkgname = ${PACKAGE_NAME}
EOF
