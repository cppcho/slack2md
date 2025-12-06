package slack

import (
	"fmt"
	"time"

	"github.com/slack-go/slack"
)

// Message represents a Slack message with its metadata
type Message struct {
	Timestamp       time.Time
	ThreadTS        string // Thread timestamp (parent message timestamp)
	Text            string
	UserDisplayName string // Display name of the user who sent the message
	Replies         []Message
	IsParent        bool // True if this message has replies
}

// FetchResult contains the fetched messages and channel information
type FetchResult struct {
	ChannelID string
	Messages  []Message
}

// FetchChannelMessages fetches all messages from a channel within the specified time range
func (c *Client) FetchChannelMessages(channelID string, startTime, endTime time.Time) (*FetchResult, error) {
	// Fetch messages with pagination
	var allMessages []Message
	cursor := ""
	oldest := fmt.Sprintf("%d", startTime.Unix())
	latest := fmt.Sprintf("%d", endTime.Unix())

	for {
		params := &slack.GetConversationHistoryParameters{
			ChannelID: channelID,
			Oldest:    oldest,
			Latest:    latest,
			Limit:     200,
			Cursor:    cursor,
		}

		history, err := c.api.GetConversationHistory(params)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch conversation history: %w", err)
		}

		// Process messages
		for _, msg := range history.Messages {
			// Skip messages with no text (e.g., attachment-only messages)
			if msg.Text == "" {
				continue
			}

			// Parse timestamp
			ts, err := parseSlackTimestamp(msg.Timestamp)
			if err != nil {
				continue // Skip messages with invalid timestamps
			}

			// Get user display name
			displayName := c.getUserDisplayName(msg.User)

			message := Message{
				Timestamp:       ts,
				ThreadTS:        msg.ThreadTimestamp,
				Text:            convertSlackToMarkdown(msg.Text),
				UserDisplayName: displayName,
				IsParent:        msg.ReplyCount > 0,
			}

			// If this is a parent message with replies, fetch the thread
			if message.IsParent && msg.ThreadTimestamp == msg.Timestamp {
				replies, err := c.fetchThreadReplies(channelID, msg.Timestamp)
				if err != nil {
					// Log error but continue - we still want the parent message
					fmt.Printf("Warning: failed to fetch thread replies for %s: %v\n", msg.Timestamp, err)
				} else {
					message.Replies = replies
				}
			}

			allMessages = append(allMessages, message)
		}

		// Check if there are more messages
		if !history.HasMore {
			break
		}

		cursor = history.ResponseMetaData.NextCursor
	}

	return &FetchResult{
		ChannelID: channelID,
		Messages:  allMessages,
	}, nil
}

// fetchThreadReplies fetches all replies in a thread
func (c *Client) fetchThreadReplies(channelID, threadTS string) ([]Message, error) {
	params := &slack.GetConversationRepliesParameters{
		ChannelID: channelID,
		Timestamp: threadTS,
	}

	msgs, _, _, err := c.api.GetConversationReplies(params)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch thread replies: %w", err)
	}

	var replies []Message

	// Skip the first message (it's the parent, which we already have)
	for i, msg := range msgs {
		if i == 0 {
			continue // Skip parent message
		}

		// Skip messages with no text
		if msg.Text == "" {
			continue
		}

		ts, err := parseSlackTimestamp(msg.Timestamp)
		if err != nil {
			continue
		}

		// Get user display name
		displayName := c.getUserDisplayName(msg.User)

		reply := Message{
			Timestamp:       ts,
			ThreadTS:        msg.ThreadTimestamp,
			Text:            convertSlackToMarkdown(msg.Text),
			UserDisplayName: displayName,
		}

		replies = append(replies, reply)
	}

	return replies, nil
}

// parseSlackTimestamp converts Slack's timestamp format to time.Time
// Slack timestamps are Unix timestamps with microseconds (e.g., "1234567890.123456")
func parseSlackTimestamp(ts string) (time.Time, error) {
	var sec, nsec int64
	_, err := fmt.Sscanf(ts, "%d.%d", &sec, &nsec)
	if err != nil {
		return time.Time{}, err
	}

	// Convert microseconds to nanoseconds
	nsec = nsec * 1000

	return time.Unix(sec, nsec), nil
}
