# Clean Architecture Refactoring Plan for slack2md

## Overview

Refactor slack2md from a layered architecture to Clean Architecture with:
- **Domain Layer**: Business entities (Channel, Message, TimeRange, ExportConfig)
- **Use Case Layer**: ExportService with repository/gateway interfaces
- **Adapter Layer**: Repository implementations (formatters are concrete helpers)
- **Infrastructure Layer**: Slack client, file system, logging, configuration
- **Selective Dependency Injection**: Only inject external dependencies (Slack API, file system, logger)
- **Enhanced Logging**: Comprehensive logging for all Slack API calls
- **Configuration Validation**: Strict validation with clear error messages

## User Requirements
- **Architecture**: Flexible clean architecture (focus on layer separation, domain can use stdlib)
- **DI Pattern**: Selective dependency injection - only inject external dependencies (Slack API, file system, logger)
- **Use Cases**: Service-based (ExportService with multiple methods)
- **Enhancements**: Add configuration validation + comprehensive logging

## New Directory Structure

```
slack2md/
├── cmd/slack2md/
│   └── main.go                           # Composition root (DI wiring)
├── internal/
│   ├── domain/                           # Domain Layer
│   │   ├── entities/
│   │   │   ├── channel.go
│   │   │   ├── message.go
│   │   │   └── export_result.go
│   │   └── valueobjects/
│   │       ├── time_range.go
│   │       └── export_config.go
│   ├── usecase/                          # Use Case Layer
│   │   ├── interfaces/
│   │   │   ├── slack_repository.go
│   │   │   ├── file_repository.go
│   │   │   └── logger.go
│   │   ├── dto/
│   │   │   ├── export_input.go
│   │   │   └── export_output.go
│   │   └── service/
│   │       ├── export_service.go
│   │       └── export_service_test.go
│   ├── adapter/                          # Interface Adapters Layer
│   │   ├── repository/
│   │   │   ├── slack_repository_impl.go
│   │   │   ├── slack_repository_impl_test.go
│   │   │   ├── file_repository_impl.go
│   │   │   └── file_repository_impl_test.go
│   │   ├── formatter/
│   │   │   ├── markdown_formatter.go
│   │   │   └── markdown_formatter_test.go
│   │   └── presenter/
│   │       └── console_presenter.go
│   └── infrastructure/                   # Infrastructure Layer
│       ├── slack/
│       │   ├── client.go
│       │   ├── client_test.go
│       │   └── user_cache.go
│       ├── filesystem/
│       │   ├── writer.go
│       │   └── writer_test.go
│       ├── logger/
│       │   ├── logger_impl.go
│       │   └── logger_test.go
│       └── config/
│           ├── env_loader.go
│           ├── validator.go
│           └── validator_test.go
```

## Key Architectural Decisions

### 1. Domain Entities
- `Channel`: ID + Name
- `Message`: Timestamp, ThreadTS, Text, UserDisplayName, Replies, IsParent
- `ExportResult`: Channel, MessageCount, DateCount, Success, Error
- `TimeRange`: Start + End (value object with validation)
- `ExportConfig`: ExportPath + TimeRange (value object)

### 2. Use Case Interfaces (defined in usecase layer)

**SlackRepository**
```go
FetchAllChannels(ctx) ([]Channel, error)
FetchChannelMessages(ctx, channelID, timeRange) ([]Message, error)
GetUserDisplayName(ctx, userID) (string, error)
```

**FileRepository**
```go
WriteMessages(ctx, config, channel, messagesByDate) error
EnsureDirectoryExists(path) error
```

**Logger**
```go
Info(msg, args...)
Warn(msg, args...)
Error(msg, args...)
Debug(msg, args...)
```

### 3. ExportService (Use Case Service)

**Dependencies (injected):**
- SlackRepository (interface)
- FileRepository (interface)
- Logger (interface)

**Internal helpers (created directly, not injected):**
- Formatter (concrete MarkdownFormatter)
- Organizer logic (implemented as private methods)

