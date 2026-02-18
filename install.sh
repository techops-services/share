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

# Add install dir to PATH if needed
case ":$PATH:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    SHELL_NAME=$(basename "${SHELL:-/bin/bash}")
    case "$SHELL_NAME" in
      zsh)  RC_FILE="$HOME/.zshrc" ;;
      bash)
        if [ -f "$HOME/.bashrc" ]; then
          RC_FILE="$HOME/.bashrc"
        else
          RC_FILE="$HOME/.bash_profile"
        fi
        ;;
      fish) RC_FILE="$HOME/.config/fish/config.fish" ;;
      *)    RC_FILE="$HOME/.profile" ;;
    esac

    EXPORT_LINE="export PATH=\"${INSTALL_DIR}:\$PATH\""
    if [ "$SHELL_NAME" = "fish" ]; then
      EXPORT_LINE="fish_add_path ${INSTALL_DIR}"
    fi

    if ! grep -qF "$INSTALL_DIR" "$RC_FILE" 2>/dev/null; then
      echo "" >> "$RC_FILE"
      echo "$EXPORT_LINE" >> "$RC_FILE"
      echo "Added ${INSTALL_DIR} to PATH in ${RC_FILE}"
    fi

    export PATH="${INSTALL_DIR}:$PATH"
    ;;
esac

echo "Run 'share init' to configure."
