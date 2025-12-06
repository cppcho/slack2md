package main

import (
	"errors"
	"os"
	"strings"
)

// Config holds the configuration for the Slack export tool
type Config struct {
	BotToken   string
	AppToken   string
	ChannelIDs []string
	ExportPath string
	DaysBack   int
}

// LoadConfig reads configuration from environment variables
func LoadConfig() (*Config, error) {
	config := &Config{
		DaysBack: 7, // Default: last 7 days
	}

	// Required: Bot token
	config.BotToken = os.Getenv("SLACK_BOT_TOKEN")
	if config.BotToken == "" {
		return nil, errors.New("SLACK_BOT_TOKEN environment variable is required")
	}

	// Required: App token
	config.AppToken = os.Getenv("SLACK_APP_TOKEN")

	// Optional: Channel IDs (comma-separated)
	// If not provided, auto-discovery will be used
	channelIDsStr := os.Getenv("SLACK_CHANNEL_IDS")
	if channelIDsStr != "" {
		// Parse comma-separated channel IDs
		channels := strings.Split(channelIDsStr, ",")
		for _, ch := range channels {
			trimmed := strings.TrimSpace(ch)
			if trimmed != "" {
				config.ChannelIDs = append(config.ChannelIDs, trimmed)
			}
		}
	}
	// Note: Empty ChannelIDs is valid - will trigger auto-discovery in main()

	// Required: Export path
	config.ExportPath = os.Getenv("SLACK_EXPORT_PATH")
	if config.ExportPath == "" {
		return nil, errors.New("SLACK_EXPORT_PATH environment variable is required")
	}

	return config, nil
}
