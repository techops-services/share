---
name: share
description: |
  Detect when an HTML file has been created during the conversation and offer to
  share it as a live page using the share CLI. Activate when the user asks to
  publish, host, or share HTML content, or when a complete HTML page has just
  been written to disk.
---

<objective>
Proactively offer to share HTML files that were created or written during the current
conversation. When the user accepts, upload the file using the `share` CLI and return
the live URL.
</objective>

<trigger_conditions>
Activate when ANY of these are true:
- User says "share this", "host this", "publish this", "make this live", "get me a link", or "make this accessible"
- User mentions wanting to show, send, or demo something to someone else
- Claude just wrote an .html file that looks like a complete page (contains `<!DOCTYPE` or `<html` tag)
</trigger_conditions>

<do_not_trigger>
Do NOT activate when:
- The HTML file is a template, partial, or component (no doctype/html tag, or filename suggests partial like `_header.html`)
- The file is a test fixture or inside `node_modules/`, `.git/`, `__tests__/`, or `test/` directories
- The user already declined to share this file in the current session
- The file was read but not written by Claude in this session
- The HTML content is a fragment being used as part of a larger build process
</do_not_trigger>

<process>
1. **Detect HTML file creation**: After writing an .html file, check if it looks like a complete page:
   - Contains `<!DOCTYPE` or `<html` tag
   - Is not inside an excluded directory
   - Has not already been offered for sharing

2. **Offer to share**: Ask the user once per file:
   ```
   I created an HTML file at <path>. Would you like me to share it as a live page?
   ```

3. **If user accepts**:
   a. Check that `share` is installed: `which share`
   b. If not installed, show:
      ```
      The share CLI is needed. Install with:
        brew install techops-services/tap/share
      Then run: share init
      ```
   c. Run `share <path>` via Bash
   d. Parse the URL from stdout (first line)
   e. Present: `Shared at <url> (copied to clipboard)`

4. **If user declines**: Acknowledge and do not offer again for this file.
</process>

<success_criteria>
- Only offer for complete HTML pages, not fragments or test files
- Offer exactly once per file
- If accepted, URL returned and displayed
- If share CLI missing, clear install instructions shown
- Never pushy -- one offer, then move on
</success_criteria>
