package filesystem

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/cppcho/slack2md/internal/domain/entities"
	"github.com/cppcho/slack2md/internal/domain/valueobjects"
	"github.com/cppcho/slack2md/internal/usecase/interfaces"
)

// Repository implements the interfaces.FileRepository interface
type Repository struct {
	writer FileWriter
	logger interfaces.Logger
}

// NewFileRepository creates a new Repository
func NewFileRepository(writer FileWriter, logger interfaces.Logger) *Repository {
	return &Repository{
		writer: writer,
		logger: logger,
	}
}

// WriteMessages writes messages to markdown files organized by date
func (r *Repository) WriteMessages(ctx context.Context, config valueobjects.ExportConfig, channel entities.Channel, messagesByDate map[string][]entities.Message) error {
	// Sanitize channel name
	sanitizedName := r.sanitizeChannelName(channel.Name)

	// Create channel directory
	channelDir := filepath.Join(config.ExportPath, sanitizedName)
	if err := r.EnsureDirectoryExists(channelDir); err != nil {
		r.logger.Error("Failed to create channel directory %s: %v", channelDir, err)
		return fmt.Errorf("failed to create channel directory: %w", err)
	}

	r.logger.Debug("Created directory: %s", channelDir)

	// Get sorted dates
	dates := r.getSortedDates(messagesByDate)

	// Write file for each date
	for _, date := range dates {
		messages := messagesByDate[date]
		if len(messages) == 0 {
			continue
		}

		fileName := fmt.Sprintf("%s.md", date)
		filePath := filepath.Join(channelDir, fileName)

		r.logger.Debug("Writing messages to %s", filePath)

		// Parse the date for parent date reference
		parentDate, err := time.Parse("2006-01-02", date)
		if err != nil {
			r.logger.Error("Failed to parse date %s: %v", date, err)
			return fmt.Errorf("failed to parse date %s: %w", date, err)
		}

		if err := r.writeMarkdownFile(filePath, messages, date, parentDate); err != nil {
			r.logger.Error("Failed to write file %s: %v", filePath, err)
			return fmt.Errorf("failed to write file for date %s: %w", date, err)
		}

		r.logger.Info("Wrote %d messages to %s", len(messages), filePath)
	}

	return nil
}

// EnsureDirectoryExists creates a directory if it doesn't exist
func (r *Repository) EnsureDirectoryExists(path string) error {
	return r.writer.MkdirAll(path, 0755)
}

// writeMarkdownFile writes messages to a markdown file for a specific date
func (r *Repository) writeMarkdownFile(filePath string, messages []entities.Message, date string, parentDate time.Time) error {
	// Build markdown content
	var content strings.Builder

	// Write header
	content.WriteString(fmt.Sprintf("# %s\n\n", date))

	// Write messages
	for _, msg := range messages {
		r.writeMessage(&content, msg, parentDate)
	}

	// Write to file
	return r.writer.WriteFile(filePath, []byte(content.String()))
}

// writeMessage writes a single message and its thread replies
func (r *Repository) writeMessage(buf *strings.Builder, msg entities.Message, parentDate time.Time) {
	// Write parent message with display name
	timestamp := msg.Timestamp.Format("15:04")
	buf.WriteString(fmt.Sprintf("### %s %s\n%s\n\n", timestamp, msg.UserDisplayName, msg.Text))

	// Write thread replies if any
	for _, reply := range msg.Replies {
		r.writeThreadReply(buf, reply, msg.Timestamp)
	}
}

// writeThreadReply writes a thread reply message
func (r *Repository) writeThreadReply(buf *strings.Builder, reply entities.Message, parentTime time.Time) {
	// Check if reply is on the same day as parent
	sameDay := r.isSameDay(reply.Timestamp, parentTime)

	var timestamp string
	if sameDay {
		// Same day: just show time
		timestamp = reply.Timestamp.Format("15:04")
	} else {
		// Different day: show full date and time
		timestamp = reply.Timestamp.Format("2006-01-02 15:04")
	}

	buf.WriteString(fmt.Sprintf("###### %s %s\n%s\n\n", timestamp, reply.UserDisplayName, reply.Text))
}

// isSameDay checks if two timestamps are on the same day
func (r *Repository) isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// sanitizeChannelName sanitizes a channel name for use as a directory name
func (r *Repository) sanitizeChannelName(name string) string {
	// Replace any character that's not alphanumeric, hyphen, or underscore with underscore
	re := regexp.MustCompile(`[^a-zA-Z0-9\-_]`)
	return re.ReplaceAllString(name, "_")
}

// getSortedDates returns a sorted slice of date strings from the map
func (r *Repository) getSortedDates(messagesByDate map[string][]entities.Message) []string {
	dates := make([]string, 0, len(messagesByDate))
	for date := range messagesByDate {
		dates = append(dates, date)
	}

	// Sort dates chronologically (alphabetically works for YYYY-MM-DD format)
	r.sortDates(dates)

	return dates
}

// sortDates sorts date strings in chronological order
func (r *Repository) sortDates(dates []string) {
	// Simple bubble sort since we likely don't have many dates
	for i := 0; i < len(dates); i++ {
		for j := i + 1; j < len(dates); j++ {
			if strings.Compare(dates[i], dates[j]) > 0 {
				dates[i], dates[j] = dates[j], dates[i]
			}
		}
	}
}
