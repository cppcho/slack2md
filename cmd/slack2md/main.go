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
	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		common.Error(fmt.Sprintf("Configuration error: %v", err))
		os.Exit(1)
	}

	// Initialize Slack client
	client := slack.NewClient(config.BotToken, config.AppToken)

	// Auto-discover channels if not manually specified
	channels, err := discoverChannels(client)
	if err != nil {
		common.Error(fmt.Sprintf("Channel discovery failed: %v", err))
		os.Exit(1)
	}

	if len(config.ChannelIDs) > 0 {
		// filter out discovered channels to only those specified
		var filtered []slack.Channel
		channelIDSet := make(map[string]struct{})
		for _, id := range config.ChannelIDs {
			channelIDSet[id] = struct{}{}
		}
		for _, ch := range channels {
			if _, exists := channelIDSet[ch.ID]; exists {
				filtered = append(filtered, ch)
			}
		}
		channels = filtered
	}

	// Validate we have channels to export
	if len(channels) == 0 {
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
	for _, channel := range channels {
		fmt.Printf("Processing channel %s...\n", channel.Name)

		if err := processChannel(client, channel, startTime, endTime, config.ExportPath); err != nil {
			common.Error(fmt.Sprintf("Failed to process channel %s: %v", channel.Name, err))
			failureCount++
		} else {
			common.Success(fmt.Sprintf("Successfully exported channel %s", channel.Name))
			successCount++
		}

		fmt.Println()
	}

	// Print summary
	fmt.Printf("\n--- Summary ---\n")
	fmt.Printf("Total channels: %d\n", len(channels))
	fmt.Printf("Successful exports: %d\n", successCount)
	fmt.Printf("Failed exports: %d\n", failureCount)

	if failureCount > 0 {
		common.Error("Some channels failed to export")
		os.Exit(1)
	}

	common.Success("All channels exported successfully!")
}

// discoverChannels auto-discovers channels if not manually specified
func discoverChannels(client *slack.Client) ([]slack.Channel, error) {
	// Auto-discover channels
	fmt.Println("Auto-discovering channels...")
	channels, err := client.FetchAllChannels()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch channels: %w", err)
	}

	if len(channels) == 0 {
		return nil, fmt.Errorf("no channels found - bot may not be invited to any channels")
	}

	fmt.Printf("Found %d channel(s):\n", len(channels))
	for _, ch := range channels {
		fmt.Printf("  - %s (%s)\n", ch.Name, ch.ID)
	}
	fmt.Println()

	return channels, nil
}

// processChannel handles the export for a single channel
func processChannel(client *slack.Client, channel slack.Channel, startTime, endTime time.Time, exportPath string) error {
	// Fetch messages from Slack
	fmt.Printf("  Fetching messages from Slack...\n")
	result, err := client.FetchChannelMessages(channel.ID, startTime, endTime)
	if err != nil {
		return fmt.Errorf("failed to fetch messages: %w", err)
	}

	fmt.Printf("  Channel name: %s\n", channel.Name)
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
			fwMessages = append(fwMessages, ConvertMessage(msg))
		}
		filewriterMessages[date] = fwMessages
	}

	// Write markdown files
	fmt.Printf("  Writing markdown files...\n")
	if err := filewriter.WriteChannelMessages(exportPath, channel.Name, filewriterMessages); err != nil {
		return fmt.Errorf("failed to write files: %w", err)
	}

	return nil
}

// convertMessage converts slack.Message to filewriter.Message
func ConvertMessage(msg slack.Message) filewriter.Message {
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
			fwMsg.Replies[i] = ConvertMessage(reply)
		}
	}

	return fwMsg
}
