package interfaces

import (
	"context"

	"github.com/cppcho/slack2md/internal/domain/entities"
	"github.com/cppcho/slack2md/internal/domain/valueobjects"
)

// SlackRepository defines the interface for Slack data operations
type SlackRepository interface {
	// FetchAllChannels discovers all channels where bot is a member
	FetchAllChannels(ctx context.Context) ([]entities.Channel, error)

	// FetchChannelMessages fetches messages within time range
	FetchChannelMessages(ctx context.Context, channelID string, timeRange valueobjects.TimeRange) ([]entities.Message, error)

	// GetUserDisplayName fetches user display name (with caching)
	GetUserDisplayName(ctx context.Context, userID string) (string, error)
}
