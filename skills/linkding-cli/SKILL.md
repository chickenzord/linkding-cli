---
name: linkding-cli
description: Use the linkding CLI to manage bookmarks and tags on a self-hosted linkding instance. Use when the user wants to list, search, create, update, delete, archive, or check bookmarks, or manage tags.
metadata:
  {
    "author": "chickenzord",
    "version": "1.0.0",
    "openclaw":
      {
        "emoji": "🔖",
        "homepage": "https://github.com/chickenzord/linkding-cli",
        "keywords": ["linkding", "bookmarks", "tags", "self-hosted", "cli", "agent", "json"],
        "requires": { "bins": ["linkding"] },
        "install":
          [
            {
              "id": "brew-linkding-cli",
              "kind": "brew",
              "tap": "chickenzord/tap",
              "formula": "linkding-cli",
              "bins": ["linkding"],
              "label": "Install linkding-cli (brew)",
            },
          ],
      },
  }
---

# linkding-cli Skill

Use the `linkding` CLI to interact with the user's self-hosted linkding bookmark manager.

## Prerequisites

The CLI reads configuration from environment variables:
- `LINKDING_URL` — base URL of the linkding instance
- `LINKDING_TOKEN` — API token from Settings → Integrations

These are available in `.env` (already loaded in the shell via `source .env` or `make`).

If the environment variables are not set, pass `--url` and `--token` explicitly.

## Build

```bash
go build -o linkding ./cmd/linkding
```

Or use the Makefile which auto-loads `.env`:

```bash
make build
```

## Available Commands

### Bookmarks

```bash
# List bookmarks (human-readable table)
./linkding bookmark list

# Search
./linkding bookmark list --search "golang"

# Filter archived / unread / shared
./linkding bookmark list --archived
./linkding bookmark list --unread

# Paginate
./linkding bookmark list --limit 20 --offset 40

# Machine-readable JSON (stdout; warnings → stderr)
./linkding bookmark list --json

# Pipe-friendly bare URLs
./linkding bookmark list -q

# Get a single bookmark
./linkding bookmark get <id>
./linkding bookmark get <id> --json

# Check if a URL is already bookmarked (also returns metadata + suggested tags)
./linkding bookmark check --url <url>
./linkding bookmark check --url <url> --json

# Create a bookmark
./linkding bookmark create --url <url>
./linkding bookmark create --url <url> --title "Title" --description "Desc" --tags go,cli
./linkding bookmark create --url <url> --json        # returns created bookmark as JSON
./linkding bookmark create --url <url> --dry-run     # preview without mutating

# Update a bookmark (PATCH — only fields provided are changed)
./linkding bookmark update <id> --title "New Title"
./linkding bookmark update <id> --tags go,programming
./linkding bookmark update <id> --dry-run

# Delete a bookmark
./linkding bookmark delete <id> --yes                # --yes required (or non-TTY)
./linkding bookmark delete <id> --dry-run

# Archive / unarchive
./linkding bookmark archive <id>
./linkding bookmark unarchive <id>
```

### Tags

```bash
# List all tags
./linkding tag list
./linkding tag list --json
./linkding tag list -q                 # one name per line

# Get a tag by ID
./linkding tag get <id>

# Create a tag
./linkding tag create --name golang
./linkding tag create --name golang --dry-run
```

### Version

```bash
./linkding version
./linkding version --json
```

## Agent-Friendly Patterns

| Need | Flag/behavior |
|---|---|
| Machine-readable output | `--json` → JSON to stdout; all human text to stderr |
| Pipe-friendly | `-q` / `--quiet` → one item per line |
| Preview mutations | `--dry-run` → shows what would change, no API call |
| Skip prompts | `--yes` on `delete` |
| Non-TTY | prompts auto-skipped; use `--yes` to be explicit |

## Exit Codes

| Code | Meaning |
|---|---|
| 0 | Success |
| 1 | General error |
| 2 | Usage / config error |
| 3 | Resource not found |
| 4 | Permission denied |
| 5 | Conflict |

## Structured Error Format (--json mode)

```json
{
  "error": "not_found",
  "message": "bookmark 99 not found",
  "input": {"id": 99},
  "suggestion": "Run 'linkding bookmark list' to see available bookmarks",
  "retryable": false
}
```

## Common Agentic Patterns

```bash
# Check before creating (idempotent-style)
./linkding bookmark check --url https://example.com --json | jq '.bookmark'

# Collect all URLs as a list
./linkding bookmark list -q

# Add a bookmark and capture the new ID
ID=$(./linkding bookmark create --url https://example.com --json | jq '.id')

# Archive a bookmark after processing
./linkding bookmark archive "$ID"

# Batch: list unread bookmarks as JSON, process each
./linkding bookmark list --unread --json | jq -r '.results[].id'
```
