package entities

import "time"

// Message represents a Slack message with its metadata
type Message struct {
	Timestamp       time.Time
	ThreadTS        string // Thread timestamp (parent message timestamp)
	Text            string
	UserDisplayName string // Display name of the user who sent the message
	Replies         []Message
	IsParent        bool // True if this message has replies
}

// NewMessage creates a new Message
func NewMessage(timestamp time.Time, text, userDisplayName string) *Message {
	return &Message{
		Timestamp:       timestamp,
		Text:            text,
		UserDisplayName: userDisplayName,
		Replies:         []Message{},
		IsParent:        false,
	}
}

// AddReply adds a reply to this message
func (m *Message) AddReply(reply Message) {
	m.Replies = append(m.Replies, reply)
	m.IsParent = true
}
