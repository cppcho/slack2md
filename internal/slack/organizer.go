package slack

import (
	"sort"
)

// MessagesByDate maps date strings (YYYY-MM-DD) to messages on that date
type MessagesByDate map[string][]Message

// OrganizeMessagesByDate groups messages by date, with thread replies grouped under parent's date
func OrganizeMessagesByDate(messages []Message) MessagesByDate {
	organized := make(MessagesByDate)

	for _, msg := range messages {
		// Use parent message's date for grouping
		date := msg.Timestamp.Format("2006-01-02")

		// Only include top-level messages (not replies in the main list)
		// Thread replies are already attached to their parent message
		if msg.ThreadTS == "" || msg.IsParent {
			organized[date] = append(organized[date], msg)
		}
	}

	// Sort messages within each date by timestamp
	for date := range organized {
		sort.Slice(organized[date], func(i, j int) bool {
			return organized[date][i].Timestamp.Before(organized[date][j].Timestamp)
		})
	}

	return organized
}
