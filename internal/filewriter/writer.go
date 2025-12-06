package filewriter

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Message represents a message to be written (matches slack.Message structure)
type Message struct {
	Timestamp       time.Time
	ThreadTS        string
	Text            string
	UserDisplayName string // Display name of the user who sent the message
	Replies         []Message
	IsParent        bool
}

// writeMarkdownFile writes messages to a markdown file for a specific date
func writeMarkdownFile(filePath string, messages []Message, date string, parentDate time.Time) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Open file for writing (create or truncate)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()

	// Write header
	fmt.Fprintf(file, "# %s\n\n", date)

	// Write messages
	for _, msg := range messages {
		if err := writeMessage(file, msg, parentDate); err != nil {
			return err
		}
	}

	return nil
}

// writeMessage writes a single message and its thread replies
func writeMessage(file *os.File, msg Message, parentDate time.Time) error {
	// Write parent message with display name
	timestamp := msg.Timestamp.Format("15:04")
	fmt.Fprintf(file, "### %s %s\n%s\n", timestamp, msg.UserDisplayName, msg.Text)
	fmt.Fprintln(file)

	// Write thread replies if any
	for _, reply := range msg.Replies {
		if err := writeThreadReply(file, reply, msg.Timestamp); err != nil {
			return err
		}
	}

	return nil
}

// writeThreadReply writes a thread reply message
func writeThreadReply(file *os.File, reply Message, parentTime time.Time) error {
	// Check if reply is on the same day as parent
	sameDay := isSameDay(reply.Timestamp, parentTime)

	var timestamp string
	if sameDay {
		// Same day: just show time
		timestamp = reply.Timestamp.Format("15:04")
	} else {
		// Different day: show full date and time
		timestamp = reply.Timestamp.Format("2006-01-02 15:04")
	}

	fmt.Fprintf(file, "###### %s %s\n%s\n", timestamp, reply.UserDisplayName, reply.Text)
	fmt.Fprintln(file)

	return nil
}

// isSameDay checks if two timestamps are on the same day
func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// sanitizeChannelName sanitizes a channel name for use as a directory name
func sanitizeChannelName(name string) string {
	// Replace any character that's not alphanumeric, hyphen, or underscore with underscore
	re := regexp.MustCompile(`[^a-zA-Z0-9\-_]`)
	return re.ReplaceAllString(name, "_")
}

// WriteChannelMessages writes all messages for a channel, organized by date
func WriteChannelMessages(exportPath, channelName string, messagesByDate map[string][]Message) error {
	// Sanitize channel name
	sanitizedName := sanitizeChannelName(channelName)

	// Create channel directory
	channelDir := filepath.Join(exportPath, sanitizedName)
	if err := os.MkdirAll(channelDir, 0755); err != nil {
		return fmt.Errorf("failed to create channel directory: %w", err)
	}

	// Get sorted dates
	dates := getSortedDates(messagesByDate)

	// Write file for each date
	for _, date := range dates {
		messages := messagesByDate[date]
		if len(messages) == 0 {
			continue
		}

		fileName := fmt.Sprintf("%s.md", date)
		filePath := filepath.Join(channelDir, fileName)

		// Parse the date for parent date reference
		parentDate, err := time.Parse("2006-01-02", date)
		if err != nil {
			return fmt.Errorf("failed to parse date %s: %w", date, err)
		}

		if err := writeMarkdownFile(filePath, messages, date, parentDate); err != nil {
			return fmt.Errorf("failed to write file for date %s: %w", date, err)
		}
	}

	return nil
}

// getSortedDates returns a sorted slice of date strings from the map
func getSortedDates(messagesByDate map[string][]Message) []string {
	dates := make([]string, 0, len(messagesByDate))
	for date := range messagesByDate {
		dates = append(dates, date)
	}

	// Sort dates chronologically
	sortDates(dates)

	return dates
}

// sortDates sorts date strings in chronological order
func sortDates(dates []string) {
	// Simple bubble sort since we likely don't have many dates
	for i := 0; i < len(dates); i++ {
		for j := i + 1; j < len(dates); j++ {
			if strings.Compare(dates[i], dates[j]) > 0 {
				dates[i], dates[j] = dates[j], dates[i]
			}
		}
	}
}
