#!/usr/bin/env bash
set -euo pipefail

# Install the share CLI plugin for Claude Code.
# Usage: curl -fsSL <raw-url>/install.sh | bash

REPO="techops-services/share"
PLUGIN_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors (suppressed if not a terminal).
if [ -t 1 ]; then
  BOLD="\033[1m"
  DIM="\033[2m"
  RESET="\033[0m"
else
  BOLD=""
  DIM=""
  RESET=""
fi

info()  { echo -e "${BOLD}$*${RESET}"; }
dim()   { echo -e "${DIM}$*${RESET}"; }

# --- 1. Install share CLI if missing ---

if command -v share &>/dev/null; then
  dim "share CLI already installed: $(which share)"
else
  info "Installing share CLI..."
  if command -v brew &>/dev/null; then
    brew install techops-services/tap/share
  elif [[ "$(uname)" == "Linux" ]]; then
    ARCH="$(uname -m)"
    case "$ARCH" in
      x86_64)  ARCH="amd64" ;;
      aarch64) ARCH="arm64" ;;
    esac
    LATEST=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep tag_name | cut -d '"' -f 4)
    URL="https://github.com/${REPO}/releases/download/${LATEST}/share_${LATEST#v}_linux_${ARCH}.tar.gz"
    TMP=$(mktemp -d)
    curl -fsSL "$URL" | tar xz -C "$TMP"
    sudo mv "$TMP/share" /usr/local/bin/share
    rm -rf "$TMP"
    info "Installed share ${LATEST} to /usr/local/bin/share"
  else
    echo "Could not auto-install. Install manually:"
    echo "  brew install techops-services/tap/share"
    echo "  go install github.com/${REPO}/cmd/share@latest"
    exit 1
  fi
fi

# --- 2. Install Claude Code plugin files ---

CLAUDE_DIR="${HOME}/.claude"
CMD_DIR="${CLAUDE_DIR}/commands"
SKILL_DIR="${CLAUDE_DIR}/skills/share"

mkdir -p "$CMD_DIR" "$SKILL_DIR"

cp "${PLUGIN_DIR}/commands/share.md" "$CMD_DIR/share.md"
cp "${PLUGIN_DIR}/skills/share/SKILL.md" "$SKILL_DIR/SKILL.md"

info "Plugin installed."
echo ""
echo "  /share command  -> ${CMD_DIR}/share.md"
echo "  auto-share skill -> ${SKILL_DIR}/SKILL.md"
echo ""
dim "Usage:"
echo "  /share                  Share the most recent .html file"
echo "  /share page.html        Share a specific file"
echo "  /share --clipboard      Share HTML from clipboard"
echo ""
dim "If you haven't configured share yet, run: share init"