Methods:
- `ExportChannels(ctx, input)` - Main orchestration
- `discoverChannels(ctx, channelIDs)` - Auto-discover or filter channels
- `exportChannel(ctx, channel, config)` - Export single channel
- `organizeMessagesByDate(messages)` - Group messages by date

### 4. Infrastructure Implementations

**SlackClient** (infrastructure/slack/client.go)
- Wraps `slack-go/slack` library
- Provides interface for mocking
- Methods: GetConversationsForUser, GetConversationHistory, GetConversationReplies, GetUserInfo

**UserCache** (infrastructure/slack/user_cache.go)
- Thread-safe cache with sync.RWMutex
- Stores user ID → display name mappings

**FileSystemWriter** (infrastructure/filesystem/writer.go)
- Wraps os.WriteFile and os.MkdirAll
- Interface allows testing without actual file I/O

**Logger** (infrastructure/logger/logger_impl.go)
- Simple stdout logger with log levels (DEBUG, INFO, WARN, ERROR)
- Configurable minimum level

**ConfigValidator** (infrastructure/config/validator.go)
- Validates BotToken (required, starts with "xoxb-")
- Validates AppToken (if provided, starts with "xapp-")
- Validates ExportPath (required, non-empty)
- Validates DaysBack (0-365 range)
- Returns formatted multi-line error message

### 5. Logging Strategy

**What to log:**
- SlackRepository: "Fetching channels", "Found X channels", "Fetching messages for channel X", "Fetched X messages", errors
- FileRepository: "Writing messages to {path}", "Created directory {path}", errors
- ExportService: "Starting export for X channels", "Processing channel X", "Channel X exported successfully", errors

**Where to log:**
- Before/after Slack API calls (with channel names/IDs)
- Before/after file operations (with paths)
- Success/failure for each channel export

### 6. Dependency Injection Wiring (main.go)

```go
1. Load config from environment
2. Validate config
3. Create infrastructure: logger, slackClient, fileWriter, userCache
4. Create adapters: slackRepo, fileRepo, presenter
5. Create use case service: exportService (creates formatter internally)
6. Build input DTO
7. Execute use case
8. Present results
```

## Implementation Steps

### Phase 1: Domain Layer (No external dependencies)

**Files to create:**
1. `internal/domain/entities/channel.go`
   - Channel struct with NewChannel constructor
   - Validation: non-empty ID and Name

2. `internal/domain/entities/message.go`
   - Message struct with NewMessage constructor
   - AddReply method to append replies

3. `internal/domain/entities/export_result.go`
   - ExportResult struct (Channel, MessageCount, DateCount, Success, Error)
   - ExportSummary struct (TotalChannels, SuccessCount, FailureCount, Results)

4. `internal/domain/valueobjects/time_range.go`
   - TimeRange struct with NewTimeRange constructor
   - NewTimeRangeFromDaysBack factory
   - Validation: end must be after start

5. `internal/domain/valueobjects/export_config.go`
   - ExportConfig struct with NewExportConfig constructor
   - Validation: non-empty export path

**Verify:** `make test` passes

### Phase 2: Use Case Interfaces

**Files to create:**
6. `internal/usecase/interfaces/slack_repository.go`
   - SlackRepository interface (FetchAllChannels, FetchChannelMessages, GetUserDisplayName)

7. `internal/usecase/interfaces/file_repository.go`
   - FileRepository interface (WriteMessages, EnsureDirectoryExists)

8. `internal/usecase/interfaces/logger.go`
   - Logger interface (Info, Warn, Error, Debug)

9. `internal/usecase/dto/export_input.go`
   - ExportChannelsInput struct (BotToken, AppToken, ChannelIDs, ExportPath, DaysBack)

10. `internal/usecase/dto/export_output.go`
    - ExportChannelsOutput struct (Summary)

### Phase 3: Infrastructure Layer

