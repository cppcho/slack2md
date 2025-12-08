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
   - Maintain clear layer separation (Domain → Use Case → Adapter → Infrastructure)
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

The application follows Clean Architecture with clear separation of concerns across four main layers:

1. **Domain** (innermost) - Business entities and value objects
2. **Use Case** - Application business rules and orchestration
3. **Adapter** - Interface adapters (repositories, presenters, formatters)
4. **Infrastructure** (outermost) - External frameworks and tools

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
│   ├── adapter/                          # Adapter Layer
│   │   ├── repository/
│   │   │   ├── slack_repository_impl.go # Slack API repository
│   │   │   └── file_repository_impl.go  # File system repository
│   │   ├── formatter/
│   │   │   └── markdown_formatter.go    # Slack→Markdown formatter
│   │   └── presenter/
│   │       └── console_presenter.go     # Console output presenter
│   └── infrastructure/                   # Infrastructure Layer
│       ├── slack/
│       │   ├── client.go                # Slack API client wrapper
│       │   └── user_cache.go            # User display name cache
│       ├── filesystem/
│       │   └── writer.go                # File system writer
│       ├── logger/
│       │   └── logger_impl.go           # Logger implementation
│       └── config/
│           ├── env_loader.go            # Environment config loader
│           └── validator.go             # Config validation
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

Interface adapters that convert between use cases and external interfaces.

**Repositories:**
- `SlackRepositoryImpl`: Implements Slack data fetching
  - Wraps infrastructure Slack client
  - Creates MarkdownFormatter internally (selective DI)
  - Handles pagination, threads, user display name caching
  - Converts Slack messages to domain entities

- `FileRepositoryImpl`: Implements file writing
  - Wraps infrastructure FileWriter
  - Organizes files by date
  - Formats markdown with proper headers

**Formatter:**
- `MarkdownFormatter`: Converts Slack mrkdwn to standard Markdown
  - Handles bold, italic, strikethrough, links
  - Processes HTML entities
  - Converts bullet points

**Presenter:**
- `ConsolePresenter`: CLI output formatting
  - Success/error messages
  - Channel lists
  - Export summaries
  - Progress information

### Infrastructure Layer (`internal/infrastructure/`)

External frameworks, tools, and drivers.

**Slack:**
- `SlackClient`: Wraps slack-go/slack library with interface
- `UserCache`: Thread-safe cache for user display names

**FileSystem:**
- `FileWriter`: Wraps os.WriteFile and os.MkdirAll

**Logger:**
- `LoggerImpl`: stdout logger with log levels (DEBUG, INFO, WARN, ERROR)

**Config:**
- `EnvConfig`: Environment variable loader
- `ConfigValidator`: Validates configuration with detailed error messages
  - Token format validation (xoxb-, xapp-)
  - Required field checks
  - Range validation (DaysBack 0-365)

## Dependency Injection Pattern

**Selective Dependency Injection** - Only inject external dependencies:

- `SlackRepository`, `FileRepository`, `Logger` are injected
- Internal helpers like `MarkdownFormatter` are created directly by repositories
- Composition root in `main.go` wires all dependencies

**Main.go Composition:**
```go
1. Load config from environment
2. Validate config
3. Create infrastructure (logger, slackClient, fileWriter, userCache)
4. Create adapters (slackRepo, fileRepo, presenter)
5. Create use case service (exportService)
6. Execute use case with input DTO
7. Present results
```

## Data Flow

```
Environment Variables
    ↓
Config Validation
    ↓
Infrastructure Layer (Slack Client, File Writer, Logger, Cache)
    ↓
Adapter Layer (Repositories wrap infrastructure)
    ↓
Use Case Layer (ExportService orchestrates)
    ↓
Domain Entities (Channel, Message, TimeRange, etc.)
    ↓
Adapter Layer (Repositories format and write)
    ↓
Infrastructure Layer (Files written to disk)
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
4. **Implementation**: Add in `internal/adapter/` or `internal/infrastructure/`
5. **Wire dependencies**: Update `main.go` composition root

### Import Paths
- Domain: `github.com/cppcho/slack2md/internal/domain/{entities|valueobjects}`
- Use Cases: `github.com/cppcho/slack2md/internal/usecase/{interfaces|dto|service}`
- Adapters: `github.com/cppcho/slack2md/internal/adapter/{repository|formatter|presenter}`
- Infrastructure: `github.com/cppcho/slack2md/internal/infrastructure/{slack|filesystem|logger|config}`
