---
description: Share an HTML file as a live page via the share CLI. Returns a short URL.
argument-hint: [file.html | --clipboard]
allowed-tools: [Bash, Glob, Read]
---

<objective>
Upload an HTML file using the `share` CLI and return the live URL. Supports sharing a specific file, the most recently modified .html file, or HTML content from the system clipboard.

Arguments:
- (no args) -- find and share the most recently modified .html file in the working directory
- `path/to/file.html` -- share the specified file
- `--clipboard` -- share HTML content from the system clipboard via `pbpaste | share`
</objective>

<context>
Share CLI installed: !`which share 2>/dev/null || echo "NOT_INSTALLED"`
</context>

<process>
1. **Check if share CLI is installed** by examining the context above
   - If `NOT_INSTALLED`, print install instructions and stop:
     ```
     The share CLI is not installed. Install it with:
       brew install techops-services/tap/share
     or:
       go install github.com/techops-services/share/cmd/share@latest
     Then run: share init
     ```

2. **Determine the target** based on $ARGUMENTS:
   - If `--clipboard`: run `pbpaste | share` via Bash
   - If a file path is provided: verify the file exists with Bash, then run `share <path>`
   - If no arguments: use Glob to find `**/*.html` files in the working directory, pick the most recently modified one (first result from Glob, which sorts by modification time). Exclude files inside `node_modules/`, `.git/`, `dist/`, `build/`, and `__tests__/` directories. If no .html files found, tell the user.

3. **Run the share command** via Bash:
   ```
   share <file>
   ```
   The CLI prints the URL to stdout and copies it to the clipboard automatically.

4. **Parse and present the result**:
   - The first line of stdout is the URL
   - Present: `Shared at <url> (copied to clipboard)`
   - If the command fails, show the error message and suggest running `share init` if it looks like a config issue
</process>

<success_criteria>
- URL returned and displayed to user
- If share CLI is missing, clear install instructions shown
- If no HTML file found, user informed with helpful context
</success_criteria>
