# linkding CLI

Command-line interface for any self-hosted [linkding](https://github.com/sissbruecker/linkding) bookmark manager instance. Built to be friendly for both humans and AI agents.

## Installation

**Homebrew (macOS/Linux):**
```bash
brew install chickenzord/tap/linkding
```

**Go install:**
```bash
go install github.com/chickenzord/linkding-cli/cmd/linkding@latest
```

**Docker:**
```bash
docker run --rm ghcr.io/chickenzord/linkding-cli:latest --help
```

**Binary download:** grab the latest release from the [releases page](https://github.com/chickenzord/linkding-cli/releases).

## Configuration

Set these environment variables (or pass them as flags):

| Variable | Flag | Description |
|---|---|---|
| `LINKDING_URL` | `--url` | Base URL of your linkding instance |
| `LINKDING_TOKEN` | `--token` | API token from Settings → Integrations |

```bash
export LINKDING_URL=https://links.example.com
export LINKDING_TOKEN=your-token-here
```

## Usage

```
linkding [command] [flags]

Commands:
  bookmark    Manage bookmarks
  tag         Manage tags
  version     Print version information
```

### Bookmarks

```bash
# List bookmarks
linkding bookmark list
linkding bookmark list --search "golang"
linkding bookmark list --archived
linkding bookmark list --limit 20 --offset 40

# Get a single bookmark
linkding bookmark get 42

# Check if a URL is already bookmarked (also returns metadata + suggested tags)
linkding bookmark check --url https://example.com

# Create a bookmark
linkding bookmark create --url https://example.com
linkding bookmark create --url https://example.com --title "Title" --tags go,cli

# Update a bookmark (PATCH — only provided fields are changed)
linkding bookmark update 42 --title "New Title" --tags go,programming

# Delete a bookmark
linkding bookmark delete 42 --yes

# Archive / unarchive
linkding bookmark archive 42
linkding bookmark unarchive 42
```

### Tags

```bash
# List tags
linkding tag list

# Get a tag
linkding tag get 5

# Create a tag
linkding tag create --name golang
```

## Agent-Friendly Features

All commands support flags designed for programmatic use:

| Flag | Effect |
|---|---|
| `--json` | Output machine-readable JSON to stdout; all human text goes to stderr |
| `-q` / `--quiet` | One item per line — pipe-friendly |
| `--dry-run` | Show what would change without mutating anything |
| `--yes` | Skip confirmation prompts (required for destructive ops in non-TTY) |

**Exit codes:**

| Code | Meaning |
|---|---|
| 0 | Success |
| 1 | General error |
| 2 | Usage / config error |
| 3 | Resource not found |
| 4 | Permission denied |
| 5 | Conflict |

**Structured errors (in `--json` mode):**

```json
{
  "error": "not_found",
  "message": "bookmark 99 not found",
  "input": { "id": 99 },
  "suggestion": "Run 'linkding bookmark list' to see available bookmarks",
  "retryable": false
}
```

**Common agentic patterns:**

```bash
# Check before creating
linkding bookmark check --url https://example.com --json | jq '.bookmark'

# Capture the ID of a newly created bookmark
ID=$(linkding bookmark create --url https://example.com --json | jq '.id')

# Collect all URLs as a plain list
linkding bookmark list -q

# Process all unread bookmarks
linkding bookmark list --unread --json | jq -r '.results[].id'
```

## License

MIT
