package slack

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/slack-go/slack"
)

func TestParseSlackTimestamp(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		validate func(t *testing.T, result time.Time)
	}{
		{
			name:    "standard timestamp format",
			input:   "1234567890.123456",
			wantErr: false,
			validate: func(t *testing.T, result time.Time) {
				expected := time.Unix(1234567890, 123456000)
				if !result.Equal(expected) {
					t.Errorf("got %v, want %v", result, expected)
				}
			},
		},
		{
			name:    "timestamp with zero microseconds",
			input:   "1234567890.000000",
			wantErr: false,
			validate: func(t *testing.T, result time.Time) {
				expected := time.Unix(1234567890, 0)
				if !result.Equal(expected) {
					t.Errorf("got %v, want %v", result, expected)
				}
			},
		},
		{
			name:    "timestamp with max microseconds",
			input:   "1234567890.999999",
			wantErr: false,
			validate: func(t *testing.T, result time.Time) {
				expected := time.Unix(1234567890, 999999000)
				if !result.Equal(expected) {
					t.Errorf("got %v, want %v", result, expected)
				}
			},
		},
		{
			name:    "recent timestamp (2024)",
			input:   "1704067200.500000",
			wantErr: false,
			validate: func(t *testing.T, result time.Time) {
				// 2024-01-01 00:00:00 UTC
				expected := time.Unix(1704067200, 500000000)
				if !result.Equal(expected) {
					t.Errorf("got %v, want %v", result, expected)
				}
			},
		},
		{
			name:    "old timestamp (2015)",
			input:   "1420070400.000000",
			wantErr: false,
			validate: func(t *testing.T, result time.Time) {
				// 2015-01-01 00:00:00 UTC
				expected := time.Unix(1420070400, 0)
				if !result.Equal(expected) {
					t.Errorf("got %v, want %v", result, expected)
				}
			},
		},
		{
			name:    "midnight timestamp",
			input:   "1704067200.000000",
			wantErr: false,
			validate: func(t *testing.T, result time.Time) {
				// 2024-01-01 00:00:00 UTC
				expected := time.Unix(1704067200, 0)
				if !result.Equal(expected) {
					t.Errorf("got %v, want %v", result, expected)
				}
			},
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid format - no decimal",
			input:   "1234567890",
			wantErr: true,
		},
		{
			name:    "invalid format - letters in timestamp",
			input:   "123abc.456",
			wantErr: true,
		},
		{
			name:    "invalid format - text",
			input:   "not-a-timestamp",
			wantErr: true,
		},
		{
			name:    "invalid format - missing seconds",
			input:   ".123456",
			wantErr: true,
		},
		{
			name:    "invalid format - trailing dot",
			input:   "1234567890.",
			wantErr: true,
		},
		{
			name:    "zero timestamp (epoch)",
			input:   "0.000000",
			wantErr: false,
			validate: func(t *testing.T, result time.Time) {
				expected := time.Unix(0, 0)
				if !result.Equal(expected) {
					t.Errorf("got %v, want %v", result, expected)
				}
			},
		},
		{
			name:    "extremely large timestamp",
			input:   "99999999999.999999",
			wantErr: false,
			validate: func(t *testing.T, result time.Time) {
				expected := time.Unix(99999999999, 999999000)
				if !result.Equal(expected) {
					t.Errorf("got %v, want %v", result, expected)
				}
			},
		},
		{
			name:    "timestamp with fewer microseconds digits",
			input:   "1234567890.123",
			wantErr: false,
			validate: func(t *testing.T, result time.Time) {
				expected := time.Unix(1234567890, 123000)
				if !result.Equal(expected) {
					t.Errorf("got %v, want %v", result, expected)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseSlackTimestamp(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseSlackTimestamp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestFetchThreadReplies(t *testing.T) {
	tests := []struct {
		name      string
		channelID string
		threadTS  string
		mockSetup func(*mockSlackAPI)
		want      []Message
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "single reply excludes parent",
			channelID: "C123",
			threadTS:  "1234567890.000000",
			mockSetup: func(m *mockSlackAPI) {
				m.getConversationRepliesFunc = func(params *slack.GetConversationRepliesParameters) ([]slack.Message, bool, string, error) {
					return []slack.Message{
						{
							Msg: slack.Msg{
								Timestamp: "1234567890.000000",
								Text:      "Parent message",
								User:      "U001",
							},
						},
						{
							Msg: slack.Msg{
								Timestamp: "1234567891.000000",
								Text:      "Reply 1",
								User:      "U002",
							},
						},
					}, false, "", nil
				}
				m.getUserInfoFunc = func(userID string) (*slack.User, error) {
					return &slack.User{
						Profile: slack.UserProfile{DisplayName: "User " + userID},
					}, nil
				}
			},
			want: []Message{
				{
					Timestamp:       time.Unix(1234567891, 0),
					ThreadTS:        "",
					Text:            "Reply 1",
					UserDisplayName: "User U002",
				},
			},
			wantErr: false,
		},
		{
			name:      "multiple replies",
			channelID: "C123",
			threadTS:  "1234567890.000000",
			mockSetup: func(m *mockSlackAPI) {
				m.getConversationRepliesFunc = func(params *slack.GetConversationRepliesParameters) ([]slack.Message, bool, string, error) {
					return []slack.Message{
						{Msg: slack.Msg{Timestamp: "1234567890.000000", Text: "Parent", User: "U001"}},
						{Msg: slack.Msg{Timestamp: "1234567891.000000", Text: "Reply 1", User: "U002"}},
						{Msg: slack.Msg{Timestamp: "1234567892.000000", Text: "Reply 2", User: "U003"}},
						{Msg: slack.Msg{Timestamp: "1234567893.000000", Text: "Reply 3", User: "U004"}},
					}, false, "", nil
				}
				m.getUserInfoFunc = func(userID string) (*slack.User, error) {
					return &slack.User{Profile: slack.UserProfile{DisplayName: "User"}}, nil
				}
			},
			want: []Message{
				{Timestamp: time.Unix(1234567891, 0), Text: "Reply 1", UserDisplayName: "User"},
				{Timestamp: time.Unix(1234567892, 0), Text: "Reply 2", UserDisplayName: "User"},
				{Timestamp: time.Unix(1234567893, 0), Text: "Reply 3", UserDisplayName: "User"},
			},
			wantErr: false,
		},
		{
			name:      "reply with empty text skipped",
			channelID: "C123",
			threadTS:  "1234567890.000000",
			mockSetup: func(m *mockSlackAPI) {
				m.getConversationRepliesFunc = func(params *slack.GetConversationRepliesParameters) ([]slack.Message, bool, string, error) {
					return []slack.Message{
						{Msg: slack.Msg{Timestamp: "1234567890.000000", Text: "Parent", User: "U001"}},
						{Msg: slack.Msg{Timestamp: "1234567891.000000", Text: "", User: "U002"}}, // Empty, should be skipped
						{Msg: slack.Msg{Timestamp: "1234567892.000000", Text: "Reply 2", User: "U003"}},
					}, false, "", nil
				}
				m.getUserInfoFunc = func(userID string) (*slack.User, error) {
					return &slack.User{Profile: slack.UserProfile{DisplayName: "User"}}, nil
				}
			},
			want: []Message{
				{Timestamp: time.Unix(1234567892, 0), Text: "Reply 2", UserDisplayName: "User"},
			},
			wantErr: false,
		},
		{
			name:      "api error propagates",
			channelID: "C123",
			threadTS:  "1234567890.000000",
			mockSetup: func(m *mockSlackAPI) {
				m.getConversationRepliesFunc = func(params *slack.GetConversationRepliesParameters) ([]slack.Message, bool, string, error) {
					return nil, false, "", fmt.Errorf("api error")
				}
			},
			wantErr: true,
			errMsg:  "failed to fetch thread replies",
		},
		{
			name:      "empty thread returns empty slice",
			channelID: "C123",
			threadTS:  "1234567890.000000",
			mockSetup: func(m *mockSlackAPI) {
				m.getConversationRepliesFunc = func(params *slack.GetConversationRepliesParameters) ([]slack.Message, bool, string, error) {
					return []slack.Message{
						{Msg: slack.Msg{Timestamp: "1234567890.000000", Text: "Parent only", User: "U001"}},
					}, false, "", nil
				}
			},
			want:    []Message{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockSlackAPI{}
			tt.mockSetup(mock)
			client := NewClientWithAPI(mock)

			replies, err := client.fetchThreadReplies(tt.channelID, tt.threadTS)

			if (err != nil) != tt.wantErr {
				t.Errorf("fetchThreadReplies() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("error message = %q, want to contain %q", err.Error(), tt.errMsg)
			}

			if !tt.wantErr {
				if len(replies) != len(tt.want) {
					t.Errorf("fetchThreadReplies() returned %d replies, want %d", len(replies), len(tt.want))
					return
				}
				for i, reply := range replies {
					if !reply.Timestamp.Equal(tt.want[i].Timestamp) || reply.Text != tt.want[i].Text {
						t.Errorf("fetchThreadReplies()[%d] = %+v, want %+v", i, reply, tt.want[i])
					}
				}
			}
		})
	}
}

func TestFetchChannelMessages(t *testing.T) {
	tests := []struct {
		name      string
		channelID string
		startTime time.Time
		endTime   time.Time
		mockSetup func(*mockSlackAPI)
		validate  func(t *testing.T, result *FetchResult)
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "single message no pagination",
			channelID: "C123",
			startTime: time.Unix(1000000, 0),
			endTime:   time.Unix(2000000, 0),
			mockSetup: func(m *mockSlackAPI) {
				m.getConversationInfoFunc = func(params *slack.GetConversationInfoInput) (*slack.Channel, error) {
					return &slack.Channel{GroupConversation: slack.GroupConversation{Name: "test-channel"}}, nil
				}
				m.getConversationHistoryFunc = func(params *slack.GetConversationHistoryParameters) (*slack.GetConversationHistoryResponse, error) {
					return &slack.GetConversationHistoryResponse{
						Messages: []slack.Message{
							{Msg: slack.Msg{Timestamp: "1500000.000000", Text: "Hello", User: "U001"}},
						},
						HasMore: false,
					}, nil
				}
				m.getUserInfoFunc = func(userID string) (*slack.User, error) {
					return &slack.User{Profile: slack.UserProfile{DisplayName: "TestUser"}}, nil
				}
			},
			validate: func(t *testing.T, result *FetchResult) {
				if result.ChannelName != "test-channel" {
					t.Errorf("ChannelName = %q, want %q", result.ChannelName, "test-channel")
				}
				if len(result.Messages) != 1 {
					t.Errorf("got %d messages, want 1", len(result.Messages))
					return
				}
				if result.Messages[0].Text != "Hello" {
					t.Errorf("message text = %q, want %q", result.Messages[0].Text, "Hello")
				}
			},
			wantErr: false,
		},
		{
			name:      "multiple pages of messages",
			channelID: "C123",
			startTime: time.Unix(1000000, 0),
			endTime:   time.Unix(2000000, 0),
			mockSetup: func(m *mockSlackAPI) {
				callCount := 0
				m.getConversationInfoFunc = func(params *slack.GetConversationInfoInput) (*slack.Channel, error) {
					return &slack.Channel{GroupConversation: slack.GroupConversation{Name: "test"}}, nil
				}
				m.getConversationHistoryFunc = func(params *slack.GetConversationHistoryParameters) (*slack.GetConversationHistoryResponse, error) {
					callCount++
					if callCount == 1 {
						return &slack.GetConversationHistoryResponse{
							Messages: []slack.Message{
								{Msg: slack.Msg{Timestamp: "1500000.000000", Text: "Msg1", User: "U001"}},
							},
							HasMore:          true,
						}, nil
					}
					return &slack.GetConversationHistoryResponse{
						Messages: []slack.Message{
							{Msg: slack.Msg{Timestamp: "1600000.000000", Text: "Msg2", User: "U001"}},
						},
						HasMore: false,
					}, nil
				}
				m.getUserInfoFunc = func(userID string) (*slack.User, error) {
					return &slack.User{Profile: slack.UserProfile{DisplayName: "User"}}, nil
				}
			},
			validate: func(t *testing.T, result *FetchResult) {
				if len(result.Messages) != 2 {
					t.Errorf("got %d messages, want 2", len(result.Messages))
				}
			},
			wantErr: false,
		},
		{
			name:      "message with empty text skipped",
			channelID: "C123",
			startTime: time.Unix(1000000, 0),
			endTime:   time.Unix(2000000, 0),
			mockSetup: func(m *mockSlackAPI) {
				m.getConversationInfoFunc = func(params *slack.GetConversationInfoInput) (*slack.Channel, error) {
					return &slack.Channel{GroupConversation: slack.GroupConversation{Name: "test"}}, nil
				}
				m.getConversationHistoryFunc = func(params *slack.GetConversationHistoryParameters) (*slack.GetConversationHistoryResponse, error) {
					return &slack.GetConversationHistoryResponse{
						Messages: []slack.Message{
							{Msg: slack.Msg{Timestamp: "1500000.000000", Text: "", User: "U001"}},        // Empty, skip
							{Msg: slack.Msg{Timestamp: "1600000.000000", Text: "Valid", User: "U001"}}, // Keep
						},
						HasMore: false,
					}, nil
				}
				m.getUserInfoFunc = func(userID string) (*slack.User, error) {
					return &slack.User{Profile: slack.UserProfile{DisplayName: "User"}}, nil
				}
			},
			validate: func(t *testing.T, result *FetchResult) {
				if len(result.Messages) != 1 {
					t.Errorf("got %d messages, want 1", len(result.Messages))
					return
				}
				if result.Messages[0].Text != "Valid" {
					t.Errorf("message text = %q, want %q", result.Messages[0].Text, "Valid")
				}
			},
			wantErr: false,
		},
		{
			name:      "parent message with replies",
			channelID: "C123",
			startTime: time.Unix(1000000, 0),
			endTime:   time.Unix(2000000, 0),
			mockSetup: func(m *mockSlackAPI) {
				m.getConversationInfoFunc = func(params *slack.GetConversationInfoInput) (*slack.Channel, error) {
					return &slack.Channel{GroupConversation: slack.GroupConversation{Name: "test"}}, nil
				}
				m.getConversationHistoryFunc = func(params *slack.GetConversationHistoryParameters) (*slack.GetConversationHistoryResponse, error) {
					return &slack.GetConversationHistoryResponse{
						Messages: []slack.Message{
							{
								Msg: slack.Msg{
									Timestamp:       "1500000.000000",
									ThreadTimestamp: "1500000.000000",
									Text:            "Parent",
									User:            "U001",
									ReplyCount:      2,
								},
							},
						},
						HasMore: false,
					}, nil
				}
				m.getConversationRepliesFunc = func(params *slack.GetConversationRepliesParameters) ([]slack.Message, bool, string, error) {
					return []slack.Message{
						{Msg: slack.Msg{Timestamp: "1500000.000000", Text: "Parent", User: "U001"}},
						{Msg: slack.Msg{Timestamp: "1500001.000000", Text: "Reply1", User: "U002"}},
						{Msg: slack.Msg{Timestamp: "1500002.000000", Text: "Reply2", User: "U003"}},
					}, false, "", nil
				}
				m.getUserInfoFunc = func(userID string) (*slack.User, error) {
					return &slack.User{Profile: slack.UserProfile{DisplayName: "User"}}, nil
				}
			},
			validate: func(t *testing.T, result *FetchResult) {
				if len(result.Messages) != 1 {
					t.Errorf("got %d messages, want 1", len(result.Messages))
					return
				}
				if !result.Messages[0].IsParent {
					t.Error("message should be marked as parent")
				}
				if len(result.Messages[0].Replies) != 2 {
					t.Errorf("got %d replies, want 2", len(result.Messages[0].Replies))
				}
			},
			wantErr: false,
		},
		{
			name:      "GetChannelName fails",
			channelID: "C123",
			startTime: time.Unix(1000000, 0),
			endTime:   time.Unix(2000000, 0),
			mockSetup: func(m *mockSlackAPI) {
				m.getConversationInfoFunc = func(params *slack.GetConversationInfoInput) (*slack.Channel, error) {
					return nil, fmt.Errorf("channel not found")
				}
			},
			wantErr: true,
			errMsg:  "failed to get channel info",
		},
		{
			name:      "GetConversationHistory fails",
			channelID: "C123",
			startTime: time.Unix(1000000, 0),
			endTime:   time.Unix(2000000, 0),
			mockSetup: func(m *mockSlackAPI) {
				m.getConversationInfoFunc = func(params *slack.GetConversationInfoInput) (*slack.Channel, error) {
					return &slack.Channel{GroupConversation: slack.GroupConversation{Name: "test"}}, nil
				}
				m.getConversationHistoryFunc = func(params *slack.GetConversationHistoryParameters) (*slack.GetConversationHistoryResponse, error) {
					return nil, fmt.Errorf("api error")
				}
			},
			wantErr: true,
			errMsg:  "failed to fetch conversation history",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockSlackAPI{}
			tt.mockSetup(mock)
			client := NewClientWithAPI(mock)

			result, err := client.FetchChannelMessages(tt.channelID, tt.startTime, tt.endTime)

			if (err != nil) != tt.wantErr {
				t.Errorf("FetchChannelMessages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("error message = %q, want to contain %q", err.Error(), tt.errMsg)
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}
