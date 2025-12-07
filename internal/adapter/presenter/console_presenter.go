package presenter

import (
	"fmt"

	"github.com/cppcho/slack2md/internal/domain/entities"
)

// ConsolePresenter handles console output presentation
type ConsolePresenter struct{}

// NewConsolePresenter creates a new ConsolePresenter
func NewConsolePresenter() *ConsolePresenter {
	return &ConsolePresenter{}
}

// Success prints a success message
func (p *ConsolePresenter) Success(message string) {
	fmt.Printf("✓ %s\n", message)
}

// Error prints an error message
func (p *ConsolePresenter) Error(message string) {
	fmt.Printf("✗ %s\n", message)
}

// PrintChannelList prints a list of channels
func (p *ConsolePresenter) PrintChannelList(channels []entities.Channel) {
	fmt.Printf("Found %d channel(s):\n", len(channels))
	for _, ch := range channels {
		fmt.Printf("  - %s (%s)\n", ch.Name, ch.ID)
	}
	fmt.Println()
}

// PrintSummary prints the export summary
func (p *ConsolePresenter) PrintSummary(summary entities.ExportSummary) {
	fmt.Printf("\n--- Summary ---\n")
	fmt.Printf("Total channels: %d\n", summary.TotalChannels)
	fmt.Printf("Successful exports: %d\n", summary.SuccessCount)
	fmt.Printf("Failed exports: %d\n", summary.FailureCount)
}

// PrintProcessingChannel prints a message when processing a channel
func (p *ConsolePresenter) PrintProcessingChannel(channelName string) {
	fmt.Printf("Processing channel %s...\n", channelName)
}

// PrintExportInfo prints export configuration information
func (p *ConsolePresenter) PrintExportInfo(daysBack int, startDate, endDate, exportPath string) {
	fmt.Printf("Exporting messages from last %d days (%s to %s)\n", daysBack, startDate, endDate)
	fmt.Printf("Export path: %s\n\n", exportPath)
}
