package slack

import (
	"context"
	"fmt"
	"time"

	"github.com/cppcho/slack2md/internal/domain/entities"
	"github.com/cppcho/slack2md/internal/domain/valueobjects"
	"github.com/cppcho/slack2md/internal/usecase/interfaces"
	slackapi "github.com/slack-go/slack"
)

// SlackRepositoryImpl implements the SlackRepository interface
type SlackRepositoryImpl struct {
	client    SlackClient
	formatter *MarkdownFormatter
	logger    interfaces.Logger
	userCache *UserCache
}

// NewSlackRepository creates a new SlackRepositoryImpl
func NewSlackRepository(
	client SlackClient,
	logger interfaces.Logger,
	userCache *UserCache,
) *SlackRepositoryImpl {
	return &SlackRepositoryImpl{
		client:    client,
		formatter: NewMarkdownFormatter(), // Created internally, not injected
		logger:    logger,
		userCache: userCache,
	}
}

// FetchAllChannels fetches all channels where the bot is a member
func (r *SlackRepositoryImpl) FetchAllChannels(ctx context.Context) ([]entities.Channel, error) {
	r.logger.Info("Fetching channels for user")

	var allChannels []entities.Channel
	cursor := ""

	for {
		// Get conversations with pagination
		params := &slackapi.GetConversationsForUserParameters{
			Cursor:          cursor,
			ExcludeArchived: true,
			Limit:           200,
			Types:           []string{"public_channel", "private_channel"},
		}

		channels, nextCursor, err := r.client.GetConversationsForUser(params)
		if err != nil {
			r.logger.Error("Failed to fetch channels: %v", err)
			return nil, fmt.Errorf("failed to fetch conversations: %w", err)
		}

		// Convert to our Channel entities
		for _, ch := range channels {
			channel, err := entities.NewChannel(ch.ID, ch.Name)
			if err != nil {
				r.logger.Warn("Skipping invalid channel: %v", err)
				continue
			}
			allChannels = append(allChannels, *channel)
		}

		r.logger.Debug("Fetched batch of %d channels, cursor: %s", len(channels), nextCursor)

		// Check if there are more pages
		if nextCursor == "" {
			break
		}
		cursor = nextCursor
	}

	r.logger.Info("Found %d channels", len(allChannels))
	return allChannels, nil
}

// FetchChannelMessages fetches all messages from a channel within the specified time range
func (r *SlackRepositoryImpl) FetchChannelMessages(ctx context.Context, channelID string, timeRange valueobjects.TimeRange) ([]entities.Message, error) {
	r.logger.Info("Fetching messages for channel %s", channelID)

	var allMessages []entities.Message
	cursor := ""
	oldest := fmt.Sprintf("%d", timeRange.Start.Unix())
	latest := fmt.Sprintf("%d", timeRange.End.Unix())

	for {
		params := &slackapi.GetConversationHistoryParameters{
			ChannelID: channelID,
			Oldest:    oldest,
			Latest:    latest,
			Limit:     200,
			Cursor:    cursor,
		}

		history, err := r.client.GetConversationHistory(params)
		if err != nil {
			r.logger.Error("Failed to fetch conversation history for channel %s: %v", channelID, err)
			return nil, fmt.Errorf("failed to fetch conversation history: %w", err)
		}

		r.logger.Debug("Fetched batch of %d messages, cursor: %s", len(history.Messages), cursor)

		// Process messages
		for _, msg := range history.Messages {
			// Skip messages with no text (e.g., attachment-only messages)
			if msg.Text == "" {
				continue
			}

			// Parse timestamp
			ts, err := r.parseSlackTimestamp(msg.Timestamp)
			if err != nil {
				r.logger.Warn("Skipping message with invalid timestamp: %s", msg.Timestamp)
				continue
			}

			// Get user display name
			displayName, err := r.GetUserDisplayName(ctx, msg.User)
			if err != nil {
				r.logger.Warn("Failed to get display name for user %s: %v", msg.User, err)
				displayName = msg.User // Fallback to user ID
			}

			// Convert Slack format to Markdown
			formattedText := r.formatter.ConvertSlackToMarkdown(msg.Text)

			message := entities.NewMessage(ts, formattedText, displayName)
			message.ThreadTS = msg.ThreadTimestamp
			message.IsParent = msg.ReplyCount > 0

			// If this is a parent message with replies, fetch the thread
			if message.IsParent && msg.ThreadTimestamp == msg.Timestamp {
				replies, err := r.fetchThreadReplies(ctx, channelID, msg.Timestamp)
				if err != nil {
					r.logger.Warn("Failed to fetch thread replies for message %s: %v", msg.Timestamp, err)
				} else {
					message.Replies = replies
				}
			}

			allMessages = append(allMessages, *message)
		}

		// Check if there are more messages
		if !history.HasMore {
			break
		}

		cursor = history.ResponseMetaData.NextCursor
	}

	r.logger.Info("Fetched %d messages from channel %s", len(allMessages), channelID)
	return allMessages, nil
}

