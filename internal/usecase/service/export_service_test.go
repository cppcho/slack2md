package service

import (
	"context"
	"testing"
	"time"

	"github.com/cppcho/slack2md/internal/adapter/logger"
	"github.com/cppcho/slack2md/internal/domain/entities"
	"github.com/cppcho/slack2md/internal/domain/valueobjects"
	"github.com/cppcho/slack2md/internal/usecase/dto"
)

// Mock repositories for testing with configurable behavior
type mockSlackRepo struct {
	channels    []entities.Channel
	channelsErr error
	// messagesByChannel maps channel ID to messages for that channel
	messagesByChannel map[string][]entities.Message
	messagesErr       error
}

func (m *mockSlackRepo) FetchAllChannels(ctx context.Context) ([]entities.Channel, error) {
	return m.channels, m.channelsErr
}

func (m *mockSlackRepo) FetchChannelMessages(ctx context.Context, channelID string, timeRange valueobjects.TimeRange) ([]entities.Message, error) {
	if m.messagesErr != nil {
		return nil, m.messagesErr
	}
	if m.messagesByChannel != nil {
		if messages, ok := m.messagesByChannel[channelID]; ok {
			return messages, nil
		}
	}
	// Return empty slice if no specific messages configured for this channel
	return []entities.Message{}, nil
}

func (m *mockSlackRepo) GetUserDisplayName(ctx context.Context, userID string) (string, error) {
	return "", nil
}

type mockFileRepo struct {
	writeErr error
	// Track write calls for verification
	writeCalls int
}

func (m *mockFileRepo) WriteMessages(ctx context.Context, config valueobjects.ExportConfig, channel entities.Channel, messages map[string][]entities.Message) error {
	m.writeCalls++
	return m.writeErr
}

func (m *mockFileRepo) EnsureDirectoryExists(path string) error {
	return nil
}

func TestNewExportService(t *testing.T) {
	slackRepo := &mockSlackRepo{}
	fileRepo := &mockFileRepo{}
	log := logger.NewLogger(logger.INFO)

	service := NewExportService(slackRepo, fileRepo, log)
	if service == nil {
		t.Error("Expected service to be created")
	}
	if service.slackRepo == nil {
		t.Error("Expected slack repository to be set")
	}
	if service.fileRepo == nil {
		t.Error("Expected file repository to be set")
	}
	if service.logger == nil {
		t.Error("Expected logger to be set")
	}
}

func TestExportService_OrganizeMessagesByDate(t *testing.T) {
	slackRepo := &mockSlackRepo{}
	fileRepo := &mockFileRepo{}
	log := logger.NewLogger(logger.ERROR)
	service := NewExportService(slackRepo, fileRepo, log)

	// Create test messages
	t1 := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 15, 14, 0, 0, 0, time.UTC)
	t3 := time.Date(2024, 1, 16, 9, 0, 0, 0, time.UTC)

	msg1 := entities.NewMessage(t1, "Message 1", "User1")
	msg2 := entities.NewMessage(t2, "Message 2", "User2")
	msg3 := entities.NewMessage(t3, "Message 3", "User3")

	messages := []entities.Message{*msg1, *msg2, *msg3}

	organized := service.organizeMessagesByDate(messages)

	// Check that we have 2 dates
	if len(organized) != 2 {
		t.Errorf("Expected 2 dates, got %d", len(organized))
	}

	// Check 2024-01-15 has 2 messages
	date1 := "2024-01-15"
	if len(organized[date1]) != 2 {
		t.Errorf("Expected 2 messages for %s, got %d", date1, len(organized[date1]))
	}

	// Check 2024-01-16 has 1 message
	date2 := "2024-01-16"
	if len(organized[date2]) != 1 {
		t.Errorf("Expected 1 message for %s, got %d", date2, len(organized[date2]))
	}

	// Check messages are sorted by time within each date
	if !organized[date1][0].Timestamp.Before(organized[date1][1].Timestamp) {
		t.Error("Expected messages to be sorted by timestamp")
	}
}

