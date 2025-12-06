package slack

import (
	"github.com/slack-go/slack"
)

// SlackAPI defines the interface for Slack API operations.
// This allows mocking in tests without hitting real Slack API.
type SlackAPI interface {
	// GetConversationInfo retrieves information about a conversation
	GetConversationInfo(params *slack.GetConversationInfoInput) (*slack.Channel, error)

	// GetConversations retrieves a list of conversations
	GetConversations(params *slack.GetConversationsParameters) ([]slack.Channel, string, error)

	// GetUserInfo retrieves information about a user
	GetUserInfo(userID string) (*slack.User, error)

	// GetConversationHistory retrieves message history for a conversation
	GetConversationHistory(params *slack.GetConversationHistoryParameters) (*slack.GetConversationHistoryResponse, error)

	// GetConversationReplies retrieves replies to a thread
	GetConversationReplies(params *slack.GetConversationRepliesParameters) ([]slack.Message, bool, string, error)
}

// Ensure *slack.Client implements SlackAPI interface at compile time
var _ SlackAPI = (*slack.Client)(nil)