// GetUserDisplayName fetches user display name with caching
func (r *SlackRepositoryImpl) GetUserDisplayName(ctx context.Context, userID string) (string, error) {
	// Check cache first
	if displayName, ok := r.userCache.Get(userID); ok {
		return displayName, nil
	}

	// Handle empty user IDs
	if userID == "" {
		return "Unknown", nil
	}

	// Fetch from API
	user, err := r.client.GetUserInfo(userID)
	if err != nil {
		r.logger.Debug("Failed to fetch user info for %s: %v", userID, err)
		return userID, err
	}

	// Prefer display name, fall back to real name, then username
	displayName := userID
	if user.Profile.DisplayName != "" {
		displayName = user.Profile.DisplayName
	} else if user.RealName != "" {
		displayName = user.RealName
	} else if user.Name != "" {
		displayName = user.Name
	}

	// Cache the result
	r.userCache.Set(userID, displayName)
	r.logger.Debug("Cached display name for user %s: %s", userID, displayName)

	return displayName, nil
}

// fetchThreadReplies fetches all replies in a thread
func (r *SlackRepositoryImpl) fetchThreadReplies(ctx context.Context, channelID, threadTS string) ([]entities.Message, error) {
	r.logger.Debug("Fetching thread replies for message %s in channel %s", threadTS, channelID)

	params := &slackapi.GetConversationRepliesParameters{
		ChannelID: channelID,
		Timestamp: threadTS,
	}

	msgs, _, _, err := r.client.GetConversationReplies(params)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch thread replies: %w", err)
	}

	var replies []entities.Message

	// Skip the first message (it's the parent, which we already have)
	for i, msg := range msgs {
		if i == 0 {
			continue // Skip parent message
		}

		// Skip messages with no text
		if msg.Text == "" {
			continue
		}

		ts, err := r.parseSlackTimestamp(msg.Timestamp)
		if err != nil {
			r.logger.Warn("Skipping reply with invalid timestamp: %s", msg.Timestamp)
			continue
		}

		// Get user display name
		displayName, err := r.GetUserDisplayName(ctx, msg.User)
		if err != nil {
			r.logger.Warn("Failed to get display name for user %s: %v", msg.User, err)
			displayName = msg.User
		}

		// Convert Slack format to Markdown
		formattedText := r.formatter.ConvertSlackToMarkdown(msg.Text)

		reply := entities.NewMessage(ts, formattedText, displayName)
		reply.ThreadTS = msg.ThreadTimestamp

		replies = append(replies, *reply)
	}

	r.logger.Debug("Fetched %d replies for thread %s", len(replies), threadTS)
	return replies, nil
}

// parseSlackTimestamp converts Slack's timestamp format to time.Time
// Slack timestamps are Unix timestamps with microseconds (e.g., "1234567890.123456")
func (r *SlackRepositoryImpl) parseSlackTimestamp(ts string) (time.Time, error) {
	var sec, nsec int64
	_, err := fmt.Sscanf(ts, "%d.%d", &sec, &nsec)
	if err != nil {
		return time.Time{}, err
	}

	// Convert microseconds to nanoseconds
	nsec = nsec * 1000

	return time.Unix(sec, nsec), nil
}
