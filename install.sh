#!/usr/bin/env bash
set -euo pipefail

REPO="techops-services/share"
INSTALL_DIR="/usr/local/bin"

die() { echo "Error: $*" >&2; exit 1; }

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  arm64)   ARCH="arm64" ;;
  *)       die "Unsupported architecture: $ARCH" ;;
esac

case "$OS" in
  darwin|linux) ;;
  *)            die "Unsupported OS: $OS" ;;
esac

VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
[ -n "$VERSION" ] || die "Could not determine latest version"

URL="https://github.com/${REPO}/releases/download/${VERSION}/share_${VERSION#v}_${OS}_${ARCH}.tar.gz"

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

echo "Installing share ${VERSION} (${OS}/${ARCH})..."
curl -fsSL "$URL" -o "$TMP/share.tar.gz"
tar xzf "$TMP/share.tar.gz" -C "$TMP"

if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP/share" "$INSTALL_DIR/share"
else
  sudo mv "$TMP/share" "$INSTALL_DIR/share"
fi

echo "Installed share to ${INSTALL_DIR}/share"
echo "Run 'share init' to configure."
