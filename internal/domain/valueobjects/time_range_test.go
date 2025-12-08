package valueobjects

import (
	"testing"
	"time"
)

func TestNewTimeRange(t *testing.T) {
	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour)
	oneHourLater := now.Add(1 * time.Hour)

	tests := []struct {
		name    string
		start   time.Time
		end     time.Time
		wantErr bool
	}{
		{
			name:    "valid range - end after start",
			start:   oneHourAgo,
			end:     oneHourLater,
			wantErr: false,
		},
		{
			name:    "valid range - same time (edge case)",
			start:   now,
			end:     now,
			wantErr: false,
		},
		{
			name:    "invalid range - end before start",
			start:   oneHourLater,
			end:     oneHourAgo,
			wantErr: true,
		},
		{
			name:    "valid range - one second difference",
			start:   now,
			end:     now.Add(1 * time.Second),
			wantErr: false,
		},
		{
			name:    "valid range - large difference",
			start:   now.AddDate(-1, 0, 0), // 1 year ago
			end:     now,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr, err := NewTimeRange(tt.start, tt.end)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewTimeRange() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tr == nil {
					t.Error("Expected TimeRange to be created, got nil")
					return
				}
				if !tr.Start.Equal(tt.start) {
					t.Errorf("Expected Start = %v, got = %v", tt.start, tr.Start)
				}
				if !tr.End.Equal(tt.end) {
					t.Errorf("Expected End = %v, got = %v", tt.end, tr.End)
				}
			} else {
				if tr != nil {
					t.Error("Expected nil TimeRange on error, got non-nil")
				}
			}
		})
	}
}

func TestNewTimeRange_ErrorMessage(t *testing.T) {
	now := time.Now()
	later := now.Add(1 * time.Hour)

	_, err := NewTimeRange(later, now)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expectedMsg := "end time must be after start time"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message = %q, got = %q", expectedMsg, err.Error())
	}
}

func TestNewTimeRangeFromDaysBack(t *testing.T) {
	tests := []struct {
		name     string
		daysBack int
		wantErr  bool
	}{
		{
			name:     "valid - 1 day",
			daysBack: 1,
			wantErr:  false,
		},
		{
			name:     "valid - 7 days",
			daysBack: 7,
			wantErr:  false,
		},
		{
			name:     "valid - 30 days",
			daysBack: 30,
			wantErr:  false,
		},
		{
			name:     "valid - 90 days",
			daysBack: 90,
			wantErr:  false,
		},
		{
			name:     "valid - 365 days",
			daysBack: 365,
			wantErr:  false,
		},
		{
			name:     "valid - 0 days (today)",
			daysBack: 0,
			wantErr:  false,
		},
		{
			name:     "invalid - negative days",
			daysBack: -1,
			wantErr:  true,
		},
		{
			name:     "invalid - large negative",
			daysBack: -100,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeCall := time.Now()
			tr, err := NewTimeRangeFromDaysBack(tt.daysBack)
			afterCall := time.Now()

			if (err != nil) != tt.wantErr {
				t.Errorf("NewTimeRangeFromDaysBack() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tr == nil {
					t.Error("Expected TimeRange to be created, got nil")
					return
				}

				// Verify End is approximately "now"
				if tr.End.Before(beforeCall) || tr.End.After(afterCall) {
					t.Errorf("Expected End to be between %v and %v, got %v",
						beforeCall, afterCall, tr.End)
				}

				// Verify Start is approximately daysBack from now
				expectedStart := tr.End.AddDate(0, 0, -tt.daysBack)
				diff := tr.Start.Sub(expectedStart)
				if diff < -1*time.Second || diff > 1*time.Second {
					t.Errorf("Expected Start to be %d days before End, got difference of %v",
						tt.daysBack, diff)
				}

				// Verify Start is before or equal to End
				if tr.Start.After(tr.End) {
					t.Errorf("Start (%v) should not be after End (%v)", tr.Start, tr.End)
				}
			} else {
				if tr != nil {
					t.Error("Expected nil TimeRange on error, got non-nil")
				}
			}
		})
	}
}

func TestNewTimeRangeFromDaysBack_ErrorMessage(t *testing.T) {
	_, err := NewTimeRangeFromDaysBack(-5)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expectedMsg := "daysBack must be non-negative"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message = %q, got = %q", expectedMsg, err.Error())
	}
}

func TestNewTimeRangeFromDaysBack_ZeroDays(t *testing.T) {
	tr, err := NewTimeRangeFromDaysBack(0)
	if err != nil {
		t.Fatalf("Expected no error for 0 days, got %v", err)
	}

	// For 0 days, Start and End should be approximately the same (within a second)
	diff := tr.End.Sub(tr.Start)
	if diff < 0 || diff > 1*time.Second {
		t.Errorf("For 0 days, expected Start and End to be very close, got difference of %v", diff)
	}
}
