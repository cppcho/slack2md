# slack2md

Export Slack channel messages to organized markdown files.

## Features

- Exports Slack channel messages to markdown format
- Organizes messages by date into separate files
- Supports thread conversations
- Auto-discovers channels where bot is invited
- Converts Slack formatting to standard Markdown

## Structure

```
slack2md/
├── cmd/
│   └── slack2md/        # Main binary
├── internal/            # Shared utilities
│   ├── common/          # CLI helpers (banner, success/error messages)
│   ├── slack/           # Slack API integration
│   │   ├── client.go    # API wrapper
│   │   ├── fetcher.go   # Message pagination
│   │   ├── formatter.go # Markdown conversion
│   │   └── organizer.go # Date grouping
│   └── filewriter/      # Markdown file I/O
└── bin/                 # Compiled binaries (gitignored)
```

## Quick Start

### Build
```bash
# Build the binary
make build

# Output: bin/slack2md
```

### Run
```bash
# Run directly with go
go run ./cmd/slack2md

# Or run via make
make run

# Run the built binary
./bin/slack2md
```

### Install
```bash
# Install to GOPATH/bin
make install

# Then run from anywhere
slack2md
```

## Configuration

The tool is configured via environment variables:

| Variable | Required | Description |
|----------|----------|-------------|
| `SLACK_BOT_TOKEN` | Yes | Bot User OAuth Token (starts with `xoxb-`) |
| `SLACK_APP_TOKEN` | No | App-Level Token (starts with `xapp-`) |
| `SLACK_CHANNEL_IDS` | No | Comma-separated list of channel IDs (e.g., `C1234567890,C0987654321`). If omitted, auto-discovers all channels where bot is invited. |
| `SLACK_EXPORT_PATH` | Yes | Output directory path for markdown files |

### Example

```bash
export SLACK_BOT_TOKEN="xoxb-your-token-here"
export SLACK_APP_TOKEN="xapp-your-token-here"
export SLACK_CHANNEL_IDS="C1234567890,C0987654321"  # Optional
export SLACK_EXPORT_PATH="./slack-exports"
```

### Required Slack Permissions

- `channels:history` or `groups:history` (for reading messages)
- `channels:read` or `groups:read` (for reading channel info)

For private channels, the bot must be explicitly invited.

### Finding Your Channel ID

1. In Slack, right-click on the channel name
2. Select "View channel details"
3. Scroll down to find the Channel ID (starts with `C`)

## Development

```bash
# Run tests
make test

# Format code
make fmt

# Run go vet
make vet

# Clean build artifacts
make clean
```

## Output Format

### Directory Structure

Messages are organized into markdown files:
```
<export_path>/
└── <channel_name>/
    ├── 2024-01-15.md
    ├── 2024-01-16.md
    └── 2024-01-17.md
```

### Markdown Format

Each message is formatted with:
- `### HH:MM` for parent messages
- `###### HH:MM` for thread replies

Example:
```markdown
# 2025-12-04

### 10:05 First message content
Multi-line message content
is preserved

### 11:00 Parent message with thread

###### 13:00 Thread reply same day
Reply content

###### 2025-12-05 09:00 Thread reply different day
(Full date shown when reply is on different day than parent)
```

### Formatting Conversion

The tool automatically converts Slack formatting to standard Markdown:

| Slack | Markdown |
|-------|----------|
| `*bold*` | `**bold**` |
| `_italic_` | `*italic*` |
| `~strike~` | `~~strike~~` |
| `` `code` `` | `` `code` `` |
| `<url\|text>` | `[text](url)` |
| `> quote` | `> quote` |

## Advanced Usage

### Using a .env File

Create a `.env` file for easy configuration:

```bash
# .env
export SLACK_BOT_TOKEN="xoxb-..."
export SLACK_APP_TOKEN="xapp-..."
export SLACK_CHANNEL_IDS="C1234567890"
export SLACK_EXPORT_PATH="$HOME/slack-exports"
```

Source it before running:

```bash
source .env
./bin/slack2md
```

### Periodic Execution

To keep your exports up-to-date, schedule the tool to run periodically using cron:

```bash
# Edit crontab
crontab -e

# Add this line to run daily at 2 AM
0 2 * * * cd /path/to/slack2md && /path/to/slack2md/bin/slack2md
```

Make sure to set environment variables in a way that's accessible to cron (e.g., in a `.env` file that's sourced).

## Troubleshooting

### "Channel not found" error
- Verify the channel ID is correct
- Ensure the bot has been invited to the channel (private channels require explicit invitation)
- Check that the bot has the required permissions

### "Missing permissions" error
- Verify your bot has `channels:history` or `groups:history` scope
- Re-install the app to your workspace if you added scopes after installation

### Empty exports
- Check that the channel has messages in the last 7 days
- Verify the time range matches your expectations

## Notes

- The tool exports the last 7 days by default
- Thread replies are grouped under the parent message's date
- Attachments (images, files) are not downloaded, only text is exported
- User mentions (`@user`) are kept as-is and not resolved to usernames
- Message edits/deletions are reflected on the next run (idempotent design)
