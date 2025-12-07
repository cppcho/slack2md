package dto

import "github.com/cppcho/slack2md/internal/domain/entities"

// ExportChannelsOutput contains the output result of channel export
type ExportChannelsOutput struct {
	Summary entities.ExportSummary
}
