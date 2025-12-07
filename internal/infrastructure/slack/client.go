package slack

import (
	slackapi "github.com/slack-go/slack"
)

// SlackClient defines the interface for Slack API operations
type SlackClient interface {
	GetConversationsForUser(params *slackapi.GetConversationsForUserParameters) ([]slackapi.Channel, string, error)
	GetConversationHistory(params *slackapi.GetConversationHistoryParameters) (*slackapi.GetConversationHistoryResponse, error)
	GetConversationReplies(params *slackapi.GetConversationRepliesParameters) ([]slackapi.Message, bool, string, error)
	GetUserInfo(userID string) (*slackapi.User, error)
}

// SlackClientImpl wraps the slack-go/slack Client
type SlackClientImpl struct {
	api *slackapi.Client
}

// NewSlackClient creates a new Slack API client
func NewSlackClient(botToken, appToken string) *SlackClientImpl {
	opts := []slackapi.Option{}
	if appToken != "" {
		opts = append(opts, slackapi.OptionAppLevelToken(appToken))
	}

	return &SlackClientImpl{
		api: slackapi.New(botToken, opts...),
	}
}

// GetConversationsForUser gets conversations for the user
func (c *SlackClientImpl) GetConversationsForUser(params *slackapi.GetConversationsForUserParameters) ([]slackapi.Channel, string, error) {
	return c.api.GetConversationsForUser(params)
}

// GetConversationHistory gets the conversation history
func (c *SlackClientImpl) GetConversationHistory(params *slackapi.GetConversationHistoryParameters) (*slackapi.GetConversationHistoryResponse, error) {
	return c.api.GetConversationHistory(params)
}

// GetConversationReplies gets conversation replies
func (c *SlackClientImpl) GetConversationReplies(params *slackapi.GetConversationRepliesParameters) ([]slackapi.Message, bool, string, error) {
	return c.api.GetConversationReplies(params)
}

// GetUserInfo gets user information
func (c *SlackClientImpl) GetUserInfo(userID string) (*slackapi.User, error) {
	return c.api.GetUserInfo(userID)
}