**Files to create:**
11. `internal/infrastructure/logger/logger_impl.go`
    - LogLevel constants (DEBUG, INFO, WARN, ERROR)
    - LoggerImpl struct with level field
    - NewLogger constructor
    - Implement Logger interface methods

12. `internal/infrastructure/logger/logger_test.go`
    - Test Info, Warn, Error with different log levels

13. `internal/infrastructure/config/env_loader.go`
    - Move from `cmd/slack2md/config.go`
    - EnvConfig struct
    - LoadFromEnv() function

14. `internal/infrastructure/config/validator.go`
    - ConfigValidator struct
    - Validate method with comprehensive checks:
      - BotToken (required, starts with "xoxb-")
      - AppToken (if provided, starts with "xapp-")
      - ExportPath (required, non-empty)
      - DaysBack (0-365 range)
    - Return formatted multi-line error

15. `internal/infrastructure/config/validator_test.go`
    - Test validation for each field
    - Test edge cases

16. `internal/infrastructure/slack/user_cache.go`
    - UserCache struct with map and sync.RWMutex
    - NewUserCache constructor
    - Get and Set methods (thread-safe)

17. `internal/infrastructure/slack/client.go`
    - SlackClient interface (for mocking)
    - SlackClientImpl struct wrapping *slack.Client
    - NewSlackClient constructor
    - Implement interface methods (delegate to slack.Client)

18. `internal/infrastructure/slack/client_test.go`
    - Test basic client creation

19. `internal/infrastructure/filesystem/writer.go`
    - FileWriter interface
    - FileSystemWriter struct
    - NewFileSystemWriter constructor
    - WriteFile and MkdirAll methods (wrap os functions)

20. `internal/infrastructure/filesystem/writer_test.go`
    - Test file writing to temp directory

**Verify:** `make test` passes

### Phase 4: Adapter Layer

**Files to create:**
21. `internal/adapter/formatter/markdown_formatter.go`
    - Move logic from `internal/slack/formatter.go`
    - MarkdownFormatter struct
    - NewMarkdownFormatter constructor
    - ConvertSlackToMarkdown method (concrete implementation, not interface)

22. `internal/adapter/formatter/markdown_formatter_test.go`
    - Move tests from slack package
    - Test bold, italic, links, code blocks, bullets

23. `internal/adapter/presenter/console_presenter.go`
    - Move from `internal/common/common.go`
    - ConsolePresenter struct
    - Success, Error methods
    - PrintChannelList, PrintSummary methods

24. `internal/adapter/repository/slack_repository_impl.go`
    - Refactor from `internal/slack/client.go` + `internal/slack/fetcher.go`
    - SlackRepositoryImpl struct (client, logger, userCache)
    - Creates MarkdownFormatter internally (not injected)
    - NewSlackRepository constructor
    - Implement SlackRepository interface:
      - FetchAllChannels: pagination + logging
      - FetchChannelMessages: pagination, thread fetching, formatting + logging
      - GetUserDisplayName: with caching + logging
    - Helper: fetchThreadReplies, parseSlackTimestamp

25. `internal/adapter/repository/slack_repository_impl_test.go`
    - Test with mock SlackClient
    - Test pagination, threads, caching

26. `internal/adapter/repository/file_repository_impl.go`
    - Refactor from `internal/filewriter/writer.go`
    - FileRepositoryImpl struct (writer, logger)
    - NewFileRepository constructor
    - Implement FileRepository interface:
      - WriteMessages: organize by date, write markdown files + logging
      - EnsureDirectoryExists
    - Helpers: writeMarkdownFile, writeMessage, writeThreadReply, sanitizeChannelName

27. `internal/adapter/repository/file_repository_impl_test.go`
    - Test with mock FileWriter
    - Test markdown formatting, directory creation

**Verify:** `make test` passes

### Phase 5: Use Case Service