func TestExportService_OrganizeMessagesByDate_ThreadHandling(t *testing.T) {
	slackRepo := &mockSlackRepo{}
	fileRepo := &mockFileRepo{}
	log := logger.NewLogger(logger.ERROR)
	service := NewExportService(slackRepo, fileRepo, log)

	// Create parent message and reply
	t1 := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 15, 10, 5, 0, 0, time.UTC)

	parent := entities.NewMessage(t1, "Parent message", "User1")
	parent.IsParent = true

	reply := entities.NewMessage(t2, "Reply message", "User2")
	reply.ThreadTS = "1234567890.123456"

	parent.AddReply(*reply)

	messages := []entities.Message{*parent, *reply}

	organized := service.organizeMessagesByDate(messages)

	// Only parent should be in the organized map (reply is attached to parent)
	date := "2024-01-15"
	if len(organized[date]) != 1 {
		t.Errorf("Expected 1 message (parent only), got %d", len(organized[date]))
	}

	// Verify the parent has the reply attached
	if len(organized[date][0].Replies) != 1 {
		t.Errorf("Expected parent to have 1 reply, got %d", len(organized[date][0].Replies))
	}
}

// Test discover Channels - channel filtering logic
func TestExportService_DiscoverChannels_NoFilter(t *testing.T) {
	ch1, _ := entities.NewChannel("C123", "general")
	ch2, _ := entities.NewChannel("C456", "random")
	ch3, _ := entities.NewChannel("C789", "dev")

	slackRepo := &mockSlackRepo{
		channels: []entities.Channel{*ch1, *ch2, *ch3},
	}
	fileRepo := &mockFileRepo{}
	log := logger.NewLogger(logger.ERROR)
	service := NewExportService(slackRepo, fileRepo, log)

	// No filter - should return all channels
	channels, err := service.discoverChannels(context.Background(), []string{})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(channels) != 3 {
		t.Errorf("Expected 3 channels, got %d", len(channels))
	}
}

func TestExportService_DiscoverChannels_WithFilter(t *testing.T) {
	ch1, _ := entities.NewChannel("C123", "general")
	ch2, _ := entities.NewChannel("C456", "random")
	ch3, _ := entities.NewChannel("C789", "dev")

	slackRepo := &mockSlackRepo{
		channels: []entities.Channel{*ch1, *ch2, *ch3},
	}
	fileRepo := &mockFileRepo{}
	log := logger.NewLogger(logger.ERROR)
	service := NewExportService(slackRepo, fileRepo, log)

	// Filter for 2 channels
	channels, err := service.discoverChannels(context.Background(), []string{"C123", "C456"})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(channels) != 2 {
		t.Errorf("Expected 2 channels, got %d", len(channels))
	}

	// Verify correct channels returned
	foundC123 := false
	foundC456 := false
	for _, ch := range channels {
		if ch.ID == "C123" {
			foundC123 = true
		}
		if ch.ID == "C456" {
			foundC456 = true
		}
	}
	if !foundC123 || !foundC456 {
		t.Error("Expected channels C123 and C456 to be returned")
	}
}

func TestExportService_DiscoverChannels_PartialMatch(t *testing.T) {
	ch1, _ := entities.NewChannel("C123", "general")
	ch2, _ := entities.NewChannel("C456", "random")

	slackRepo := &mockSlackRepo{
		channels: []entities.Channel{*ch1, *ch2},
	}
	fileRepo := &mockFileRepo{}
	log := logger.NewLogger(logger.ERROR)
	service := NewExportService(slackRepo, fileRepo, log)

	// Filter includes one existing and one non-existing channel
	channels, err := service.discoverChannels(context.Background(), []string{"C123", "C999"})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(channels) != 1 {
		t.Errorf("Expected 1 channel, got %d", len(channels))
	}
	if channels[0].ID != "C123" {
		t.Errorf("Expected channel C123, got %s", channels[0].ID)
	}
}

func TestExportService_DiscoverChannels_NoMatch(t *testing.T) {
	ch1, _ := entities.NewChannel("C123", "general")

	slackRepo := &mockSlackRepo{
		channels: []entities.Channel{*ch1},
	}
	fileRepo := &mockFileRepo{}
	log := logger.NewLogger(logger.ERROR)
	service := NewExportService(slackRepo, fileRepo, log)

	// Filter for non-existing channel
	channels, err := service.discoverChannels(context.Background(), []string{"C999"})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(channels) != 0 {
		t.Errorf("Expected 0 channels, got %d", len(channels))
	}
}

