package filewriter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestIsSameDay tests the isSameDay function
func TestIsSameDay(t *testing.T) {
	tests := []struct {
		name     string
		t1       time.Time
		t2       time.Time
		expected bool
	}{
		{
			name:     "same date different times",
			t1:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			t2:       time.Date(2024, 1, 15, 18, 45, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "same date and time",
			t1:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			t2:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "different dates same time",
			t1:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			t2:       time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC),
			expected: false,
		},
		{
			name:     "different dates different times",
			t1:       time.Date(2024, 1, 15, 23, 59, 59, 0, time.UTC),
			t2:       time.Date(2024, 1, 16, 0, 0, 1, 0, time.UTC),
			expected: false,
		},
		{
			name:     "midnight boundary - 1 second apart",
			t1:       time.Date(2024, 1, 15, 23, 59, 59, 0, time.UTC),
			t2:       time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC),
			expected: false,
		},
		{
			name:     "different months same day number",
			t1:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			t2:       time.Date(2024, 2, 15, 10, 30, 0, 0, time.UTC),
			expected: false,
		},
		{
			name:     "different years same month and day",
			t1:       time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC),
			t2:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			expected: false,
		},
		{
			name:     "different timezones same UTC day",
			t1:       time.Date(2024, 1, 15, 22, 0, 0, 0, time.UTC),
			t2:       time.Date(2024, 1, 15, 14, 0, 0, 0, time.FixedZone("PST", -8*3600)),
			expected: true,
		},
		{
			name:     "zero time values",
			t1:       time.Time{},
			t2:       time.Time{},
			expected: true,
		},
		{
			name:     "one zero time value",
			t1:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			t2:       time.Time{},
			expected: false,
		},
		{
			name:     "start and end of same day",
			t1:       time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			t2:       time.Date(2024, 1, 15, 23, 59, 59, 999999999, time.UTC),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSameDay(tt.t1, tt.t2)
			if result != tt.expected {
				t.Errorf("isSameDay(%v, %v) = %v, want %v", tt.t1, tt.t2, result, tt.expected)
			}
		})
	}
}

// TestSanitizeChannelName tests the SanitizeChannelName function
func TestSanitizeChannelName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "alphanumeric only",
			input:    "general",
			expected: "general",
		},
		{
			name:     "with hyphen",
			input:    "dev-team",
			expected: "dev-team",
		},
		{
			name:     "with underscore",
			input:    "dev_team",
			expected: "dev_team",
		},
		{
			name:     "with spaces",
			input:    "general discussion",
			expected: "general_discussion",
		},
		{
			name:     "with multiple spaces",
			input:    "general   discussion   channel",
			expected: "general___discussion___channel",
		},
		{
			name:     "with slash",
			input:    "team/project",
			expected: "team_project",
		},
		{
			name:     "with hash",
			input:    "#general",
			expected: "_general",
		},
		{
			name:     "with special characters",
			input:    "team@work!channel",
			expected: "team_work_channel",
		},
		{
			name:     "with parentheses",
			input:    "dev(backend)",
			expected: "dev_backend_",
		},
		{
			name:     "with brackets",
			input:    "team[important]",
			expected: "team_important_",
		},
		{
			name:     "with dots",
			input:    "v1.0.release",
			expected: "v1_0_release",
		},
		{
			name:     "with commas",
			input:    "team,project,alpha",
			expected: "team_project_alpha",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only special characters",
			input:    "@#$%",
			expected: "____",
		},
		{
			name:     "mixed valid and invalid",
			input:    "dev-team_2024@company.com",
			expected: "dev-team_2024_company_com",
		},
		{
			name:     "unicode characters",
			input:    "チーム",
			expected: "___",
		},
		{
			name:     "emoji",
			input:    "team🚀rocket",
			expected: "team_rocket",
		},
		{
			name:     "very long name",
			input:    "this-is-a-very-long-channel-name-with-many-characters",
			expected: "this-is-a-very-long-channel-name-with-many-characters",
		},
		{
			name:     "numbers only",
			input:    "123456",
			expected: "123456",
		},
		{
			name:     "leading and trailing special chars",
			input:    "@team-name!",
			expected: "_team-name_",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeChannelName(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeChannelName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestSortDates tests the sortDates function
func TestSortDates(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "already sorted",
			input:    []string{"2024-01-15", "2024-01-16", "2024-01-17"},
			expected: []string{"2024-01-15", "2024-01-16", "2024-01-17"},
		},
		{
			name:     "reverse order",
			input:    []string{"2024-01-17", "2024-01-16", "2024-01-15"},
			expected: []string{"2024-01-15", "2024-01-16", "2024-01-17"},
		},
		{
			name:     "mixed order",
			input:    []string{"2024-01-16", "2024-01-15", "2024-01-18", "2024-01-17"},
			expected: []string{"2024-01-15", "2024-01-16", "2024-01-17", "2024-01-18"},
		},
		{
			name:     "single date",
			input:    []string{"2024-01-15"},
			expected: []string{"2024-01-15"},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "different months",
			input:    []string{"2024-03-15", "2024-01-15", "2024-02-15"},
			expected: []string{"2024-01-15", "2024-02-15", "2024-03-15"},
		},
		{
			name:     "different years",
			input:    []string{"2025-01-15", "2023-12-31", "2024-06-15"},
			expected: []string{"2023-12-31", "2024-06-15", "2025-01-15"},
		},
		{
			name:     "year boundary",
			input:    []string{"2024-01-01", "2023-12-31"},
			expected: []string{"2023-12-31", "2024-01-01"},
		},
		{
			name:     "two dates same",
			input:    []string{"2024-01-15", "2024-01-15"},
			expected: []string{"2024-01-15", "2024-01-15"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid modifying test data
			input := make([]string, len(tt.input))
			copy(input, tt.input)

			sortDates(input)

			if len(input) != len(tt.expected) {
				t.Errorf("sortDates() result length = %d, want %d", len(input), len(tt.expected))
				return
			}

			for i := range input {
				if input[i] != tt.expected[i] {
					t.Errorf("sortDates() result[%d] = %q, want %q", i, input[i], tt.expected[i])
				}
			}
		})
	}
}

