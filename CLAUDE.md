# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Principles

When working with this codebase, follow these core principles:

1. **Golang Best Practices**
   - Use idiomatic Go code patterns
   - Proper error handling (always check errors)
   - Effective use of interfaces and composition
   - Clear naming conventions
   - Leverage Go's standard library

2. **Clean Architecture (Practical Application)**
   - Maintain clear layer separation (Domain → Use Case → Adapter)
   - Follow dependency rule: dependencies point inward
   - Keep domain logic independent of external frameworks
   - Use selective dependency injection for external dependencies only

3. **Git Commit Convention**
   - Follow conventional commit format: `type: description`
   - Do not include scope (e.g., `feat: add user cache`, not `feat(cache): add user cache`)
   - Use short bullet points in commit body for details
   - Common types: `feat`, `fix`, `refactor`, `docs`, `test`, `chore`

4. **Test & Build Validation**
   - Every completed change must pass tests and build successfully
   - Run `make test` and `make build` before considering work complete
   - Fix any failing tests or build errors before committing
   - Never leave the codebase in a broken state

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

This repository exports Slack channel messages to organized markdown files using **Clean Architecture** principles with selective dependency injection.

### Clean Architecture Layers

The application follows Clean Architecture with clear separation of concerns across three main layers:

1. **Domain** (innermost) - Business entities and value objects
2. **Use Case** - Application business rules and orchestration
3. **Adapter** (outermost) - Interface adapters and external integrations

### Directory Structure

```
slack2md/
├── cmd/slack2md/
│   └── main.go                           # Composition root (DI wiring)
├── internal/
│   ├── domain/                           # Domain Layer
│   │   ├── entities/
│   │   │   ├── channel.go               # Channel entity
│   │   │   ├── message.go               # Message entity
│   │   │   └── export_result.go         # Export result entities
│   │   └── valueobjects/
│   │       ├── time_range.go            # TimeRange value object
│   │       └── export_config.go         # ExportConfig value object
│   ├── usecase/                          # Use Case Layer
│   │   ├── interfaces/
│   │   │   ├── slack_repository.go      # Slack repository interface
│   │   │   ├── file_repository.go       # File repository interface
│   │   │   └── logger.go                # Logger interface
│   │   ├── dto/
│   │   │   ├── export_input.go          # Input DTO
│   │   │   └── export_output.go         # Output DTO
│   │   └── service/
│   │       └── export_service.go        # Main export orchestration
│   └── adapter/                          # Adapter Layer
│       ├── slack/
│       │   ├── client.go                # Slack API client wrapper + interface
│       │   ├── user_cache.go            # User display name cache
│       │   ├── repository.go            # Slack API repository implementation
│       │   └── formatter.go             # Slack→Markdown formatter
│       ├── filesystem/
│       │   ├── writer.go                # File system writer + interface
│       │   └── repository.go            # File system repository implementation
│       ├── config/
│       │   ├── env_loader.go            # Environment config loader
│       │   └── validator.go             # Config validation
│       ├── logger/
│       │   └── logger.go                # Logger implementation
│       └── presenter/
│           └── console_presenter.go     # Console output presenter
└── bin/                                  # Compiled binaries (gitignored)
```

## Configuration (Environment Variables)

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

## Layer Responsibilities

### Domain Layer (`internal/domain/`)

Pure business logic with no external dependencies. Contains:

**Entities:**
- `Channel`: ID and Name with validation
- `Message`: Timestamp, Text, UserDisplayName, Replies, thread metadata
- `ExportResult`: Success/failure status for channel exports
- `ExportSummary`: Aggregate results for all exports

**Value Objects:**
- `TimeRange`: Start and End times with validation
- `ExportConfig`: ExportPath and TimeRange configuration

### Use Case Layer (`internal/usecase/`)

Application business rules that orchestrate the export process.

**Interfaces** (defined here, implemented in adapters):
- `SlackRepository`: Fetch channels and messages from Slack
- `FileRepository`: Write messages to markdown files
- `Logger`: Logging operations

**Service:**
- `ExportService`: Orchestrates channel discovery, message fetching, organization, and file writing
  - `ExportChannels()`: Main entry point
  - `discoverChannels()`: Auto-discover or filter channels
  - `exportChannel()`: Export single channel
  - `organizeMessagesByDate()`: Group messages by date

