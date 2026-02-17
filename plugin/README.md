# share -- Claude Code Plugin

Claude Code plugin for the `share` CLI. Adds a `/share` slash command and an auto-share skill that detects when HTML files are created during a conversation.

## Prerequisites

The `share` CLI must be installed and configured:

```bash
brew install techops-services/tap/share
share init
```

Or install from source:

```bash
go install github.com/techops-services/share/cmd/share@latest
share init
```

## Installation

Run the install script from the plugin directory:

```bash
bash plugin/install.sh
```

This copies the command and skill files into `~/.claude/commands/` and `~/.claude/skills/share/`.

## Usage

### /share command

```
/share                  # Share the most recently modified .html file
/share page.html        # Share a specific file
/share --clipboard      # Share HTML from system clipboard
```

The command runs `share <file>`, prints the live URL, and copies it to the clipboard.

### Auto-share skill

When Claude writes a complete HTML file (one with a doctype or html tag) during a conversation, it will offer to share it:

> I created an HTML file at ./output.html. Would you like me to share it as a live page?

The offer is made once per file and only for complete pages -- not templates, test fixtures, or fragments.

### When it does not activate

- HTML fragments or partials without a doctype
- Files inside node_modules, .git, test directories
- Files that were read but not written in the session
- After the user has already declined for a given file

## Configuration

The `share` CLI reads its config from `~/.config/share/config.toml`. Run `share init` for the interactive setup wizard. Key options:

- **endpoint** -- the Cloudflare Worker URL
- **api_key** -- optional, for higher rate limits and share management
- **default_ttl** -- default expiry (e.g. `24h`, `7d`, `30d`)
- **clipboard** -- whether to auto-copy URLs (default: true)

## Troubleshooting

**"share: command not found"** -- Install the CLI with `brew install techops-services/tap/share` or check your PATH.

**Upload fails with "config" error** -- Run `share init` to set up the endpoint URL.

**Rate limit exceeded** -- Add an API key via `share init` or pass `--api-key` for higher limits.

**File too large** -- The upload limit is 5 MB. Minify HTML or remove inline assets.
