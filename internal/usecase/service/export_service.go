package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/cppcho/slack2md/internal/domain/entities"
	"github.com/cppcho/slack2md/internal/domain/valueobjects"
	"github.com/cppcho/slack2md/internal/usecase/dto"
	"github.com/cppcho/slack2md/internal/usecase/interfaces"
)

// ExportService handles the export of Slack channels to markdown
type ExportService struct {
	slackRepo interfaces.SlackRepository
	fileRepo  interfaces.FileRepository
	logger    interfaces.Logger
}

// NewExportService creates a new ExportService
func NewExportService(
	slackRepo interfaces.SlackRepository,
	fileRepo interfaces.FileRepository,
	logger interfaces.Logger,
) *ExportService {
	return &ExportService{
		slackRepo: slackRepo,
		fileRepo:  fileRepo,
		logger:    logger,
	}
}

// ExportChannels orchestrates the entire export process
func (s *ExportService) ExportChannels(ctx context.Context, input dto.ExportChannelsInput) (*dto.ExportChannelsOutput, error) {
	s.logger.Info("Starting export process")

	// Create time range from days back
	timeRange, err := valueobjects.NewTimeRangeFromDaysBack(input.DaysBack)
	if err != nil {
		return nil, fmt.Errorf("failed to create time range: %w", err)
	}

	// Create export config
	exportConfig, err := valueobjects.NewExportConfig(input.ExportPath, *timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to create export config: %w", err)
	}

	// Discover or filter channels
	channels, err := s.discoverChannels(ctx, input.ChannelIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to discover channels: %w", err)
	}

	if len(channels) == 0 {
		return nil, fmt.Errorf("no channels to export")
	}

	s.logger.Info("Starting export for %d channels", len(channels))

	// Export each channel
	var results []entities.ExportResult
	successCount := 0
	failureCount := 0

	for _, channel := range channels {
		s.logger.Info("Processing channel %s (%s)", channel.Name, channel.ID)

		result := s.exportChannel(ctx, channel, *exportConfig)
		results = append(results, result)

		if result.Success {
			successCount++
			s.logger.Info("Channel %s exported successfully: %d messages across %d days", channel.Name, result.MessageCount, result.DateCount)
		} else {
			failureCount++
			s.logger.Error("Failed to export channel %s: %v", channel.Name, result.Error)
		}
	}

	// Create summary
	summary := entities.ExportSummary{
		TotalChannels: len(channels),
		SuccessCount:  successCount,
		FailureCount:  failureCount,
		Results:       results,
	}

	return &dto.ExportChannelsOutput{
		Summary: summary,
	}, nil
}

// discoverChannels auto-discovers or filters channels based on channel IDs
func (s *ExportService) discoverChannels(ctx context.Context, channelIDs []string) ([]entities.Channel, error) {
	// Fetch all channels where bot is a member
	allChannels, err := s.slackRepo.FetchAllChannels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch channels: %w", err)
	}

	// If no specific channels requested, return all
	if len(channelIDs) == 0 {
		s.logger.Info("Using all %d discovered channels", len(allChannels))
		return allChannels, nil
	}

	// Filter channels by requested IDs
	channelIDSet := make(map[string]struct{})
	for _, id := range channelIDs {
		channelIDSet[id] = struct{}{}
	}

	var filtered []entities.Channel
	for _, ch := range allChannels {
		if _, exists := channelIDSet[ch.ID]; exists {
			filtered = append(filtered, ch)
		}
	}

	s.logger.Info("Filtered to %d channels from %d total", len(filtered), len(allChannels))
	return filtered, nil
}

// exportChannel exports a single channel
func (s *ExportService) exportChannel(ctx context.Context, channel entities.Channel, config valueobjects.ExportConfig) entities.ExportResult {
	// Fetch messages
	messages, err := s.slackRepo.FetchChannelMessages(ctx, channel.ID, config.TimeRange)
	if err != nil {
		return entities.ExportResult{
			Channel: channel,
			Success: false,
			Error:   fmt.Errorf("failed to fetch messages: %w", err),
		}
	}

	s.logger.Debug("Fetched %d messages from channel %s", len(messages), channel.Name)

	// Organize messages by date
	messagesByDate := s.organizeMessagesByDate(messages)

	s.logger.Debug("Organized messages into %d dates", len(messagesByDate))

	// Write to files
	if err := s.fileRepo.WriteMessages(ctx, config, channel, messagesByDate); err != nil {
		return entities.ExportResult{
			Channel:      channel,
			MessageCount: len(messages),
			DateCount:    len(messagesByDate),
			Success:      false,
			Error:        fmt.Errorf("failed to write messages: %w", err),
		}
	}

	return entities.ExportResult{
		Channel:      channel,
		MessageCount: len(messages),
		DateCount:    len(messagesByDate),
		Success:      true,
	}
}

// organizeMessagesByDate groups messages by date, with thread replies grouped under parent's date
func (s *ExportService) organizeMessagesByDate(messages []entities.Message) map[string][]entities.Message {
	organized := make(map[string][]entities.Message)

	for _, msg := range messages {
		// Use parent message's date for grouping
		date := msg.Timestamp.Format("2006-01-02")

		// Only include top-level messages (not replies in the main list)
		// Thread replies are already attached to their parent message
		if msg.ThreadTS == "" || msg.IsParent {
			organized[date] = append(organized[date], msg)
		}
	}

	// Sort messages within each date by timestamp
	for date := range organized {
		sort.Slice(organized[date], func(i, j int) bool {
			return organized[date][i].Timestamp.Before(organized[date][j].Timestamp)
		})
	}

	return organized
}