// Test exportChannel - single channel export
func TestExportService_ExportChannel_Success(t *testing.T) {
	ch, _ := entities.NewChannel("C123", "general")
	msg1 := entities.NewMessage(time.Now(), "Hello", "User1")

	slackRepo := &mockSlackRepo{
		messagesByChannel: map[string][]entities.Message{
			"C123": {*msg1},
		},
	}
	fileRepo := &mockFileRepo{}
	log := logger.NewLogger(logger.ERROR)
	service := NewExportService(slackRepo, fileRepo, log)

	timeRange, _ := valueobjects.NewTimeRangeFromDaysBack(7)
	config, _ := valueobjects.NewExportConfig("/tmp/export", *timeRange)

	result := service.exportChannel(context.Background(), *ch, *config)

	if !result.Success {
		t.Errorf("Expected success, got failure: %v", result.Error)
	}
	if result.MessageCount != 1 {
		t.Errorf("Expected 1 message, got %d", result.MessageCount)
	}
	if result.Channel.ID != "C123" {
		t.Errorf("Expected channel C123, got %s", result.Channel.ID)
	}
}

func TestExportService_ExportChannel_FetchMessagesError(t *testing.T) {
	ch, _ := entities.NewChannel("C123", "general")

	slackRepo := &mockSlackRepo{
		messagesErr: context.DeadlineExceeded,
	}
	fileRepo := &mockFileRepo{}
	log := logger.NewLogger(logger.ERROR)
	service := NewExportService(slackRepo, fileRepo, log)

	timeRange, _ := valueobjects.NewTimeRangeFromDaysBack(7)
	config, _ := valueobjects.NewExportConfig("/tmp/export", *timeRange)

	result := service.exportChannel(context.Background(), *ch, *config)

	if result.Success {
		t.Error("Expected failure, got success")
	}
	if result.Error == nil {
		t.Error("Expected error, got nil")
	}
}

func TestExportService_ExportChannel_WriteMessagesError(t *testing.T) {
	ch, _ := entities.NewChannel("C123", "general")
	msg1 := entities.NewMessage(time.Now(), "Hello", "User1")

	slackRepo := &mockSlackRepo{
		messagesByChannel: map[string][]entities.Message{
			"C123": {*msg1},
		},
	}
	fileRepo := &mockFileRepo{
		writeErr: context.DeadlineExceeded,
	}
	log := logger.NewLogger(logger.ERROR)
	service := NewExportService(slackRepo, fileRepo, log)

	timeRange, _ := valueobjects.NewTimeRangeFromDaysBack(7)
	config, _ := valueobjects.NewExportConfig("/tmp/export", *timeRange)

	result := service.exportChannel(context.Background(), *ch, *config)

	if result.Success {
		t.Error("Expected failure, got success")
	}
	if result.Error == nil {
		t.Error("Expected error, got nil")
	}
	// Should still have message count even if write failed
	if result.MessageCount != 1 {
		t.Errorf("Expected MessageCount = 1, got %d", result.MessageCount)
	}
}

func TestExportService_ExportChannel_EmptyMessages(t *testing.T) {
	ch, _ := entities.NewChannel("C123", "general")

	slackRepo := &mockSlackRepo{
		messagesByChannel: map[string][]entities.Message{
			"C123": {},
		},
	}
	fileRepo := &mockFileRepo{}
	log := logger.NewLogger(logger.ERROR)
	service := NewExportService(slackRepo, fileRepo, log)

	timeRange, _ := valueobjects.NewTimeRangeFromDaysBack(7)
	config, _ := valueobjects.NewExportConfig("/tmp/export", *timeRange)

	result := service.exportChannel(context.Background(), *ch, *config)

	// Empty messages should still be success
	if !result.Success {
		t.Errorf("Expected success with empty messages, got error: %v", result.Error)
	}
	if result.MessageCount != 0 {
		t.Errorf("Expected 0 messages, got %d", result.MessageCount)
	}
}

