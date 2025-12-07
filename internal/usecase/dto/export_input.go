package dto

// ExportChannelsInput contains the input parameters for channel export
type ExportChannelsInput struct {
	BotToken   string
	AppToken   string
	ChannelIDs []string // Empty means auto-discover
	ExportPath string
	DaysBack   int
}
