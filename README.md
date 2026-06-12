# cc2md

[![CI](https://github.com/magarcia/cc2md/actions/workflows/ci.yml/badge.svg)](https://github.com/magarcia/cc2md/actions/workflows/ci.yml)

Convert Claude Code session logs into clean, shareable markdown.

Claude Code stores session logs as JSONL files in `~/.claude/projects/`. `cc2md` reads those logs and renders them in the terminal with [glamour](https://github.com/charmbracelet/glamour), or exports raw markdown to files. When invoked with no arguments in a TTY, it launches an interactive session picker.

## Installation

### Homebrew

```bash
brew install magarcia/tap/cc2md
```

### Go

```bash
go install github.com/magarcia/cc2md@latest
```

### Binary

Download a pre-built binary from [GitHub Releases](https://github.com/magarcia/cc2md/releases).

## Quick Start

```bash
# Open interactive session picker
cc2md

# View the most recent session
cc2md --last 1

# Export a session to a markdown file
cc2md session.jsonl --output session.md

# List all sessions
cc2md list
```

## Usage

### Commands

```
cc2md [options] [file]       View a session (interactive picker if no args)
cc2md list [project]         List available sessions, optionally filter by project
```

### Flags

**Output**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | | Write to file instead of stdout |
| `--raw` | | | Output raw markdown, skip glamour rendering |
| `--no-pager` | | | Disable pager even on TTY |

**Rendering**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--last` | `-n` | `1` | View the Nth most recent session |
| `--style` | `-s` | `auto` | Glamour style: `auto`, `dark`, `light`, `notty` |
| `--width` | `-w` | terminal width | Word wrap width |

**Formatting**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--thinking` | `-t` | | Include thinking blocks |
| `--collapse` | `-c` | `true` | Collapse tool calls into `<details>` tags |
| `--max-lines` | | `100` | Max lines per tool output before truncation |
| `--markdown` | `-m` | | Markdown flavor: `gfm`, `commonmark`, `obsidian` |

**List subcommand**

| Flag | Default | Description |
|------|---------|-------------|
| `--json` | | Output as JSON array |

### Environment Variables

| Variable | Description |
|----------|-------------|
| `$PAGER` | Pager command to use (default: `less`) |
| `$NO_COLOR` | Disable color output |

### Markdown Flavors

By default, `cc2md` uses **CommonMark** for terminal rendering and **GFM** when writing to a file (`--output`) or emitting raw markdown (`--raw`). Use `--markdown` to override:

- `gfm` — GitHub Flavored Markdown: `> [!NOTE]` alerts, `<details>` tags
- `commonmark` — Portable markdown: `> **Note:**` blockquotes, expanded sections
- `obsidian` — Like `gfm`, but tool calls and thinking render as plain markdown (no `<details>`, no `>` callouts). Obsidian renders markdown inside `<details>` poorly and breaks fenced code nested in callouts, so the obsidian flavor avoids both — at the cost of foldability.

## Examples

```bash
# Open interactive fuzzy session picker
cc2md

# View the 3rd most recent session with thinking blocks
cc2md --last 3 --thinking

# Export a session as GFM markdown for GitHub
cc2md session.jsonl --output session.md

# Export with explicit GFM flavor
cc2md session.jsonl --raw --markdown gfm > session.md

# Export for Obsidian (foldable callouts instead of <details>)
cc2md session.jsonl --markdown obsidian --output ~/vault/Inbox/session.md

# Pipe raw markdown to another tool
cc2md --raw | pandoc -o session.pdf

# List sessions as JSON for scripting
cc2md list --json | jq '.[0].path'

# Filter sessions by project name
cc2md list my-project

# View without pager
cc2md --last 1 --no-pager
```

## Archiving Your Conversations in Git

Claude Code deletes session logs after 30 days. You can use `cc2md` to build a permanent, version-controlled archive of every AI conversation — searchable with `grep`, reviewable in any editor, and backed up to a remote.

**Set up a log repository:**

```bash
mkdir -p ~/claude-code-logs && cd ~/claude-code-logs
git init
```

**Export all sessions:**

```bash
for path in $(cc2md list --json | jq -r '.[].path'); do
  id=$(basename "$path" .jsonl)
  cc2md "$path" --raw --markdown gfm -o "${id}.md"
done
```

**Commit and push:**

```bash
git add .
git commit -m "Update logs $(date +%Y-%m-%d)"
git push
```

Run this on a schedule (cron, launchd) to keep the archive growing automatically. Over time you get a complete record of how you work with AI — which patterns you return to, how your prompting evolves, and which approaches work across projects.

## Auto-export with a Claude Code hook

`cc2md hook` reads a Claude Code Stop-hook payload from stdin and writes the current session as markdown. Add to `~/.claude/settings.json`:

```json
{
  "hooks": {
    "Stop": [
      { "hooks": [{ "type": "command", "command": "cc2md hook" }] }
    ]
  }
}
```

Files land at `~/claude-code-logs/<project>/<date>-<title>-<id8>.md` by default. The project folder is the basename of the session's working directory (e.g. `~/Projects/some-app` → `some-app`), falling back to decoding the transcript path under `~/.claude/projects/` when no working directory is available. Leading dots are trimmed (`~/.dotfiles` → `dotfiles/`) so archive folders are never hidden from tools like Obsidian.

**Configuration** (flag or env var; flag wins):

| Flag | Env var | Default |
|---|---|---|
| `--dir` | `CC2MD_HOOK_DIR` | `~/claude-code-logs` |
| `--flavor` | `CC2MD_HOOK_FLAVOR` | `obsidian` |
| `--thinking` | `CC2MD_HOOK_THINKING` | `true` |
| `--collapse` | `CC2MD_HOOK_COLLAPSE` | `true` |
| `--tool-output` | `CC2MD_HOOK_TOOL_OUTPUT` | `true` |
| `--max-lines` | `CC2MD_HOOK_MAX_LINES` | `100` |

Set `--tool-output=false` to keep each tool call's one-line header (so you can still see what Claude did) while dropping the often-large output blocks — handy for a lean conversation-focused archive.

The hook always exits 0; errors and a one-line `cc2md hook: wrote <path>` confirmation go to stderr.

**Manual alternative** (no subcommand, just jq):

```json
{
  "hooks": {
    "Stop": [
      { "hooks": [{ "type": "command",
        "command": "jq -r .transcript_path | xargs -I{} cc2md {} --markdown obsidian --thinking -o ~/claude-code-logs/$(date +%F).md" }] }
    ]
  }
}
```

## Building from Source

Requires Go 1.25+ and [just](https://github.com/casey/just).

```bash
just build
```

## License

MIT
