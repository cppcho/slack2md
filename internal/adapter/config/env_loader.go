package config

import (
	"os"
	"strings"
)

// EnvConfig holds the configuration loaded from environment variables
type EnvConfig struct {
	BotToken   string
	AppToken   string
	ChannelIDs []string
	ExportPath string
	DaysBack   int
}

// LoadFromEnv reads configuration from environment variables
func LoadFromEnv() (*EnvConfig, error) {
	config := &EnvConfig{
		DaysBack: 7, // Default: last 7 days
	}

	// Bot token
	config.BotToken = os.Getenv("SLACK_BOT_TOKEN")

	// App token (optional)
	config.AppToken = os.Getenv("SLACK_APP_TOKEN")

	// Channel IDs (optional, comma-separated)
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

	// Export path
	config.ExportPath = os.Getenv("SLACK_EXPORT_PATH")

	return config, nil
}
