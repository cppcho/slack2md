package slack

import (
	"testing"

	"github.com/cppcho/slack2md/internal/adapter/logger"
)

// MockSlackClient for testing
type MockSlackClient struct{}

func (m *MockSlackClient) GetConversationsForUser(params interface{}) ([]interface{}, string, error) {
	return nil, "", nil
}

func (m *MockSlackClient) GetConversationHistory(params interface{}) (interface{}, error) {
	return nil, nil
}

func (m *MockSlackClient) GetConversationReplies(params interface{}) ([]interface{}, bool, string, error) {
	return nil, false, "", nil
}

func (m *MockSlackClient) GetUserInfo(userID string) (interface{}, error) {
	return nil, nil
}

func TestNewSlackRepository(t *testing.T) {
	client := NewSlackClient("xoxb-test", "")
	log := logger.NewLogger(logger.INFO)
	cache := NewUserCache()

	repo := NewSlackRepository(client, log, cache)
	if repo == nil {
		t.Error("Expected repository to be created")
	}
	if repo.client == nil {
		t.Error("Expected client to be set")
	}
	if repo.formatter == nil {
		t.Error("Expected formatter to be created internally")
	}
	if repo.logger == nil {
		t.Error("Expected logger to be set")
	}
	if repo.userCache == nil {
		t.Error("Expected user cache to be set")
	}
}

func TestSlackRepository_ParseSlackTimestamp(t *testing.T) {
	client := NewSlackClient("xoxb-test", "")
	log := logger.NewLogger(logger.ERROR) // Suppress logs during test
	cache := NewUserCache()
	repo := NewSlackRepository(client, log, cache)

	tests := []struct {
		name        string
		timestamp   string
		wantErr     bool
		wantUnixSec int64
	}{
		{
			name:        "valid timestamp",
			timestamp:   "1234567890.123456",
			wantErr:     false,
			wantUnixSec: 1234567890,
		},
		{
			name:      "invalid timestamp - no decimal",
			timestamp: "1234567890",
			wantErr:   true,
		},
		{
			name:      "invalid timestamp - text",
			timestamp: "invalid",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, err := repo.parseSlackTimestamp(tt.timestamp)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSlackTimestamp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && ts.Unix() != tt.wantUnixSec {
				t.Errorf("parseSlackTimestamp() Unix = %v, want %v", ts.Unix(), tt.wantUnixSec)
			}
		})
	}
}
