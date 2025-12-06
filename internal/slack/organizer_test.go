package slack

import (
	"testing"
	"time"
)

func TestOrganizeMessagesByDate(t *testing.T) {
	// Helper function to create a time from date string
	parseTime := func(s string) time.Time {
		t, _ := time.Parse("2006-01-02 15:04:05", s)
		return t
	}

	tests := []struct {
		name     string
		messages []Message
		validate func(t *testing.T, result MessagesByDate)
	}{
		{
			name:     "empty message list",
			messages: []Message{},
			validate: func(t *testing.T, result MessagesByDate) {
				if len(result) != 0 {
					t.Errorf("expected empty map, got %d entries", len(result))
				}
			},
		},
		{
			name: "single message",
			messages: []Message{
				{
					Timestamp:       parseTime("2024-01-15 10:30:00"),
					Text:            "Hello",
					UserDisplayName: "User1",
					ThreadTS:        "",
					IsParent:        false,
				},
			},
			validate: func(t *testing.T, result MessagesByDate) {
				if len(result) != 1 {
					t.Errorf("expected 1 date entry, got %d", len(result))
				}
				msgs, ok := result["2024-01-15"]
				if !ok {
					t.Errorf("expected date 2024-01-15 to exist")
					return
				}
				if len(msgs) != 1 {
					t.Errorf("expected 1 message on 2024-01-15, got %d", len(msgs))
				}
				if msgs[0].Text != "Hello" {
					t.Errorf("expected message text 'Hello', got %q", msgs[0].Text)
				}
			},
		},
		{
			name: "multiple messages same day",
			messages: []Message{
				{
					Timestamp:       parseTime("2024-01-15 10:30:00"),
					Text:            "First",
					UserDisplayName: "User1",
				},
				{
					Timestamp:       parseTime("2024-01-15 14:45:00"),
					Text:            "Second",
					UserDisplayName: "User2",
				},
				{
					Timestamp:       parseTime("2024-01-15 09:00:00"),
					Text:            "Third",
					UserDisplayName: "User3",
				},
			},
			validate: func(t *testing.T, result MessagesByDate) {
				if len(result) != 1 {
					t.Errorf("expected 1 date entry, got %d", len(result))
				}
				msgs, ok := result["2024-01-15"]
				if !ok {
					t.Errorf("expected date 2024-01-15 to exist")
					return
				}
				if len(msgs) != 3 {
					t.Errorf("expected 3 messages on 2024-01-15, got %d", len(msgs))
					return
				}
				// Verify chronological order (sorted by timestamp)
				if msgs[0].Text != "Third" {
					t.Errorf("expected first message to be 'Third', got %q", msgs[0].Text)
				}
				if msgs[1].Text != "First" {
					t.Errorf("expected second message to be 'First', got %q", msgs[1].Text)
				}
				if msgs[2].Text != "Second" {
					t.Errorf("expected third message to be 'Second', got %q", msgs[2].Text)
				}
			},
		},
		{
			name: "multiple messages different days",
			messages: []Message{
				{
					Timestamp:       parseTime("2024-01-15 10:30:00"),
					Text:            "Day 1",
					UserDisplayName: "User1",
				},
				{
					Timestamp:       parseTime("2024-01-16 14:45:00"),
					Text:            "Day 2",
					UserDisplayName: "User2",
				},
				{
					Timestamp:       parseTime("2024-01-17 09:00:00"),
					Text:            "Day 3",
					UserDisplayName: "User3",
				},
			},
			validate: func(t *testing.T, result MessagesByDate) {
				if len(result) != 3 {
					t.Errorf("expected 3 date entries, got %d", len(result))
				}
				dates := []string{"2024-01-15", "2024-01-16", "2024-01-17"}
				texts := []string{"Day 1", "Day 2", "Day 3"}
				for i, date := range dates {
					msgs, ok := result[date]
					if !ok {
						t.Errorf("expected date %s to exist", date)
						continue
					}
					if len(msgs) != 1 {
						t.Errorf("expected 1 message on %s, got %d", date, len(msgs))
						continue
					}
					if msgs[0].Text != texts[i] {
						t.Errorf("expected message text %q on %s, got %q", texts[i], date, msgs[0].Text)
					}
				}
			},
		},
		{
			name: "parent message with thread replies",
			messages: []Message{
				{
					Timestamp:       parseTime("2024-01-15 10:30:00"),
					Text:            "Parent message",
					UserDisplayName: "User1",
					ThreadTS:        "1705317000.000000",
					IsParent:        true,
					Replies: []Message{
						{
							Timestamp:       parseTime("2024-01-15 10:35:00"),
							Text:            "Reply 1",
							UserDisplayName: "User2",
							ThreadTS:        "1705317000.000000",
						},
						{
							Timestamp:       parseTime("2024-01-15 10:40:00"),
							Text:            "Reply 2",
							UserDisplayName: "User3",
							ThreadTS:        "1705317000.000000",
						},
					},
				},
			},
			validate: func(t *testing.T, result MessagesByDate) {
				if len(result) != 1 {
					t.Errorf("expected 1 date entry, got %d", len(result))
				}
				msgs, ok := result["2024-01-15"]
				if !ok {
					t.Errorf("expected date 2024-01-15 to exist")
					return
				}
				if len(msgs) != 1 {
					t.Errorf("expected 1 parent message on 2024-01-15, got %d", len(msgs))
					return
				}
				if msgs[0].Text != "Parent message" {
					t.Errorf("expected parent message text, got %q", msgs[0].Text)
				}
				if len(msgs[0].Replies) != 2 {
					t.Errorf("expected 2 replies, got %d", len(msgs[0].Replies))
				}
			},
		},
		{
			name: "thread reply without parent flag should be excluded",
			messages: []Message{
				{
					Timestamp:       parseTime("2024-01-15 10:30:00"),
					Text:            "Parent",
					UserDisplayName: "User1",
					ThreadTS:        "1705317000.000000",
					IsParent:        true,
				},
				{
					Timestamp:       parseTime("2024-01-15 10:35:00"),
					Text:            "Reply (should be excluded)",
					UserDisplayName: "User2",
					ThreadTS:        "1705317000.000000",
					IsParent:        false,
				},
			},
			validate: func(t *testing.T, result MessagesByDate) {
				msgs, ok := result["2024-01-15"]
				if !ok {
					t.Errorf("expected date 2024-01-15 to exist")
					return
				}
				if len(msgs) != 1 {
					t.Errorf("expected only parent message, got %d messages", len(msgs))
					return
				}
				if msgs[0].Text != "Parent" {
					t.Errorf("expected parent message, got %q", msgs[0].Text)
				}
			},
		},
		{
			name: "cross-midnight thread (parent and reply on different dates)",
			messages: []Message{
				{
					Timestamp:       parseTime("2024-01-15 23:55:00"),
					Text:            "Late night parent",
					UserDisplayName: "User1",
					ThreadTS:        "1705359300.000000",
					IsParent:        true,
					Replies: []Message{
						{
							Timestamp:       parseTime("2024-01-16 00:05:00"),
							Text:            "Early morning reply",
							UserDisplayName: "User2",
							ThreadTS:        "1705359300.000000",
						},
					},
				},
			},
			validate: func(t *testing.T, result MessagesByDate) {
				// Parent message should be on 2024-01-15
				msgs, ok := result["2024-01-15"]
				if !ok {
					t.Errorf("expected date 2024-01-15 to exist")
					return
				}
				if len(msgs) != 1 {
					t.Errorf("expected 1 message on 2024-01-15, got %d", len(msgs))
					return
				}
				if msgs[0].Text != "Late night parent" {
					t.Errorf("expected parent message, got %q", msgs[0].Text)
				}
				// Reply is attached to parent, not a separate date entry
				if len(msgs[0].Replies) != 1 {
					t.Errorf("expected 1 reply, got %d", len(msgs[0].Replies))
				}
			},
		},
		{
			name: "messages with timezone information",
			messages: []Message{
				{
					Timestamp:       parseTime("2024-01-15 10:30:00").UTC(),
					Text:            "UTC message",
					UserDisplayName: "User1",
				},
			},
			validate: func(t *testing.T, result MessagesByDate) {
				msgs, ok := result["2024-01-15"]
				if !ok {
					t.Errorf("expected date 2024-01-15 to exist")
					return
				}
				if len(msgs) != 1 {
					t.Errorf("expected 1 message, got %d", len(msgs))
				}
			},
		},
		{
			name: "sorting verification - out of order timestamps",
			messages: []Message{
				{
					Timestamp:       parseTime("2024-01-15 15:00:00"),
					Text:            "Third",
					UserDisplayName: "User1",
				},
				{
					Timestamp:       parseTime("2024-01-15 09:00:00"),
					Text:            "First",
					UserDisplayName: "User2",
				},
				{
					Timestamp:       parseTime("2024-01-15 22:00:00"),
					Text:            "Fourth",
					UserDisplayName: "User3",
				},
				{
					Timestamp:       parseTime("2024-01-15 12:00:00"),
					Text:            "Second",
					UserDisplayName: "User4",
				},
			},
			validate: func(t *testing.T, result MessagesByDate) {
				msgs, ok := result["2024-01-15"]
				if !ok {
					t.Errorf("expected date 2024-01-15 to exist")
					return
				}
				if len(msgs) != 4 {
					t.Errorf("expected 4 messages, got %d", len(msgs))
					return
				}
				expectedOrder := []string{"First", "Second", "Third", "Fourth"}
				for i, expected := range expectedOrder {
					if msgs[i].Text != expected {
						t.Errorf("message %d: expected %q, got %q", i, expected, msgs[i].Text)
					}
				}
			},
		},
		{
			name: "mixed parent messages and standalone messages",
			messages: []Message{
				{
					Timestamp:       parseTime("2024-01-15 10:00:00"),
					Text:            "Standalone 1",
					UserDisplayName: "User1",
					ThreadTS:        "",
					IsParent:        false,
				},
				{
					Timestamp:       parseTime("2024-01-15 11:00:00"),
					Text:            "Parent with replies",
					UserDisplayName: "User2",
					ThreadTS:        "1705316400.000000",
					IsParent:        true,
					Replies: []Message{
						{
							Timestamp:       parseTime("2024-01-15 11:05:00"),
							Text:            "Reply",
							UserDisplayName: "User3",
							ThreadTS:        "1705316400.000000",
						},
					},
				},
				{
					Timestamp:       parseTime("2024-01-15 12:00:00"),
					Text:            "Standalone 2",
					UserDisplayName: "User4",
					ThreadTS:        "",
					IsParent:        false,
				},
			},
			validate: func(t *testing.T, result MessagesByDate) {
				msgs, ok := result["2024-01-15"]
				if !ok {
					t.Errorf("expected date 2024-01-15 to exist")
					return
				}
				if len(msgs) != 3 {
					t.Errorf("expected 3 top-level messages, got %d", len(msgs))
					return
				}
				// Verify order
				if msgs[0].Text != "Standalone 1" {
					t.Errorf("expected first message 'Standalone 1', got %q", msgs[0].Text)
				}
				if msgs[1].Text != "Parent with replies" {
					t.Errorf("expected second message 'Parent with replies', got %q", msgs[1].Text)
				}
				if msgs[2].Text != "Standalone 2" {
					t.Errorf("expected third message 'Standalone 2', got %q", msgs[2].Text)
				}
				// Verify parent has reply
				if len(msgs[1].Replies) != 1 {
					t.Errorf("expected parent to have 1 reply, got %d", len(msgs[1].Replies))
				}
			},
		},
		{
			name: "date format edge case - year boundary",
			messages: []Message{
				{
					Timestamp:       parseTime("2023-12-31 23:59:59"),
					Text:            "Last message of 2023",
					UserDisplayName: "User1",
				},
				{
					Timestamp:       parseTime("2024-01-01 00:00:01"),
					Text:            "First message of 2024",
					UserDisplayName: "User2",
				},
			},
			validate: func(t *testing.T, result MessagesByDate) {
				if len(result) != 2 {
					t.Errorf("expected 2 date entries, got %d", len(result))
				}
				msgs2023, ok := result["2023-12-31"]
				if !ok || len(msgs2023) != 1 {
					t.Errorf("expected 1 message on 2023-12-31")
				}
				msgs2024, ok := result["2024-01-01"]
				if !ok || len(msgs2024) != 1 {
					t.Errorf("expected 1 message on 2024-01-01")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := OrganizeMessagesByDate(tt.messages)
			tt.validate(t, result)
		})
	}
}

// TestOrganizeMessagesByDate_DateFormat verifies the date format is YYYY-MM-DD
func TestOrganizeMessagesByDate_DateFormat(t *testing.T) {
	timestamp, _ := time.Parse("2006-01-02 15:04:05", "2024-03-05 10:30:00")
	messages := []Message{
		{
			Timestamp:       timestamp,
			Text:            "Test",
			UserDisplayName: "User1",
		},
	}

	result := OrganizeMessagesByDate(messages)

	expectedDate := "2024-03-05"
	if _, ok := result[expectedDate]; !ok {
		t.Errorf("expected date format %q, got keys: %v", expectedDate, getKeys(result))
	}
}

// Helper function to get map keys
func getKeys(m MessagesByDate) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
