package entities

import (
	"errors"
	"testing"
)

func TestExportResult_Success(t *testing.T) {
	channel, _ := NewChannel("C123", "general")

	result := ExportResult{
		Channel:      *channel,
		MessageCount: 42,
		DateCount:    3,
		Success:      true,
		Error:        nil,
	}

	if !result.Success {
		t.Error("Expected Success to be true")
	}
	if result.Error != nil {
		t.Errorf("Expected Error to be nil, got %v", result.Error)
	}
	if result.Channel.ID != "C123" {
		t.Errorf("Expected Channel.ID = C123, got %s", result.Channel.ID)
	}
	if result.MessageCount != 42 {
		t.Errorf("Expected MessageCount = 42, got %d", result.MessageCount)
	}
	if result.DateCount != 3 {
		t.Errorf("Expected DateCount = 3, got %d", result.DateCount)
	}
}

func TestExportResult_Failure(t *testing.T) {
	channel, _ := NewChannel("C456", "random")
	expectedErr := errors.New("failed to export")

	result := ExportResult{
		Channel:      *channel,
		MessageCount: 0,
		DateCount:    0,
		Success:      false,
		Error:        expectedErr,
	}

	if result.Success {
		t.Error("Expected Success to be false")
	}
	if result.Error == nil {
		t.Error("Expected Error to be non-nil")
	}
	if result.Error.Error() != "failed to export" {
		t.Errorf("Expected Error message = %q, got %q", "failed to export", result.Error.Error())
	}
	if result.MessageCount != 0 {
		t.Errorf("Expected MessageCount = 0, got %d", result.MessageCount)
	}
}

func TestExportSummary_Empty(t *testing.T) {
	summary := ExportSummary{
		TotalChannels: 0,
		SuccessCount:  0,
		FailureCount:  0,
		Results:       []ExportResult{},
	}

	if summary.TotalChannels != 0 {
		t.Errorf("Expected TotalChannels = 0, got %d", summary.TotalChannels)
	}
	if summary.SuccessCount != 0 {
		t.Errorf("Expected SuccessCount = 0, got %d", summary.SuccessCount)
	}
	if summary.FailureCount != 0 {
		t.Errorf("Expected FailureCount = 0, got %d", summary.FailureCount)
	}
	if len(summary.Results) != 0 {
		t.Errorf("Expected Results to be empty, got %d results", len(summary.Results))
	}
}

func TestExportSummary_WithResults(t *testing.T) {
	channel1, _ := NewChannel("C123", "general")
	channel2, _ := NewChannel("C456", "random")
	channel3, _ := NewChannel("C789", "dev")

	results := []ExportResult{
		{
			Channel:      *channel1,
			MessageCount: 50,
			DateCount:    5,
			Success:      true,
			Error:        nil,
		},
		{
			Channel:      *channel2,
			MessageCount: 30,
			DateCount:    3,
			Success:      true,
			Error:        nil,
		},
		{
			Channel:      *channel3,
			MessageCount: 0,
			DateCount:    0,
			Success:      false,
			Error:        errors.New("export failed"),
		},
	}

	summary := ExportSummary{
		TotalChannels: 3,
		SuccessCount:  2,
		FailureCount:  1,
		Results:       results,
	}

	if summary.TotalChannels != 3 {
		t.Errorf("Expected TotalChannels = 3, got %d", summary.TotalChannels)
	}
	if summary.SuccessCount != 2 {
		t.Errorf("Expected SuccessCount = 2, got %d", summary.SuccessCount)
	}
	if summary.FailureCount != 1 {
		t.Errorf("Expected FailureCount = 1, got %d", summary.FailureCount)
	}
	if len(summary.Results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(summary.Results))
	}
}

func TestExportSummary_CountsMatchResults(t *testing.T) {
	tests := []struct {
		name          string
		successCount  int
		failureCount  int
		results       []ExportResult
		wantConsistent bool
	}{
		{
			name:          "consistent counts",
			successCount:  2,
			failureCount:  1,
			results:       make([]ExportResult, 3),
			wantConsistent: true,
		},
		{
			name:          "total matches results",
			successCount:  5,
			failureCount:  0,
			results:       make([]ExportResult, 5),
			wantConsistent: true,
		},
		{
			name:          "all failures",
			successCount:  0,
			failureCount:  3,
			results:       make([]ExportResult, 3),
			wantConsistent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := ExportSummary{
				TotalChannels: tt.successCount + tt.failureCount,
				SuccessCount:  tt.successCount,
				FailureCount:  tt.failureCount,
				Results:       tt.results,
			}

			totalFromCounts := summary.SuccessCount + summary.FailureCount
			if summary.TotalChannels != totalFromCounts {
				t.Errorf("TotalChannels (%d) != SuccessCount + FailureCount (%d)",
					summary.TotalChannels, totalFromCounts)
			}

			if summary.TotalChannels != len(summary.Results) {
				t.Errorf("TotalChannels (%d) != len(Results) (%d)",
					summary.TotalChannels, len(summary.Results))
			}
		})
	}
}
