#!/usr/bin/env bash
set -euo pipefail

REPO="techops-services/share"

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

# Pick a writable install directory — prefer /usr/local/bin, fall back to ~/.local/bin
if [ -d "/usr/local/bin" ] && [ -w "/usr/local/bin" ]; then
  INSTALL_DIR="/usr/local/bin"
else
  INSTALL_DIR="${HOME}/.local/bin"
  mkdir -p "$INSTALL_DIR"
fi

VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
[ -n "$VERSION" ] || die "Could not determine latest version"

URL="https://github.com/${REPO}/releases/download/${VERSION}/share_${VERSION#v}_${OS}_${ARCH}.tar.gz"

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

echo "Installing share ${VERSION} (${OS}/${ARCH})..."
curl -fsSL "$URL" -o "$TMP/share.tar.gz"
tar xzf "$TMP/share.tar.gz" -C "$TMP"
mv "$TMP/share" "$INSTALL_DIR/share"

echo "Installed share to ${INSTALL_DIR}/share"

# Warn if the install dir isn't in PATH
case ":$PATH:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    echo ""
    echo "NOTE: ${INSTALL_DIR} is not in your PATH."
    echo "Add it by running:"
    echo ""
    echo "  echo 'export PATH=\"${INSTALL_DIR}:\$PATH\"' >> ~/.zshrc && source ~/.zshrc"
    echo ""
    ;;
esac

echo "Run 'share init' to configure."
