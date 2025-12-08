package config

import (
	"os"
	"testing"
)

func TestLoadFromEnv_AllFields(t *testing.T) {
	// Set environment variables
	os.Setenv("SLACK_BOT_TOKEN", "xoxb-test-token")
	os.Setenv("SLACK_APP_TOKEN", "xapp-test-token")
	os.Setenv("SLACK_CHANNEL_IDS", "C123,C456,C789")
	os.Setenv("SLACK_EXPORT_PATH", "/tmp/export")
	defer func() {
		os.Unsetenv("SLACK_BOT_TOKEN")
		os.Unsetenv("SLACK_APP_TOKEN")
		os.Unsetenv("SLACK_CHANNEL_IDS")
		os.Unsetenv("SLACK_EXPORT_PATH")
	}()

	config, err := LoadFromEnv()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if config == nil {
		t.Fatal("Expected config, got nil")
	}
	if config.BotToken != "xoxb-test-token" {
		t.Errorf("Expected BotToken = %q, got %q", "xoxb-test-token", config.BotToken)
	}
	if config.AppToken != "xapp-test-token" {
		t.Errorf("Expected AppToken = %q, got %q", "xapp-test-token", config.AppToken)
	}
	if len(config.ChannelIDs) != 3 {
		t.Fatalf("Expected 3 channel IDs, got %d", len(config.ChannelIDs))
	}
	if config.ChannelIDs[0] != "C123" {
		t.Errorf("Expected ChannelIDs[0] = %q, got %q", "C123", config.ChannelIDs[0])
	}
	if config.ChannelIDs[1] != "C456" {
		t.Errorf("Expected ChannelIDs[1] = %q, got %q", "C456", config.ChannelIDs[1])
	}
	if config.ChannelIDs[2] != "C789" {
		t.Errorf("Expected ChannelIDs[2] = %q, got %q", "C789", config.ChannelIDs[2])
	}
	if config.ExportPath != "/tmp/export" {
		t.Errorf("Expected ExportPath = %q, got %q", "/tmp/export", config.ExportPath)
	}
	if config.DaysBack != 7 {
		t.Errorf("Expected DaysBack = 7, got %d", config.DaysBack)
	}
}

func TestLoadFromEnv_DefaultDaysBack(t *testing.T) {
	config, err := LoadFromEnv()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if config.DaysBack != 7 {
		t.Errorf("Expected default DaysBack = 7, got %d", config.DaysBack)
	}
}

func TestLoadFromEnv_EmptyChannelIDs(t *testing.T) {
	os.Setenv("SLACK_CHANNEL_IDS", "")
	defer os.Unsetenv("SLACK_CHANNEL_IDS")

	config, err := LoadFromEnv()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if config.ChannelIDs != nil && len(config.ChannelIDs) != 0 {
		t.Errorf("Expected empty ChannelIDs, got %v", config.ChannelIDs)
	}
}

func TestLoadFromEnv_ChannelIDsWithWhitespace(t *testing.T) {
	os.Setenv("SLACK_CHANNEL_IDS", " C123 , C456 , C789 ")
	defer os.Unsetenv("SLACK_CHANNEL_IDS")

	config, err := LoadFromEnv()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(config.ChannelIDs) != 3 {
		t.Fatalf("Expected 3 channel IDs, got %d", len(config.ChannelIDs))
	}
	// Verify whitespace is trimmed
	if config.ChannelIDs[0] != "C123" {
		t.Errorf("Expected trimmed ChannelIDs[0] = %q, got %q", "C123", config.ChannelIDs[0])
	}
	if config.ChannelIDs[1] != "C456" {
		t.Errorf("Expected trimmed ChannelIDs[1] = %q, got %q", "C456", config.ChannelIDs[1])
	}
	if config.ChannelIDs[2] != "C789" {
		t.Errorf("Expected trimmed ChannelIDs[2] = %q, got %q", "C789", config.ChannelIDs[2])
	}
}

func TestLoadFromEnv_ChannelIDsWithEmptyEntries(t *testing.T) {
	// Empty entries should be filtered out
	os.Setenv("SLACK_CHANNEL_IDS", "C123,,C456,  ,C789")
	defer os.Unsetenv("SLACK_CHANNEL_IDS")

	config, err := LoadFromEnv()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	// Should only have 3 valid channel IDs (empty entries filtered out)
	if len(config.ChannelIDs) != 3 {
		t.Errorf("Expected 3 channel IDs (empty filtered), got %d", len(config.ChannelIDs))
	}
}

func TestLoadFromEnv_SingleChannelID(t *testing.T) {
	os.Setenv("SLACK_CHANNEL_IDS", "C123")
	defer os.Unsetenv("SLACK_CHANNEL_IDS")

	config, err := LoadFromEnv()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(config.ChannelIDs) != 1 {
		t.Fatalf("Expected 1 channel ID, got %d", len(config.ChannelIDs))
	}
	if config.ChannelIDs[0] != "C123" {
		t.Errorf("Expected ChannelIDs[0] = %q, got %q", "C123", config.ChannelIDs[0])
	}
}

func TestLoadFromEnv_OptionalAppToken(t *testing.T) {
	// App token is optional
	config, err := LoadFromEnv()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if config.AppToken != "" {
		t.Errorf("Expected empty AppToken when not set, got %q", config.AppToken)
	}
}

func TestLoadFromEnv_MinimalConfig(t *testing.T) {
	// Only required fields
	os.Setenv("SLACK_BOT_TOKEN", "xoxb-minimal")
	os.Setenv("SLACK_EXPORT_PATH", "/tmp/minimal")
	defer func() {
		os.Unsetenv("SLACK_BOT_TOKEN")
		os.Unsetenv("SLACK_EXPORT_PATH")
	}()

	config, err := LoadFromEnv()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if config.BotToken != "xoxb-minimal" {
		t.Errorf("Expected BotToken = %q, got %q", "xoxb-minimal", config.BotToken)
	}
	if config.ExportPath != "/tmp/minimal" {
		t.Errorf("Expected ExportPath = %q, got %q", "/tmp/minimal", config.ExportPath)
	}
	// Optional fields should be empty/default
	if config.AppToken != "" {
		t.Errorf("Expected empty AppToken, got %q", config.AppToken)
	}
	if len(config.ChannelIDs) != 0 {
		t.Errorf("Expected empty ChannelIDs, got %v", config.ChannelIDs)
	}
	if config.DaysBack != 7 {
		t.Errorf("Expected default DaysBack = 7, got %d", config.DaysBack)
	}
}

func TestLoadFromEnv_EmptyEnvironment(t *testing.T) {
	// Clear all environment variables
	os.Unsetenv("SLACK_BOT_TOKEN")
	os.Unsetenv("SLACK_APP_TOKEN")
	os.Unsetenv("SLACK_CHANNEL_IDS")
	os.Unsetenv("SLACK_EXPORT_PATH")

	config, err := LoadFromEnv()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if config == nil {
		t.Fatal("Expected config, got nil")
	}
	// Should return config with empty fields and default DaysBack
	if config.BotToken != "" {
		t.Errorf("Expected empty BotToken, got %q", config.BotToken)
	}
	if config.ExportPath != "" {
		t.Errorf("Expected empty ExportPath, got %q", config.ExportPath)
	}
	if config.DaysBack != 7 {
		t.Errorf("Expected default DaysBack = 7, got %d", config.DaysBack)
	}
}