// Test ExportChannels - main entry point
func TestExportService_ExportChannels_Success_AllChannels(t *testing.T) {
	ch1, _ := entities.NewChannel("C123", "general")
	ch2, _ := entities.NewChannel("C456", "random")
	msg1 := entities.NewMessage(time.Now(), "Hello", "User1")
	msg2 := entities.NewMessage(time.Now(), "World", "User2")

	slackRepo := &mockSlackRepo{
		channels: []entities.Channel{*ch1, *ch2},
		messagesByChannel: map[string][]entities.Message{
			"C123": {*msg1},
			"C456": {*msg2},
		},
	}
	fileRepo := &mockFileRepo{}
	log := logger.NewLogger(logger.ERROR)
	service := NewExportService(slackRepo, fileRepo, log)

	input := dto.ExportChannelsInput{
		BotToken:   "xoxb-test",
		ChannelIDs: []string{}, // Empty = all channels
		ExportPath: "/tmp/export",
		DaysBack:   7,
	}

	output, err := service.ExportChannels(context.Background(), input)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if output == nil {
		t.Fatal("Expected output, got nil")
	}
	if output.Summary.TotalChannels != 2 {
		t.Errorf("Expected 2 total channels, got %d", output.Summary.TotalChannels)
	}
	if output.Summary.SuccessCount != 2 {
		t.Errorf("Expected 2 successes, got %d", output.Summary.SuccessCount)
	}
	if output.Summary.FailureCount != 0 {
		t.Errorf("Expected 0 failures, got %d", output.Summary.FailureCount)
	}
}

func TestExportService_ExportChannels_Success_FilteredChannels(t *testing.T) {
	ch1, _ := entities.NewChannel("C123", "general")
	ch2, _ := entities.NewChannel("C456", "random")
	ch3, _ := entities.NewChannel("C789", "dev")
	msg1 := entities.NewMessage(time.Now(), "Hello", "User1")
	msg2 := entities.NewMessage(time.Now(), "World", "User2")

	slackRepo := &mockSlackRepo{
		channels: []entities.Channel{*ch1, *ch2, *ch3},
		messagesByChannel: map[string][]entities.Message{
			"C123": {*msg1},
			"C456": {*msg2},
		},
	}
	fileRepo := &mockFileRepo{}
	log := logger.NewLogger(logger.ERROR)
	service := NewExportService(slackRepo, fileRepo, log)

	input := dto.ExportChannelsInput{
		BotToken:   "xoxb-test",
		ChannelIDs: []string{"C123", "C456"}, // Filter to 2 channels
		ExportPath: "/tmp/export",
		DaysBack:   7,
	}

	output, err := service.ExportChannels(context.Background(), input)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	// Should only export the 2 filtered channels
	if output.Summary.TotalChannels != 2 {
		t.Errorf("Expected 2 total channels, got %d", output.Summary.TotalChannels)
	}
	if output.Summary.SuccessCount != 2 {
		t.Errorf("Expected 2 successes, got %d", output.Summary.SuccessCount)
	}
}

func TestExportService_ExportChannels_Error_ChannelDiscoveryFails(t *testing.T) {
	slackRepo := &mockSlackRepo{
		channelsErr: context.DeadlineExceeded,
	}
	fileRepo := &mockFileRepo{}
	log := logger.NewLogger(logger.ERROR)
	service := NewExportService(slackRepo, fileRepo, log)

	input := dto.ExportChannelsInput{
		BotToken:   "xoxb-test",
		ChannelIDs: []string{},
		ExportPath: "/tmp/export",
		DaysBack:   7,
	}

	output, err := service.ExportChannels(context.Background(), input)

	if err == nil {
		t.Error("Expected error when channel discovery fails, got nil")
	}
	if output != nil {
		t.Error("Expected nil output on error, got non-nil")
	}
}

func TestExportService_ExportChannels_Error_InvalidDaysBack(t *testing.T) {
	slackRepo := &mockSlackRepo{}
	fileRepo := &mockFileRepo{}
	log := logger.NewLogger(logger.ERROR)
	service := NewExportService(slackRepo, fileRepo, log)

	input := dto.ExportChannelsInput{
		BotToken:   "xoxb-test",
		ChannelIDs: []string{},
		ExportPath: "/tmp/export",
		DaysBack:   -1, // Invalid
	}

	output, err := service.ExportChannels(context.Background(), input)

	if err == nil {
		t.Error("Expected error for negative daysBack, got nil")
	}
	if output != nil {
		t.Error("Expected nil output on error, got non-nil")
	}
}

