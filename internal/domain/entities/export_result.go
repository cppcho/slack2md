package entities

// ExportResult contains the result of exporting a single channel
type ExportResult struct {
	Channel      Channel
	MessageCount int
	DateCount    int
	Success      bool
	Error        error
}

// ExportSummary contains the summary of all export operations
type ExportSummary struct {
	TotalChannels int
	SuccessCount  int
	FailureCount  int
	Results       []ExportResult
}
