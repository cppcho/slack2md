package interfaces

import (
	"context"

	"github.com/cppcho/slack2md/internal/domain/entities"
	"github.com/cppcho/slack2md/internal/domain/valueobjects"
)

// FileRepository defines the interface for file operations
type FileRepository interface {
	// WriteMessages writes messages to markdown files organized by date
	WriteMessages(ctx context.Context, config valueobjects.ExportConfig, channel entities.Channel, messagesByDate map[string][]entities.Message) error

	// EnsureDirectoryExists creates directory if it doesn't exist
	EnsureDirectoryExists(path string) error
}