func TestExportService_ExportChannels_Error_EmptyExportPath(t *testing.T) {
	ch1, _ := entities.NewChannel("C123", "general")

	slackRepo := &mockSlackRepo{
		channels: []entities.Channel{*ch1},
	}
	fileRepo := &mockFileRepo{}
	log := logger.NewLogger(logger.ERROR)
	service := NewExportService(slackRepo, fileRepo, log)

	input := dto.ExportChannelsInput{
		BotToken:   "xoxb-test",
		ChannelIDs: []string{},
		ExportPath: "", // Empty
		DaysBack:   7,
	}

	output, err := service.ExportChannels(context.Background(), input)

	if err == nil {
		t.Error("Expected error for empty export path, got nil")
	}
	if output != nil {
		t.Error("Expected nil output on error, got non-nil")
	}
}

func TestExportService_ExportChannels_Error_NoChannels(t *testing.T) {
	slackRepo := &mockSlackRepo{
		channels: []entities.Channel{}, // Empty channel list
	}
	fileRepo := &mockFileRepo{}
	log := logger.NewLogger(logger.ERROR)
	service := NewExportService(slackRepo, fileRepo, log)

	input := dto.ExportChannelsInput{
		BotToken:   "xoxb-test",
		ChannelIDs: []string{},
		ExportPath: "/tmp/export",
		DaysBack:   7,
	}

	output, err := service.ExportChannels(context.Background(), input)

	if err == nil {
		t.Error("Expected error for no channels, got nil")
	}
	if output != nil {
		t.Error("Expected nil output on error, got non-nil")
	}
}

func TestExportService_ExportChannels_PartialSuccess(t *testing.T) {
	ch1, _ := entities.NewChannel("C123", "general")
	ch2, _ := entities.NewChannel("C456", "random")
	msg1 := entities.NewMessage(time.Now(), "Hello", "User1")

	slackRepo := &mockSlackRepo{
		channels: []entities.Channel{*ch1, *ch2},
		messagesByChannel: map[string][]entities.Message{
			"C123": {*msg1},
			// C456 will get empty messages (default behavior)
		},
	}
	fileRepo := &mockFileRepo{
		writeErr: context.DeadlineExceeded, // All writes will fail
	}
	log := logger.NewLogger(logger.ERROR)
	service := NewExportService(slackRepo, fileRepo, log)

	input := dto.ExportChannelsInput{
		BotToken:   "xoxb-test",
		ChannelIDs: []string{},
		ExportPath: "/tmp/export",
		DaysBack:   7,
	}

	output, err := service.ExportChannels(context.Background(), input)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	// Both channels should fail due to write error
	if output.Summary.TotalChannels != 2 {
		t.Errorf("Expected 2 total channels, got %d", output.Summary.TotalChannels)
	}
	if output.Summary.FailureCount != 2 {
		t.Errorf("Expected 2 failures, got %d", output.Summary.FailureCount)
	}
	if output.Summary.SuccessCount != 0 {
		t.Errorf("Expected 0 successes, got %d", output.Summary.SuccessCount)
	}
}

func TestExportService_ExportChannels_EmptyMessages(t *testing.T) {
	ch1, _ := entities.NewChannel("C123", "general")

	slackRepo := &mockSlackRepo{
		channels: []entities.Channel{*ch1},
		messagesByChannel: map[string][]entities.Message{
			"C123": {}, // Empty messages
		},
	}
	fileRepo := &mockFileRepo{}
	log := logger.NewLogger(logger.ERROR)
	service := NewExportService(slackRepo, fileRepo, log)

	input := dto.ExportChannelsInput{
		BotToken:   "xoxb-test",
		ChannelIDs: []string{},
		ExportPath: "/tmp/export",
		DaysBack:   7,
	}

	output, err := service.ExportChannels(context.Background(), input)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	// Empty messages should still count as success
	if output.Summary.SuccessCount != 1 {
		t.Errorf("Expected 1 success, got %d", output.Summary.SuccessCount)
	}
	if output.Summary.FailureCount != 0 {
		t.Errorf("Expected 0 failures, got %d", output.Summary.FailureCount)
	}
}
