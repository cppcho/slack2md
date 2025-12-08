package valueobjects

import (
	"testing"
	"time"
)

func TestNewExportConfig(t *testing.T) {
	now := time.Now()
	timeRange := TimeRange{
		Start: now.AddDate(0, 0, -7),
		End:   now,
	}

	tests := []struct {
		name       string
		exportPath string
		timeRange  TimeRange
		wantErr    bool
	}{
		{
			name:       "valid config",
			exportPath: "/tmp/export",
			timeRange:  timeRange,
			wantErr:    false,
		},
		{
			name:       "valid config with relative path",
			exportPath: "./export",
			timeRange:  timeRange,
			wantErr:    false,
		},
		{
			name:       "valid config with home path",
			exportPath: "~/exports",
			timeRange:  timeRange,
			wantErr:    false,
		},
		{
			name:       "empty export path",
			exportPath: "",
			timeRange:  timeRange,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := NewExportConfig(tt.exportPath, tt.timeRange)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewExportConfig() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if cfg == nil {
					t.Error("Expected ExportConfig to be created, got nil")
					return
				}
				if cfg.ExportPath != tt.exportPath {
					t.Errorf("Expected ExportPath = %q, got = %q", tt.exportPath, cfg.ExportPath)
				}
				if !cfg.TimeRange.Start.Equal(tt.timeRange.Start) {
					t.Errorf("Expected TimeRange.Start = %v, got = %v",
						tt.timeRange.Start, cfg.TimeRange.Start)
				}
				if !cfg.TimeRange.End.Equal(tt.timeRange.End) {
					t.Errorf("Expected TimeRange.End = %v, got = %v",
						tt.timeRange.End, cfg.TimeRange.End)
				}
			} else {
				if cfg != nil {
					t.Error("Expected nil ExportConfig on error, got non-nil")
				}
			}
		})
	}
}

func TestNewExportConfig_ErrorMessage(t *testing.T) {
	timeRange := TimeRange{
		Start: time.Now().AddDate(0, 0, -7),
		End:   time.Now(),
	}

	_, err := NewExportConfig("", timeRange)
	if err == nil {
		t.Fatal("Expected error for empty export path, got nil")
	}

	expectedMsg := "export path cannot be empty"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message = %q, got = %q", expectedMsg, err.Error())
	}
}

func TestNewExportConfig_WithZeroTimeRange(t *testing.T) {
	// Zero value TimeRange (both Start and End are zero time)
	zeroTimeRange := TimeRange{}

	cfg, err := NewExportConfig("/tmp/export", zeroTimeRange)
	if err != nil {
		t.Errorf("Expected no error with zero TimeRange, got %v", err)
	}

	if cfg == nil {
		t.Fatal("Expected ExportConfig to be created")
	}

	// Verify zero times are preserved
	if !cfg.TimeRange.Start.IsZero() {
		t.Error("Expected TimeRange.Start to be zero time")
	}
	if !cfg.TimeRange.End.IsZero() {
		t.Error("Expected TimeRange.End to be zero time")
	}
}

func TestNewExportConfig_PreservesTimeRange(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)
	timeRange := TimeRange{
		Start: start,
		End:   end,
	}

	cfg, err := NewExportConfig("/tmp/export", timeRange)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify exact times are preserved
	if !cfg.TimeRange.Start.Equal(start) {
		t.Errorf("Start time not preserved: expected %v, got %v", start, cfg.TimeRange.Start)
	}
	if !cfg.TimeRange.End.Equal(end) {
		t.Errorf("End time not preserved: expected %v, got %v", end, cfg.TimeRange.End)
	}
}

func TestNewExportConfig_WithSpecialPaths(t *testing.T) {
	timeRange := TimeRange{
		Start: time.Now().AddDate(0, 0, -7),
		End:   time.Now(),
	}

	specialPaths := []string{
		"/",
		".",
		"..",
		"/var/log/exports",
		"C:\\Users\\test\\exports", // Windows path
		"exports/sub/deep/path",
		"./relative/path",
		"../parent/path",
	}

	for _, path := range specialPaths {
		t.Run("path="+path, func(t *testing.T) {
			cfg, err := NewExportConfig(path, timeRange)
			if err != nil {
				t.Errorf("Expected no error for path %q, got %v", path, err)
			}
			if cfg == nil {
				t.Error("Expected ExportConfig to be created")
			}
			if cfg != nil && cfg.ExportPath != path {
				t.Errorf("Expected path to be preserved as %q, got %q", path, cfg.ExportPath)
			}
		})
	}
}
