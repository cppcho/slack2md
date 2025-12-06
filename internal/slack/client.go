package slack

import (
	"fmt"

	"github.com/slack-go/slack"
)

// Client wraps the Slack API client
type Client struct {
	api SlackAPI
}

// NewClient creates a new Slack API client
func NewClient(token string, appLevelToken string) *Client {
	client := &Client{
		api: slack.New(token, slack.OptionAppLevelToken(appLevelToken)),
	}
	return client
}

// NewClientWithAPI creates a client with a custom API implementation (for testing)
func NewClientWithAPI(api SlackAPI) *Client {
	return &Client{
		api: api,
	}
}

// Channel represents a Slack channel
type Channel struct {
	ID   string
	Name string
}

// GetChannelName fetches the channel name for a given channel ID
func (c *Client) GetChannelName(channelID string) (string, error) {
	// Use conversations.info API to get channel information
	channel, err := c.api.GetConversationInfo(&slack.GetConversationInfoInput{
		ChannelID: channelID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get channel info for %s: %w", channelID, err)
	}

	return channel.Name, nil
}

// FetchAllChannels fetches all channels where the bot is a member
func (c *Client) FetchAllChannels() ([]Channel, error) {
	var allChannels []Channel
	cursor := ""

	for {
		// Get conversations with pagination
		params := &slack.GetConversationsParameters{
			Cursor:          cursor,
			ExcludeArchived: true,
			Limit:           200,
			Types:           []string{"public_channel", "private_channel"},
		}

		channels, nextCursor, err := c.api.GetConversations(params)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch conversations: %w", err)
		}

		// Convert to our Channel struct
		for _, ch := range channels {
			allChannels = append(allChannels, Channel{
				ID:   ch.ID,
				Name: ch.Name,
			})
		}

		// Check if there are more pages
		if nextCursor == "" {
			break
		}
		cursor = nextCursor
	}

	return allChannels, nil
}

// GetUserDisplayName fetches the display name for a given user ID
func (c *Client) GetUserDisplayName(userID string) string {
	// Handle bot users and empty user IDs
	if userID == "" {
		return "Unknown"
	}

	user, err := c.api.GetUserInfo(userID)
	if err != nil {
		// If we can't fetch user info, return the user ID
		return userID
	}

	// Prefer display name, fall back to real name, then username
	if user.Profile.DisplayName != "" {
		return user.Profile.DisplayName
	}
	if user.RealName != "" {
		return user.RealName
	}
	if user.Name != "" {
		return user.Name
	}

	return userID
}
