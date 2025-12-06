package slack

import (
	"fmt"
	"strings"
	"testing"

	"github.com/slack-go/slack"
)

// mockSlackAPI implements SlackAPI for testing
type mockSlackAPI struct {
	GetConversationsForUserFunc func(*slack.GetConversationsForUserParameters) ([]slack.Channel, string, error)
	getUserInfoFunc             func(string) (*slack.User, error)
	getConversationHistoryFunc  func(*slack.GetConversationHistoryParameters) (*slack.GetConversationHistoryResponse, error)
	getConversationRepliesFunc  func(*slack.GetConversationRepliesParameters) ([]slack.Message, bool, string, error)
}

func (m *mockSlackAPI) GetConversationsForUser(params *slack.GetConversationsForUserParameters) ([]slack.Channel, string, error) {
	if m.GetConversationsForUserFunc != nil {
		return m.GetConversationsForUserFunc(params)
	}
	return nil, "", fmt.Errorf("not implemented")
}

func (m *mockSlackAPI) GetUserInfo(userID string) (*slack.User, error) {
	if m.getUserInfoFunc != nil {
		return m.getUserInfoFunc(userID)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockSlackAPI) GetConversationHistory(params *slack.GetConversationHistoryParameters) (*slack.GetConversationHistoryResponse, error) {
	if m.getConversationHistoryFunc != nil {
		return m.getConversationHistoryFunc(params)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *mockSlackAPI) GetConversationReplies(params *slack.GetConversationRepliesParameters) ([]slack.Message, bool, string, error) {
	if m.getConversationRepliesFunc != nil {
		return m.getConversationRepliesFunc(params)
	}
	return nil, false, "", fmt.Errorf("not implemented")
}

func TestFetchAllChannels(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(*mockSlackAPI)
		want      []Channel
		wantErr   bool
		errMsg    string
	}{
		{
			name: "single page of channels",
			mockSetup: func(m *mockSlackAPI) {
				m.GetConversationsForUserFunc = func(params *slack.GetConversationsForUserParameters) ([]slack.Channel, string, error) {
					return []slack.Channel{
						{GroupConversation: slack.GroupConversation{Name: "general", Conversation: slack.Conversation{ID: "C001"}}},
						{GroupConversation: slack.GroupConversation{Name: "random", Conversation: slack.Conversation{ID: "C002"}}},
					}, "", nil
				}
			},
			want: []Channel{
				{ID: "C001", Name: "general"},
				{ID: "C002", Name: "random"},
			},
			wantErr: false,
		},
		{
			name: "multiple pages of channels",
			mockSetup: func(m *mockSlackAPI) {
				callCount := 0
				m.GetConversationsForUserFunc = func(params *slack.GetConversationsForUserParameters) ([]slack.Channel, string, error) {
					callCount++
					if callCount == 1 {
						return []slack.Channel{
							{GroupConversation: slack.GroupConversation{Name: "page1-ch1", Conversation: slack.Conversation{ID: "C001"}}},
							{GroupConversation: slack.GroupConversation{Name: "page1-ch2", Conversation: slack.Conversation{ID: "C002"}}},
						}, "cursor1", nil
					}
					return []slack.Channel{
						{GroupConversation: slack.GroupConversation{Name: "page2-ch1", Conversation: slack.Conversation{ID: "C003"}}},
						{GroupConversation: slack.GroupConversation{Name: "page2-ch2", Conversation: slack.Conversation{ID: "C004"}}},
					}, "", nil
				}
			},
			want: []Channel{
				{ID: "C001", Name: "page1-ch1"},
				{ID: "C002", Name: "page1-ch2"},
				{ID: "C003", Name: "page2-ch1"},
				{ID: "C004", Name: "page2-ch2"},
			},
			wantErr: false,
		},
		{
			name: "empty result no channels",
			mockSetup: func(m *mockSlackAPI) {
				m.GetConversationsForUserFunc = func(params *slack.GetConversationsForUserParameters) ([]slack.Channel, string, error) {
					return []slack.Channel{}, "", nil
				}
			},
			want:    []Channel{},
			wantErr: false,
		},
		{
			name: "exactly 200 channels (boundary test)",
			mockSetup: func(m *mockSlackAPI) {
				m.GetConversationsForUserFunc = func(params *slack.GetConversationsForUserParameters) ([]slack.Channel, string, error) {
					channels := make([]slack.Channel, 200)
					for i := 0; i < 200; i++ {
						channels[i] = slack.Channel{
							GroupConversation: slack.GroupConversation{
								Name: fmt.Sprintf("channel%d", i),
								Conversation: slack.Conversation{
									ID: fmt.Sprintf("C%03d", i),
								},
							},
						}
					}
					return channels, "", nil
				}
			},
			want: func() []Channel {
				channels := make([]Channel, 200)
				for i := 0; i < 200; i++ {
					channels[i] = Channel{
						ID:   fmt.Sprintf("C%03d", i),
						Name: fmt.Sprintf("channel%d", i),
					}
				}
				return channels
			}(),
			wantErr: false,
		},
		{
			name: "three pages of results",
			mockSetup: func(m *mockSlackAPI) {
				callCount := 0
				m.GetConversationsForUserFunc = func(params *slack.GetConversationsForUserParameters) ([]slack.Channel, string, error) {
					callCount++
					switch callCount {
					case 1:
						return []slack.Channel{
							{GroupConversation: slack.GroupConversation{Name: "ch1", Conversation: slack.Conversation{ID: "C001"}}},
						}, "cursor1", nil
					case 2:
						return []slack.Channel{
							{GroupConversation: slack.GroupConversation{Name: "ch2", Conversation: slack.Conversation{ID: "C002"}}},
						}, "cursor2", nil
					default:
						return []slack.Channel{
							{GroupConversation: slack.GroupConversation{Name: "ch3", Conversation: slack.Conversation{ID: "C003"}}},
						}, "", nil
					}
				}
			},
			want: []Channel{
				{ID: "C001", Name: "ch1"},
				{ID: "C002", Name: "ch2"},
				{ID: "C003", Name: "ch3"},
			},
			wantErr: false,
		},
		{
			name: "api error on first page",
			mockSetup: func(m *mockSlackAPI) {
				m.GetConversationsForUserFunc = func(params *slack.GetConversationsForUserParameters) ([]slack.Channel, string, error) {
					return nil, "", fmt.Errorf("api error")
				}
			},
			wantErr: true,
			errMsg:  "failed to fetch conversations",
		},
		{
			name: "api error on second page",
			mockSetup: func(m *mockSlackAPI) {
				callCount := 0
				m.GetConversationsForUserFunc = func(params *slack.GetConversationsForUserParameters) ([]slack.Channel, string, error) {
					callCount++
					if callCount == 1 {
						return []slack.Channel{
							{GroupConversation: slack.GroupConversation{Name: "ch1", Conversation: slack.Conversation{ID: "C001"}}},
						}, "cursor1", nil
					}
					return nil, "", fmt.Errorf("api error on page 2")
				}
			},
			wantErr: true,
			errMsg:  "failed to fetch conversations",
		},
		{
			name: "last page with partial results",
			mockSetup: func(m *mockSlackAPI) {
				callCount := 0
				m.GetConversationsForUserFunc = func(params *slack.GetConversationsForUserParameters) ([]slack.Channel, string, error) {
					callCount++
					if callCount == 1 {
						return []slack.Channel{
							{GroupConversation: slack.GroupConversation{Name: "ch1", Conversation: slack.Conversation{ID: "C001"}}},
							{GroupConversation: slack.GroupConversation{Name: "ch2", Conversation: slack.Conversation{ID: "C002"}}},
						}, "cursor1", nil
					}
					return []slack.Channel{
						{GroupConversation: slack.GroupConversation{Name: "ch3", Conversation: slack.Conversation{ID: "C003"}}},
					}, "", nil
				}
			},
			want: []Channel{
				{ID: "C001", Name: "ch1"},
				{ID: "C002", Name: "ch2"},
				{ID: "C003", Name: "ch3"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockSlackAPI{}
			tt.mockSetup(mock)
			client := NewClientWithAPI(mock)

			channels, err := client.FetchAllChannels()

			if (err != nil) != tt.wantErr {
				t.Errorf("FetchAllChannels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("error message = %q, want to contain %q", err.Error(), tt.errMsg)
			}

			if !tt.wantErr {
				if len(channels) != len(tt.want) {
					t.Errorf("FetchAllChannels() returned %d channels, want %d", len(channels), len(tt.want))
					return
				}
				for i, ch := range channels {
					if ch.ID != tt.want[i].ID || ch.Name != tt.want[i].Name {
						t.Errorf("FetchAllChannels()[%d] = %+v, want %+v", i, ch, tt.want[i])
					}
				}
			}
		})
	}
}

func TestGetUserDisplayName(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		mockSetup func(*mockSlackAPI)
		want      string
	}{
		{
			name:   "user with DisplayName set",
			userID: "U12345",
			mockSetup: func(m *mockSlackAPI) {
				m.getUserInfoFunc = func(userID string) (*slack.User, error) {
					return &slack.User{
						Profile: slack.UserProfile{
							DisplayName: "John Doe",
						},
						RealName: "John A. Doe",
						Name:     "johndoe",
					}, nil
				}
			},
			want: "John Doe",
		},
		{
			name:   "user with RealName only",
			userID: "U12345",
			mockSetup: func(m *mockSlackAPI) {
				m.getUserInfoFunc = func(userID string) (*slack.User, error) {
					return &slack.User{
						Profile: slack.UserProfile{
							DisplayName: "",
						},
						RealName: "Jane Smith",
						Name:     "janesmith",
					}, nil
				}
			},
			want: "Jane Smith",
		},
		{
			name:   "user with Name (username) only",
			userID: "U12345",
			mockSetup: func(m *mockSlackAPI) {
				m.getUserInfoFunc = func(userID string) (*slack.User, error) {
					return &slack.User{
						Profile: slack.UserProfile{
							DisplayName: "",
							RealName:    "",
						},
						Name: "bobuser",
					}, nil
				}
			},
			want: "bobuser",
		},
		{
			name:   "user with all three names - prefers DisplayName",
			userID: "U12345",
			mockSetup: func(m *mockSlackAPI) {
				m.getUserInfoFunc = func(userID string) (*slack.User, error) {
					return &slack.User{
						Profile: slack.UserProfile{
							DisplayName: "Display",
						},
						RealName: "Real",
						Name:     "username",
					}, nil
				}
			},
			want: "Display",
		},
		{
			name:   "user with DisplayName and RealName - prefers DisplayName",
			userID: "U12345",
			mockSetup: func(m *mockSlackAPI) {
				m.getUserInfoFunc = func(userID string) (*slack.User, error) {
					return &slack.User{
						Profile: slack.UserProfile{
							DisplayName: "Preferred Name",
						},
						RealName: "Legal Name",
						Name:     "",
					}, nil
				}
			},
			want: "Preferred Name",
		},
		{
			name:   "user with RealName and Name - prefers RealName",
			userID: "U12345",
			mockSetup: func(m *mockSlackAPI) {
				m.getUserInfoFunc = func(userID string) (*slack.User, error) {
					return &slack.User{
						Profile: slack.UserProfile{
							DisplayName: "",
						},
						RealName: "Alice Brown",
						Name:     "aliceb",
					}, nil
				}
			},
			want: "Alice Brown",
		},
		{
			name:   "empty user ID returns Unknown",
			userID: "",
			mockSetup: func(m *mockSlackAPI) {
				// Should not be called
			},
			want: "Unknown",
		},
		{
			name:   "api error returns userID",
			userID: "U12345",
			mockSetup: func(m *mockSlackAPI) {
				m.getUserInfoFunc = func(userID string) (*slack.User, error) {
					return nil, fmt.Errorf("user not found")
				}
			},
			want: "U12345",
		},
		{
			name:   "user not found returns userID",
			userID: "U99999",
			mockSetup: func(m *mockSlackAPI) {
				m.getUserInfoFunc = func(userID string) (*slack.User, error) {
					return nil, fmt.Errorf("user_not_found")
				}
			},
			want: "U99999",
		},
		{
			name:   "user with all empty fields returns userID",
			userID: "U12345",
			mockSetup: func(m *mockSlackAPI) {
				m.getUserInfoFunc = func(userID string) (*slack.User, error) {
					return &slack.User{
						Profile: slack.UserProfile{
							DisplayName: "",
							RealName:    "",
						},
						Name: "",
					}, nil
				}
			},
			want: "U12345",
		},
		{
			name:   "DisplayName with special characters",
			userID: "U12345",
			mockSetup: func(m *mockSlackAPI) {
				m.getUserInfoFunc = func(userID string) (*slack.User, error) {
					return &slack.User{
						Profile: slack.UserProfile{
							DisplayName: "John-Paul O'Brien",
						},
					}, nil
				}
			},
			want: "John-Paul O'Brien",
		},
		{
			name:   "DisplayName with emojis",
			userID: "U12345",
			mockSetup: func(m *mockSlackAPI) {
				m.getUserInfoFunc = func(userID string) (*slack.User, error) {
					return &slack.User{
						Profile: slack.UserProfile{
							DisplayName: "Sarah 🚀",
						},
					}, nil
				}
			},
			want: "Sarah 🚀",
		},
		{
			name:   "very long DisplayName",
			userID: "U12345",
			mockSetup: func(m *mockSlackAPI) {
				m.getUserInfoFunc = func(userID string) (*slack.User, error) {
					return &slack.User{
						Profile: slack.UserProfile{
							DisplayName: "This Is A Very Long Display Name That Someone Might Use",
						},
					}, nil
				}
			},
			want: "This Is A Very Long Display Name That Someone Might Use",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockSlackAPI{}
			tt.mockSetup(mock)
			client := NewClientWithAPI(mock)

			result := client.getUserDisplayName(tt.userID)

			if result != tt.want {
				t.Errorf("GetUserDisplayName() = %q, want %q", result, tt.want)
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name          string
		botToken      string
		appLevelToken string
		expectNonNil  bool
		expectAPISet  bool
	}{
		{
			name:          "NewClient with both tokens",
			botToken:      "xoxb-test-token",
			appLevelToken: "xapp-test-token",
			expectNonNil:  true,
			expectAPISet:  true,
		},
		{
			name:          "NewClient with bot token only",
			botToken:      "xoxb-test-token",
			appLevelToken: "",
			expectNonNil:  true,
			expectAPISet:  true,
		},
		{
			name:          "NewClient with empty tokens",
			botToken:      "",
			appLevelToken: "",
			expectNonNil:  true,
			expectAPISet:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.botToken, tt.appLevelToken)

			if tt.expectNonNil && client == nil {
				t.Error("NewClient() returned nil, want non-nil")
			}

			if tt.expectAPISet && client != nil && client.api == nil {
				t.Error("NewClient() client.api is nil, want non-nil")
			}
		})
	}
}

func TestNewClientWithAPI(t *testing.T) {
	tests := []struct {
		name         string
		mockAPI      SlackAPI
		expectNonNil bool
	}{
		{
			name:         "NewClientWithAPI with mock",
			mockAPI:      &mockSlackAPI{},
			expectNonNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClientWithAPI(tt.mockAPI)

			if tt.expectNonNil && client == nil {
				t.Error("NewClientWithAPI() returned nil, want non-nil")
			}

			if client != nil && client.api != tt.mockAPI {
				t.Error("NewClientWithAPI() client.api not set to provided mock")
			}
		})
	}
}
