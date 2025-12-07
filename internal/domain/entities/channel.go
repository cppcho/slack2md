package entities

import "errors"

// Channel represents a Slack channel
type Channel struct {
	ID   string
	Name string
}

// NewChannel creates a new Channel with validation
func NewChannel(id, name string) (*Channel, error) {
	if id == "" {
		return nil, errors.New("channel ID cannot be empty")
	}
	if name == "" {
		return nil, errors.New("channel name cannot be empty")
	}
	return &Channel{
		ID:   id,
		Name: name,
	}, nil
}
