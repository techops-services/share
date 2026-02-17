# share

Upload HTML and get a short URL. Pages are hosted on Cloudflare Workers and expire automatically.

```
share page.html                  # upload a file
echo "<h1>hi</h1>" | share      # pipe from stdin
pbpaste | share                  # upload clipboard contents
share ./dist                     # upload index.html from a directory
```

Output is a single URL, ready to pipe or paste: `https://share.techops.services/V1StGXR8`

## Install

```bash
# Homebrew
brew install techops-services/tap/share

# Go
go install github.com/techops-services/share/cmd/share@latest
```

Then run `share init` to set up your config at `~/.config/share/config.toml`.

## Commands

| Command | Description |
|---|---|
| `share <file>` | Upload an HTML file |
| `share <dir>` | Upload `index.html` from a directory |
| `share list` | List your shared pages (requires API key) |
| `share delete <id>` | Delete a shared page (requires API key) |
| `share init` | Interactive config setup |
| `share version` | Print version info |

## Options

```
--ttl <duration>    Expiry time (e.g. 1h, 7d, 30d). Default: 24h
--api-key <key>     API key for higher limits and management
--endpoint <url>    Custom API endpoint
--no-clipboard      Don't copy URL to clipboard
--verbose           Print full response details
--config <path>     Custom config file path
```

Anonymous uploads expire in 24h max. Authenticated uploads (with `--api-key`) can last up to 30 days.

## Claude Code Plugin

The share tool includes a Claude Code plugin with a `/share` slash command and auto-detection skill.

### Install the plugin

```bash
bash plugin/install.sh
```

This installs:
- `/share` command -- share HTML files directly from Claude Code
- Auto-share skill -- Claude offers to share HTML pages it creates during conversations

### Usage in Claude Code

```
/share                    # share the most recently created .html file
/share page.html          # share a specific file
/share --clipboard        # share HTML from clipboard
```

Claude will also detect when it writes a complete HTML page and offer to share it automatically.

## Architecture

- **CLI** (`cmd/` + `internal/`) -- Go, Cobra, TOML config
- **Worker** (`worker/`) -- Cloudflare Worker + R2 + KV, Hono router
- **Plugin** (`plugin/`) -- Claude Code slash command + skill

## Development

```bash
# Run Go tests
go test ./...

# Run worker locally
cd worker && npm install && npm run dev

# Run worker tests
cd worker && npm test

# Deploy worker
cd worker && npx wrangler deploy
```

## Roadmap

- **Directory uploads** -- `share ./dist` bundles all files (HTML, images, CSS, JS) and uploads them under a shared prefix. Relative links between pages and assets work as-is. Served as `/{prefix}/{path}` with `index.html` as the default.

## License

MIT
