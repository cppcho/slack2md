package valueobjects

import (
	"errors"
	"time"
)

// TimeRange represents a time range with start and end times
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// NewTimeRange creates a new TimeRange with validation
func NewTimeRange(start, end time.Time) (*TimeRange, error) {
	if end.Before(start) {
		return nil, errors.New("end time must be after start time")
	}
	return &TimeRange{
		Start: start,
		End:   end,
	}, nil
}

// NewTimeRangeFromDaysBack creates a TimeRange from current time going back N days
func NewTimeRangeFromDaysBack(daysBack int) (*TimeRange, error) {
	if daysBack < 0 {
		return nil, errors.New("daysBack must be non-negative")
	}
	end := time.Now()
	start := end.AddDate(0, 0, -daysBack)
	return &TimeRange{
		Start: start,
		End:   end,
	}, nil
}
