package service

import (
	"context"
	"testing"
	"time"

	"github.com/cppcho/slack2md/internal/adapter/logger"
	"github.com/cppcho/slack2md/internal/domain/entities"
	"github.com/cppcho/slack2md/internal/domain/valueobjects"
)

// Mock repositories for testing
type mockSlackRepo struct{}

func (m *mockSlackRepo) FetchAllChannels(ctx context.Context) ([]entities.Channel, error) {
	return nil, nil
}

func (m *mockSlackRepo) FetchChannelMessages(ctx context.Context, channelID string, timeRange valueobjects.TimeRange) ([]entities.Message, error) {
	return nil, nil
}

func (m *mockSlackRepo) GetUserDisplayName(ctx context.Context, userID string) (string, error) {
	return "", nil
}

type mockFileRepo struct{}

func (m *mockFileRepo) WriteMessages(ctx context.Context, config valueobjects.ExportConfig, channel entities.Channel, messages map[string][]entities.Message) error {
	return nil
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
