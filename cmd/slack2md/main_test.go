package main_test

import (
	"testing"
	"time"

	main "github.com/cppcho/slack2md/cmd/slack2md"
	"github.com/cppcho/slack2md/internal/filewriter"
	"github.com/cppcho/slack2md/internal/slack"
)

func TestConvertMessage(t *testing.T) {
	tests := []struct {
		name     string
		input    slack.Message
		expected filewriter.Message
	}{
		{
			name: "simple message no replies",
			input: slack.Message{
				Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				ThreadTS:  "",
				Text:      "Hello world",
				IsParent:  false,
				Replies:   nil,
			},
			expected: filewriter.Message{
				Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				ThreadTS:  "",
				Text:      "Hello world",
				IsParent:  false,
				Replies:   nil,
			},
		},
		{
			name: "message with ThreadTS",
			input: slack.Message{
				Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				ThreadTS:  "1234567890.123456",
				Text:      "Reply in thread",
				IsParent:  false,
				Replies:   nil,
			},
			expected: filewriter.Message{
				Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				ThreadTS:  "1234567890.123456",
				Text:      "Reply in thread",
				IsParent:  false,
				Replies:   nil,
			},
		},
		{
			name: "message with IsParent flag",
			input: slack.Message{
				Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				ThreadTS:  "",
				Text:      "Parent message",
				IsParent:  true,
				Replies:   nil,
			},
			expected: filewriter.Message{
				Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				ThreadTS:  "",
				Text:      "Parent message",
				IsParent:  true,
				Replies:   nil,
			},
		},
		{
			name: "message with zero timestamp",
			input: slack.Message{
				Timestamp: time.Time{},
				ThreadTS:  "",
				Text:      "Message with zero time",
				IsParent:  false,
				Replies:   nil,
			},
			expected: filewriter.Message{
				Timestamp: time.Time{},
				ThreadTS:  "",
				Text:      "Message with zero time",
				IsParent:  false,
				Replies:   nil,
			},
		},
		{
			name: "parent message with single reply",
			input: slack.Message{
				Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				ThreadTS:  "",
				Text:      "Parent",
				IsParent:  true,
				Replies: []slack.Message{
					{
						Timestamp: time.Date(2024, 1, 15, 10, 31, 0, 0, time.UTC),
						ThreadTS:  "1234567890.123456",
						Text:      "Reply 1",
						IsParent:  false,
						Replies:   nil,
					},
				},
			},
			expected: filewriter.Message{
				Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				ThreadTS:  "",
				Text:      "Parent",
				IsParent:  true,
				Replies: []filewriter.Message{
					{
						Timestamp: time.Date(2024, 1, 15, 10, 31, 0, 0, time.UTC),
						ThreadTS:  "1234567890.123456",
						Text:      "Reply 1",
						IsParent:  false,
						Replies:   nil,
					},
				},
			},
		},
		{
			name: "parent message with multiple replies",
			input: slack.Message{
				Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				ThreadTS:  "",
				Text:      "Parent",
				IsParent:  true,
				Replies: []slack.Message{
					{
						Timestamp: time.Date(2024, 1, 15, 10, 31, 0, 0, time.UTC),
						ThreadTS:  "1234567890.123456",
						Text:      "Reply 1",
						IsParent:  false,
						Replies:   nil,
					},
					{
						Timestamp: time.Date(2024, 1, 15, 10, 32, 0, 0, time.UTC),
						ThreadTS:  "1234567890.123456",
						Text:      "Reply 2",
						IsParent:  false,
						Replies:   nil,
					},
					{
						Timestamp: time.Date(2024, 1, 15, 10, 33, 0, 0, time.UTC),
						ThreadTS:  "1234567890.123456",
						Text:      "Reply 3",
						IsParent:  false,
						Replies:   nil,
					},
				},
			},
			expected: filewriter.Message{
				Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				ThreadTS:  "",
				Text:      "Parent",
				IsParent:  true,
				Replies: []filewriter.Message{
					{
						Timestamp: time.Date(2024, 1, 15, 10, 31, 0, 0, time.UTC),
						ThreadTS:  "1234567890.123456",
						Text:      "Reply 1",
						IsParent:  false,
						Replies:   nil,
					},
					{
						Timestamp: time.Date(2024, 1, 15, 10, 32, 0, 0, time.UTC),
						ThreadTS:  "1234567890.123456",
						Text:      "Reply 2",
						IsParent:  false,
						Replies:   nil,
					},
					{
						Timestamp: time.Date(2024, 1, 15, 10, 33, 0, 0, time.UTC),
						ThreadTS:  "1234567890.123456",
						Text:      "Reply 3",
						IsParent:  false,
						Replies:   nil,
					},
				},
			},
		},
		{
			name: "nested replies (3 levels deep)",
			input: slack.Message{
				Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				ThreadTS:  "",
				Text:      "Level 1",
				IsParent:  true,
				Replies: []slack.Message{
					{
						Timestamp: time.Date(2024, 1, 15, 10, 31, 0, 0, time.UTC),
						ThreadTS:  "1234567890.123456",
						Text:      "Level 2",
						IsParent:  true,
						Replies: []slack.Message{
							{
								Timestamp: time.Date(2024, 1, 15, 10, 32, 0, 0, time.UTC),
								ThreadTS:  "1234567890.123456",
								Text:      "Level 3",
								IsParent:  false,
								Replies:   nil,
							},
						},
					},
				},
			},
			expected: filewriter.Message{
				Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				ThreadTS:  "",
				Text:      "Level 1",
				IsParent:  true,
				Replies: []filewriter.Message{
					{
						Timestamp: time.Date(2024, 1, 15, 10, 31, 0, 0, time.UTC),
						ThreadTS:  "1234567890.123456",
						Text:      "Level 2",
						IsParent:  true,
						Replies: []filewriter.Message{
							{
								Timestamp: time.Date(2024, 1, 15, 10, 32, 0, 0, time.UTC),
								ThreadTS:  "1234567890.123456",
								Text:      "Level 3",
								IsParent:  false,
								Replies:   nil,
							},
						},
					},
				},
			},
		},
		{
			name: "empty replies array",
			input: slack.Message{
				Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				ThreadTS:  "",
				Text:      "Parent with empty replies",
				IsParent:  true,
				Replies:   []slack.Message{},
			},
			expected: filewriter.Message{
				Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				ThreadTS:  "",
				Text:      "Parent with empty replies",
				IsParent:  true,
				Replies:   nil,
			},
		},
		{
			name: "message with empty text",
			input: slack.Message{
				Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				ThreadTS:  "",
				Text:      "",
				IsParent:  false,
				Replies:   nil,
			},
			expected: filewriter.Message{
				Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				ThreadTS:  "",
				Text:      "",
				IsParent:  false,
				Replies:   nil,
			},
		},
		{
			name: "message with all fields populated",
			input: slack.Message{
				Timestamp: time.Date(2024, 1, 15, 10, 30, 45, 123456789, time.UTC),
				ThreadTS:  "1234567890.123456",
				Text:      "Complete message with all fields",
				IsParent:  true,
				Replies: []slack.Message{
					{
						Timestamp: time.Date(2024, 1, 15, 10, 35, 0, 0, time.UTC),
						ThreadTS:  "1234567890.123456",
						Text:      "A reply",
						IsParent:  false,
						Replies:   nil,
					},
				},
			},
			expected: filewriter.Message{
				Timestamp: time.Date(2024, 1, 15, 10, 30, 45, 123456789, time.UTC),
				ThreadTS:  "1234567890.123456",
				Text:      "Complete message with all fields",
				IsParent:  true,
				Replies: []filewriter.Message{
					{
						Timestamp: time.Date(2024, 1, 15, 10, 35, 0, 0, time.UTC),
						ThreadTS:  "1234567890.123456",
						Text:      "A reply",
						IsParent:  false,
						Replies:   nil,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := main.ConvertMessage(tt.input)

			if !messagesEqual(result, tt.expected) {
				t.Errorf("main.ConvertMessage() failed\ngot:  %+v\nwant: %+v", result, tt.expected)
			}
		})
	}
}

// messagesEqual performs deep equality check for filewriter.Message
func messagesEqual(a, b filewriter.Message) bool {
	// Check basic fields
	if !a.Timestamp.Equal(b.Timestamp) {
		return false
	}
	if a.ThreadTS != b.ThreadTS {
		return false
	}
	if a.Text != b.Text {
		return false
	}
	if a.IsParent != b.IsParent {
		return false
	}

	// Check replies length
	if len(a.Replies) != len(b.Replies) {
		return false
	}

	// Recursively check replies
	for i := range a.Replies {
		if !messagesEqual(a.Replies[i], b.Replies[i]) {
			return false
		}
	}

	return true
}