// TestGetSortedDates tests the getSortedDates function
func TestGetSortedDates(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string][]Message
		expected []string
	}{
		{
			name: "multiple dates unsorted",
			input: map[string][]Message{
				"2024-01-17": {},
				"2024-01-15": {},
				"2024-01-16": {},
			},
			expected: []string{"2024-01-15", "2024-01-16", "2024-01-17"},
		},
		{
			name:     "empty map",
			input:    map[string][]Message{},
			expected: []string{},
		},
		{
			name: "single date",
			input: map[string][]Message{
				"2024-01-15": {},
			},
			expected: []string{"2024-01-15"},
		},
		{
			name: "dates across months",
			input: map[string][]Message{
				"2024-03-10": {},
				"2024-01-25": {},
				"2024-02-14": {},
			},
			expected: []string{"2024-01-25", "2024-02-14", "2024-03-10"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getSortedDates(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("getSortedDates() length = %d, want %d", len(result), len(tt.expected))
				return
			}

			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("getSortedDates() result[%d] = %q, want %q", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

// TestWriteMarkdownFile tests the file I/O functionality
func TestWriteMarkdownFile(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()

	tests := []struct {
		name         string
		filePath     string
		messages     []Message
		date         string
		parentDate   time.Time
		wantErr      bool
		validateFile func(t *testing.T, content string)
	}{
		{
			name:     "single message",
			filePath: filepath.Join(tmpDir, "test1", "2024-01-15.md"),
			messages: []Message{
				{
					Timestamp:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					Text:            "Hello world",
					UserDisplayName: "User1",
				},
			},
			date:       "2024-01-15",
			parentDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			wantErr:    false,
			validateFile: func(t *testing.T, content string) {
				if !strings.Contains(content, "# 2024-01-15") {
					t.Error("Expected header with date")
				}
				if !strings.Contains(content, "### 10:30 User1") {
					t.Error("Expected message timestamp and user")
				}
				if !strings.Contains(content, "Hello world") {
					t.Error("Expected message text")
				}
			},
		},
		{
			name:     "message with thread replies same day",
			filePath: filepath.Join(tmpDir, "test2", "2024-01-15.md"),
			messages: []Message{
				{
					Timestamp:       time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					Text:            "Parent message",
					UserDisplayName: "User1",
					IsParent:        true,
					Replies: []Message{
						{
							Timestamp:       time.Date(2024, 1, 15, 10, 35, 0, 0, time.UTC),
							Text:            "Reply 1",
							UserDisplayName: "User2",
						},
					},
				},
			},
			date:       "2024-01-15",
			parentDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			wantErr:    false,
			validateFile: func(t *testing.T, content string) {
				if !strings.Contains(content, "### 10:30 User1") {
					t.Error("Expected parent message")
				}
				if !strings.Contains(content, "Parent message") {
					t.Error("Expected parent message text")
				}
				if !strings.Contains(content, "###### 10:35 User2") {
					t.Error("Expected reply with time only (same day)")
				}
				if !strings.Contains(content, "Reply 1") {
					t.Error("Expected reply text")
				}
			},
		},
		{
			name:     "message with thread reply different day",
			filePath: filepath.Join(tmpDir, "test3", "2024-01-15.md"),
			messages: []Message{
				{
					Timestamp:       time.Date(2024, 1, 15, 23, 55, 0, 0, time.UTC),
					Text:            "Late night message",
					UserDisplayName: "User1",
					IsParent:        true,
					Replies: []Message{
						{
							Timestamp:       time.Date(2024, 1, 16, 0, 5, 0, 0, time.UTC),
							Text:            "Next day reply",
							UserDisplayName: "User2",
						},
					},
				},
			},
			date:       "2024-01-15",
			parentDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			wantErr:    false,
			validateFile: func(t *testing.T, content string) {
				if !strings.Contains(content, "### 23:55 User1") {
					t.Error("Expected parent message")
				}
				if !strings.Contains(content, "###### 2024-01-16 00:05 User2") {
					t.Error("Expected reply with full date (different day)")
				}
			},
		},
		{
			name:     "multiple messages",
			filePath: filepath.Join(tmpDir, "test4", "2024-01-15.md"),
			messages: []Message{
				{
					Timestamp:       time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC),
					Text:            "First message",
					UserDisplayName: "User1",
				},
				{
					Timestamp:       time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC),
					Text:            "Second message",
					UserDisplayName: "User2",
				},
			},
			date:       "2024-01-15",
			parentDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			wantErr:    false,
			validateFile: func(t *testing.T, content string) {
				if !strings.Contains(content, "First message") {
					t.Error("Expected first message")
				}
				if !strings.Contains(content, "Second message") {
					t.Error("Expected second message")
				}
				if !strings.Contains(content, "### 09:00 User1") {
					t.Error("Expected first message timestamp")
				}
				if !strings.Contains(content, "### 14:30 User2") {
					t.Error("Expected second message timestamp")
				}
			},
		},
		{
			name:       "empty messages",
			filePath:   filepath.Join(tmpDir, "test5", "2024-01-15.md"),
			messages:   []Message{},
			date:       "2024-01-15",
			parentDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			wantErr:    false,
			validateFile: func(t *testing.T, content string) {
				if !strings.Contains(content, "# 2024-01-15") {
					t.Error("Expected header even with no messages")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WriteMarkdownFile(tt.filePath, tt.messages, tt.date, tt.parentDate)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("WriteMarkdownFile() unexpected error: %v", err)
			}

			// Read and validate file content
			content, err := os.ReadFile(tt.filePath)
			if err != nil {
				t.Fatalf("Failed to read test file: %v", err)
			}

			if tt.validateFile != nil {
				tt.validateFile(t, string(content))
			}
		})
	}
}

// TestWriteChannelMessages tests the full channel writing functionality
func TestWriteChannelMessages(t *testing.T) {
	tmpDir := t.TempDir()

	messagesByDate := map[string][]Message{
		"2024-01-15": {
			{
				Timestamp:       time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
				Text:            "Message on Jan 15",
				UserDisplayName: "User1",
			},
		},
		"2024-01-16": {
			{
				Timestamp:       time.Date(2024, 1, 16, 11, 0, 0, 0, time.UTC),
				Text:            "Message on Jan 16",
				UserDisplayName: "User2",
			},
		},
	}

	err := WriteChannelMessages(tmpDir, "general-chat", messagesByDate)
	if err != nil {
		t.Fatalf("WriteChannelMessages() error: %v", err)
	}

	// Verify directory structure
	channelDir := filepath.Join(tmpDir, "general-chat")
	if _, err := os.Stat(channelDir); os.IsNotExist(err) {
		t.Error("Channel directory not created")
	}

	// Verify files exist
	file1 := filepath.Join(channelDir, "2024-01-15.md")
	file2 := filepath.Join(channelDir, "2024-01-16.md")

	if _, err := os.Stat(file1); os.IsNotExist(err) {
		t.Error("File for 2024-01-15 not created")
	}
	if _, err := os.Stat(file2); os.IsNotExist(err) {
		t.Error("File for 2024-01-16 not created")
	}

	// Verify content
	content1, _ := os.ReadFile(file1)
	if !strings.Contains(string(content1), "Message on Jan 15") {
		t.Error("File for 2024-01-15 missing expected content")
	}

	content2, _ := os.ReadFile(file2)
	if !strings.Contains(string(content2), "Message on Jan 16") {
		t.Error("File for 2024-01-16 missing expected content")
	}
}

// TestWriteChannelMessages_SanitizesName verifies channel name sanitization
func TestWriteChannelMessages_SanitizesName(t *testing.T) {
	tmpDir := t.TempDir()

	messagesByDate := map[string][]Message{
		"2024-01-15": {
			{
				Timestamp:       time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
				Text:            "Test message",
				UserDisplayName: "User1",
			},
		},
	}

	// Use a channel name with special characters
	err := WriteChannelMessages(tmpDir, "team@work!channel", messagesByDate)
	if err != nil {
		t.Fatalf("WriteChannelMessages() error: %v", err)
	}

	// Verify sanitized directory name
	sanitizedDir := filepath.Join(tmpDir, "team_work_channel")
	if _, err := os.Stat(sanitizedDir); os.IsNotExist(err) {
		t.Error("Sanitized channel directory not created")
	}
}
