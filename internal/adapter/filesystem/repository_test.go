package filesystem

import (
	"testing"
	"time"

	"github.com/cppcho/slack2md/internal/domain/entities"
	"github.com/cppcho/slack2md/internal/adapter/logger"
)

func TestNewFileRepository(t *testing.T) {
	writer := NewFileSystemWriter()
	log := logger.NewLogger(logger.INFO)

	repo := NewFileRepository(writer, log)
	if repo == nil {
		t.Error("Expected repository to be created")
	}
	if repo.writer == nil {
		t.Error("Expected writer to be set")
	}
	if repo.logger == nil {
		t.Error("Expected logger to be set")
	}
}

func TestFileRepository_SanitizeChannelName(t *testing.T) {
	writer := NewFileSystemWriter()
	log := logger.NewLogger(logger.ERROR)
	repo := NewFileRepository(writer, log)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "alphanumeric preserved",
			input:    "channel123",
			expected: "channel123",
		},
		{
			name:     "hyphens and underscores preserved",
			input:    "my-channel_name",
			expected: "my-channel_name",
		},
		{
			name:     "spaces replaced",
			input:    "my channel",
			expected: "my_channel",
		},
		{
			name:     "special chars replaced",
			input:    "channel@#$%name",
			expected: "channel____name",
		},
		{
			name:     "mixed case preserved",
			input:    "MyChannel",
			expected: "MyChannel",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := repo.sanitizeChannelName(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeChannelName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFileRepository_IsSameDay(t *testing.T) {
	writer := NewFileSystemWriter()
	log := logger.NewLogger(logger.ERROR)
	repo := NewFileRepository(writer, log)

	t1 := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 15, 14, 45, 0, 0, time.UTC)
	t3 := time.Date(2024, 1, 16, 10, 30, 0, 0, time.UTC)

	if !repo.isSameDay(t1, t2) {
		t.Error("Expected same day for times on same date")
	}

	if repo.isSameDay(t1, t3) {
		t.Error("Expected different day for times on different dates")
	}
}

func TestFileRepository_GetSortedDates(t *testing.T) {
	writer := NewFileSystemWriter()
	log := logger.NewLogger(logger.ERROR)
	repo := NewFileRepository(writer, log)

	messagesByDate := map[string][]entities.Message{
		"2024-01-20": {},
		"2024-01-15": {},
		"2024-01-18": {},
	}

	sorted := repo.getSortedDates(messagesByDate)

	expected := []string{"2024-01-15", "2024-01-18", "2024-01-20"}
	if len(sorted) != len(expected) {
		t.Fatalf("Expected %d dates, got %d", len(expected), len(sorted))
	}

	for i, date := range sorted {
		if date != expected[i] {
			t.Errorf("Expected date %s at position %d, got %s", expected[i], i, date)
		}
	}
}
