package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cppcho/slack2md/internal/common"
	"github.com/cppcho/slack2md/internal/filewriter"
	"github.com/cppcho/slack2md/internal/slack"
)

func main() {
	common.PrintBanner("Slack Channel Export")

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		common.Error(fmt.Sprintf("Configuration error: %v", err))
		os.Exit(1)
	}

	// Initialize Slack client
	client := slack.NewClient(config.BotToken, config.AppToken)

	// Auto-discover channels if not manually specified
	if err := discoverChannels(client, config); err != nil {
		common.Error(fmt.Sprintf("Channel discovery failed: %v", err))
		os.Exit(1)
	}

	// Validate we have channels to export
	if len(config.ChannelIDs) == 0 {
		common.Error("No channels to export. Either provide SLACK_CHANNEL_IDS or ensure bot is invited to channels.")
		os.Exit(1)
	}

	// Calculate date range (last N days)
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -config.DaysBack)

	fmt.Printf("Exporting messages from last %d days (%s to %s)\n",
		config.DaysBack,
		startTime.Format("2006-01-02"),
		endTime.Format("2006-01-02"))
	fmt.Printf("Export path: %s\n\n", config.ExportPath)

	// Track success/failure counts
	successCount := 0
	failureCount := 0

	// Process each channel
	for _, channelID := range config.ChannelIDs {
		fmt.Printf("Processing channel %s...\n", channelID)

		if err := processChannel(client, channelID, startTime, endTime, config.ExportPath); err != nil {
			common.Error(fmt.Sprintf("Failed to process channel %s: %v", channelID, err))
			failureCount++
		} else {
			common.Success(fmt.Sprintf("Successfully exported channel %s", channelID))
			successCount++
		}

		fmt.Println()
	}

	// Print summary
	fmt.Printf("\n--- Summary ---\n")
	fmt.Printf("Total channels: %d\n", len(config.ChannelIDs))
	fmt.Printf("Successful exports: %d\n", successCount)
	fmt.Printf("Failed exports: %d\n", failureCount)

	if failureCount > 0 {
		common.Error("Some channels failed to export")
		os.Exit(1)
	}

	common.Success("All channels exported successfully!")
}

// discoverChannels auto-discovers channels if not manually specified
func discoverChannels(client *slack.Client, config *Config) error {
	// Skip discovery if channels are manually provided
	if len(config.ChannelIDs) > 0 {
		fmt.Printf("Using %d manually specified channel(s)\n\n", len(config.ChannelIDs))
		return nil
	}

	// Auto-discover channels
	fmt.Println("Auto-discovering channels...")
	channels, err := client.FetchAllChannels()
	if err != nil {
		return fmt.Errorf("failed to fetch channels: %w", err)
	}

	if len(channels) == 0 {
		return fmt.Errorf("no channels found - bot may not be invited to any channels")
	}

	fmt.Printf("Found %d channel(s):\n", len(channels))
	for _, ch := range channels {
		fmt.Printf("  - %s (%s)\n", ch.Name, ch.ID)
		config.ChannelIDs = append(config.ChannelIDs, ch.ID)
	}
	fmt.Println()

	return nil
}

// processChannel handles the export for a single channel
func processChannel(client *slack.Client, channelID string, startTime, endTime time.Time, exportPath string) error {
	// Fetch messages from Slack
	fmt.Printf("  Fetching messages from Slack...\n")
	result, err := client.FetchChannelMessages(channelID, startTime, endTime)
	if err != nil {
		return fmt.Errorf("failed to fetch messages: %w", err)
	}

	fmt.Printf("  Channel name: %s\n", result.ChannelName)
	fmt.Printf("  Fetched %d messages\n", len(result.Messages))

	// Organize messages by date
	fmt.Printf("  Organizing messages by date...\n")
	messagesByDate := slack.OrganizeMessagesByDate(result.Messages)

	fmt.Printf("  Messages span %d days\n", len(messagesByDate))

	// Convert slack.Message to filewriter.Message
	filewriterMessages := make(map[string][]filewriter.Message)
	for date, messages := range messagesByDate {
		var fwMessages []filewriter.Message
		for _, msg := range messages {
			fwMessages = append(fwMessages, convertMessage(msg))
		}
		filewriterMessages[date] = fwMessages
	}

	// Write markdown files
	fmt.Printf("  Writing markdown files...\n")
	if err := filewriter.WriteChannelMessages(exportPath, result.ChannelName, filewriterMessages); err != nil {
		return fmt.Errorf("failed to write files: %w", err)
	}

	return nil
}

// convertMessage converts slack.Message to filewriter.Message
func convertMessage(msg slack.Message) filewriter.Message {
	fwMsg := filewriter.Message{
		Timestamp: msg.Timestamp,
		ThreadTS:  msg.ThreadTS,
		Text:      msg.Text,
		IsParent:  msg.IsParent,
	}

	// Convert replies
	if len(msg.Replies) > 0 {
		fwMsg.Replies = make([]filewriter.Message, len(msg.Replies))
		for i, reply := range msg.Replies {
			fwMsg.Replies[i] = convertMessage(reply)
		}
	}

	return fwMsg
}
