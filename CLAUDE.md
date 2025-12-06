# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Development Commands

### Building
```bash
# Build the slack2md binary
make build

# Output: bin/slack2md
```

### Running
```bash
# Run via make
make run

# Run directly with go
go run ./cmd/slack2md

# Run the built binary
./bin/slack2md
```

### Testing & Code Quality
```bash
# Run all tests
make test

# Format code
make fmt

# Run go vet
make vet

# Clean build artifacts
make clean
```

## Architecture Overview

This repository exports Slack channel messages to organized markdown files.

### Directory Structure
- `cmd/slack2md/` - Main binary entry point
- `internal/` - Shared utilities private to this module
  - `common/` - Common output helpers (PrintBanner, Success, Error)
  - `slack/` - Slack API integration (client, fetcher, formatter, organizer)
  - `filewriter/` - File I/O utilities for markdown writing
- `bin/` - Compiled binaries (gitignored)

## Tool Architecture

The slack2md tool exports Slack channel messages to organized markdown files.

### Configuration (Environment Variables)
- `SLACK_BOT_TOKEN` - Bot User OAuth Token (required, starts with `xoxb-`)
- `SLACK_CHANNEL_IDS` - Comma-separated channel IDs (optional, e.g., `C1234567890,C0987654321`)
  - If empty or omitted, the tool will auto-discover all channels where the bot is invited
  - Auto-discovery includes: public and private channels where bot is a member
  - Auto-discovery excludes: archived channels, DMs, group DMs
- `SLACK_EXPORT_PATH` - Output directory path (required)
- `SLACK_APP_TOKEN` - App-Level Token (optional, starts with `xapp-`)

### Required Slack Permissions
- `channels:history` or `groups:history` (for reading messages)
- `channels:read` or `groups:read` (for reading channel info)

For private channels, the bot must be explicitly invited.

### Module Organization

**`internal/slack/client.go`** - Slack API wrapper
- `Client` struct wrapping slack-go/slack library
- `NewClient(token)` - Initialize client
- `GetChannelName(channelID)` - Fetch channel metadata
- `FetchAllChannels()` - Auto-discover all channels where bot is a member (with pagination)

**`internal/slack/fetcher.go`** - Message pagination and thread handling
- `FetchChannelMessages()` - Main fetching logic with pagination
- `fetchThreadReplies()` - Thread reply fetching
- Handles timestamp parsing and message structure

**`internal/slack/formatter.go`** - Slack→Markdown conversion
- `ConvertSlackToMarkdown()` - Converts Slack formatting to standard Markdown
- Handles bold, italic, strikethrough, links, code blocks, quotes

**`internal/slack/organizer.go`** - Date-based grouping
- `OrganizeMessagesByDate()` - Groups messages by date
- `GetSortedDates()` - Sorts dates chronologically
- Thread replies inherit parent message's date

**`internal/filewriter/writer.go`** - Markdown file I/O
- `WriteChannelMessages()` - Main export orchestration
- Creates directory structure: `<export_path>/<channel_name>/YYYY-MM-DD.md`
- `writeMessage()` - Formats parent messages with `### HH:MM` headers
- `writeThreadReply()` - Formats replies with `###### HH:MM` headers (includes full date if different from parent)

### Key Design Decisions
- **Idempotent**: Safe to run multiple times (overwrites existing files)
- **Default time window**: Last 7 days
- **Thread handling**: Replies grouped under parent message's date
- **Text-only**: Attachments/images not downloaded
- **User mentions**: Kept as-is, not resolved to usernames

## Code Patterns

### Main Binary
- Entry point: `cmd/slack2md/main.go`
- Uses `internal/common.PrintBanner()`, `Success()`, `Error()` for consistent CLI output

### Shared Utilities
- Place reusable code in `internal/<package>/`
- Internal packages are private to this module only
- Import path: `github.com/cppcho/slack2md/internal/<package>`