**Files to create:**
28. `internal/usecase/service/export_service.go`
    - Refactor orchestration from `cmd/slack2md/main.go`
    - ExportService struct (slackRepo, fileRepo, logger)
    - NewExportService constructor (only injects repositories and logger)
    - ExportChannels method:
      1. Create TimeRange from DaysBack
      2. Create ExportConfig
      3. Discover/filter channels (with logging)
      4. Loop through channels:
         - Fetch messages (with logging)
         - Organize by date
         - Write files (with logging)
         - Track results
      5. Return ExportSummary
    - Helper methods: discoverChannels, exportChannel, organizeMessagesByDate

29. `internal/usecase/service/export_service_test.go`
    - Mock all interfaces
    - Test successful export
    - Test partial failure
    - Test channel filtering
    - Test empty channels

**Verify:** `make test` passes

### Phase 6: Refactor Main (Composition Root)

**Files to modify:**
30. `cmd/slack2md/main.go`
    - Refactor to wire dependencies:
      1. Load config
      2. Validate config
      3. Create infrastructure (logger, slackClient, fileWriter, userCache)
      4. Create adapters (slackRepo, fileRepo, presenter)
      5. Create use case service (formatter created internally by repositories)
      6. Build input DTO
      7. Execute use case
      8. Present results
    - Remove processChannel, discoverChannels, ConvertMessage functions (moved to use cases)

**Verify:** `make build && ./bin/slack2md` (manual test with real Slack workspace)

### Phase 7: Cleanup

**Files to delete:**
31. `internal/slack/client.go` → moved to infrastructure + adapter
32. `internal/slack/fetcher.go` → moved to adapter/repository
33. `internal/slack/formatter.go` → moved to adapter/formatter
34. `internal/slack/organizer.go` → moved to usecase/service
35. `internal/slack/interface.go` → moved to infrastructure
36. `internal/filewriter/writer.go` → moved to adapter/repository
37. `internal/common/common.go` → moved to adapter/presenter
38. `cmd/slack2md/config.go` → moved to infrastructure/config
39. Remove empty directories: `internal/slack/`, `internal/filewriter/`, `internal/common/`

**Verify:** `make test && make build` passes

### Phase 8: Update Documentation

**Files to modify:**
40. `CLAUDE.md` - Update with new architecture
41. `README.md` - Update architecture section if needed

## Critical Files Reference

### Current files to refactor:
- `/Users/cppcho/dev/slack2md/cmd/slack2md/main.go` - Orchestration → ExportService + main wiring
- `/Users/cppcho/dev/slack2md/cmd/slack2md/config.go` - Config → infrastructure/config
- `/Users/cppcho/dev/slack2md/internal/slack/client.go` - Client → infrastructure/slack + adapter/repository
- `/Users/cppcho/dev/slack2md/internal/slack/fetcher.go` - Fetcher → adapter/repository
- `/Users/cppcho/dev/slack2md/internal/slack/formatter.go` - Formatter → adapter/formatter
- `/Users/cppcho/dev/slack2md/internal/slack/organizer.go` - Organizer → usecase/service
- `/Users/cppcho/dev/slack2md/internal/filewriter/writer.go` - Writer → adapter/repository
- `/Users/cppcho/dev/slack2md/internal/common/common.go` - Common → adapter/presenter

## Testing Strategy

### Unit Tests
- Domain entities: validation logic
- Use case service: with mocked interfaces
- Repository implementations: with mocked clients/writers
- Infrastructure: basic functionality

### Integration Tests
- Manual: Run `./bin/slack2md` with real Slack workspace
- Verify: Channels discovered, messages fetched, files written correctly

## Verification Checkpoints

After each phase:
- ✅ All tests pass: `make test`
- ✅ Code builds: `make build`
- ✅ No compilation errors

Final verification:
- ✅ All old files deleted
- ✅ Binary runs successfully: `./bin/slack2md`
- ✅ Logging shows Slack API calls
- ✅ Configuration validation rejects invalid configs
- ✅ Export produces correct markdown files
