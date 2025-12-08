package slack

import (
	slackapi "github.com/slack-go/slack"
)

// Client defines the interface for Slack API operations
type Client interface {
	GetConversationsForUser(params *slackapi.GetConversationsForUserParameters) ([]slackapi.Channel, string, error)
	GetConversationHistory(params *slackapi.GetConversationHistoryParameters) (*slackapi.GetConversationHistoryResponse, error)
	GetConversationReplies(params *slackapi.GetConversationRepliesParameters) ([]slackapi.Message, bool, string, error)
	GetUserInfo(userID string) (*slackapi.User, error)
}

// APIClient wraps the slack-go/slack Client
type APIClient struct {
	api *slackapi.Client
}

// NewClient creates a new Slack API client
func NewClient(botToken, appToken string) *APIClient {
	opts := []slackapi.Option{}
	if appToken != "" {
		opts = append(opts, slackapi.OptionAppLevelToken(appToken))
	}

	return &APIClient{
		api: slackapi.New(botToken, opts...),
	}
}

// GetConversationsForUser gets conversations for the user
func (c *APIClient) GetConversationsForUser(params *slackapi.GetConversationsForUserParameters) ([]slackapi.Channel, string, error) {
	return c.api.GetConversationsForUser(params)
}

// GetConversationHistory gets the conversation history
func (c *APIClient) GetConversationHistory(params *slackapi.GetConversationHistoryParameters) (*slackapi.GetConversationHistoryResponse, error) {
	return c.api.GetConversationHistory(params)
}

// GetConversationReplies gets conversation replies
func (c *APIClient) GetConversationReplies(params *slackapi.GetConversationRepliesParameters) ([]slackapi.Message, bool, string, error) {
	return c.api.GetConversationReplies(params)
}

// GetUserInfo gets user information
func (c *APIClient) GetUserInfo(userID string) (*slackapi.User, error) {
	return c.api.GetUserInfo(userID)
}