**DTOs:**
- `ExportChannelsInput`: Input parameters (tokens, channel IDs, export path, days back)
- `ExportChannelsOutput`: Export summary results

### Adapter Layer (`internal/adapter/`)

Interface adapters that handle external integrations and convert between use cases and external interfaces. Organized by domain for cohesion.

**Slack Domain (`adapter/slack/`):**
- `SlackClient`: Wraps slack-go/slack library with interface
- `UserCache`: Thread-safe cache for user display names
- `SlackRepositoryImpl`: Implements Slack data fetching
  - Uses SlackClient for API calls
  - Creates MarkdownFormatter internally (selective DI)
  - Handles pagination, threads, user display name caching
  - Converts Slack messages to domain entities
- `MarkdownFormatter`: Converts Slack mrkdwn to standard Markdown
  - Handles bold, italic, strikethrough, links
  - Processes HTML entities
  - Converts bullet points

**Filesystem Domain (`adapter/filesystem/`):**
- `FileWriter`: Wraps os.WriteFile and os.MkdirAll with interface
- `FileRepositoryImpl`: Implements file writing
  - Uses FileWriter for file operations
  - Organizes files by date
  - Formats markdown with proper headers

**Config (`adapter/config/`):**
- `EnvConfig`: Environment variable loader
- `ConfigValidator`: Validates configuration with detailed error messages
  - Token format validation (xoxb-, xapp-)
  - Required field checks
  - Range validation (DaysBack 0-365)

**Logger (`adapter/logger/`):**
- `LoggerImpl`: stdout logger with log levels (DEBUG, INFO, WARN, ERROR)

**Presenter (`adapter/presenter/`):**
- `ConsolePresenter`: CLI output formatting
  - Success/error messages
  - Channel lists
  - Export summaries
  - Progress information

## Dependency Injection Pattern

**Selective Dependency Injection** - Only inject external dependencies:

- `SlackRepository`, `FileRepository`, `Logger` are injected
- Internal helpers like `MarkdownFormatter` are created directly by repositories
- Composition root in `main.go` wires all dependencies

**Main.go Composition:**
```go
1. Load config from environment
2. Validate config
3. Create adapter layer (logger, slackClient, fileWriter, userCache, repositories, presenter)
4. Create use case service (exportService)
5. Build input DTO
6. Execute use case
7. Present results
```

## Data Flow

```
Environment Variables
    ↓
Config Validation (Adapter Layer)
    ↓
Adapter Layer Components (Slack Client, File Writer, Logger, Cache)
    ↓
Use Case Layer (ExportService orchestrates via repository interfaces)
    ↓
Domain Entities (Channel, Message, TimeRange, etc.)
    ↓
Adapter Layer (Repositories format and write via FileWriter)
    ↓
File System (Files written to disk)
```

## Key Design Decisions

- **Clean Architecture**: Clear separation of concerns with dependency inversion
- **Selective DI**: Only inject external dependencies (Slack API, file system, logger)
- **Idempotent**: Safe to run multiple times (overwrites existing files)
- **Default time window**: Last 7 days
- **Thread handling**: Replies grouped under parent message's date
- **Text-only**: Attachments/images not downloaded
- **Comprehensive logging**: All Slack API calls and file operations logged
- **Configuration validation**: Strict validation with detailed error messages
- **User caching**: Display names cached to reduce API calls
- **Thread-safe**: User cache uses sync.RWMutex for concurrent access

## Code Patterns

### Adding New Features

1. **Domain changes**: Add entities/value objects in `internal/domain/`
2. **Use case changes**: Update service in `internal/usecase/service/`
3. **New interfaces**: Define in `internal/usecase/interfaces/`
4. **Implementation**: Add in appropriate `internal/adapter/` subdirectory (organized by domain)
5. **Wire dependencies**: Update `main.go` composition root

### Import Paths
- Domain: `github.com/cppcho/slack2md/internal/domain/{entities|valueobjects}`
- Use Cases: `github.com/cppcho/slack2md/internal/usecase/{interfaces|dto|service}`
- Adapters: `github.com/cppcho/slack2md/internal/adapter/{slack|filesystem|config|logger|presenter}`
