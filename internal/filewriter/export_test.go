package filewriter

// This file exports private functions for testing purposes only.
// It follows the pattern used in Go's standard library (e.g., math/export_test.go).

// Exported for testing
var (
	WriteMarkdownFile   = writeMarkdownFile
	SanitizeChannelName = sanitizeChannelName
	IsSameDay           = isSameDay
	GetSortedDates      = getSortedDates
	SortDates           = sortDates
)
