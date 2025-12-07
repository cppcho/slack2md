package valueobjects

import "errors"

// ExportConfig represents the configuration for an export operation
type ExportConfig struct {
	ExportPath string
	TimeRange  TimeRange
}

// NewExportConfig creates a new ExportConfig with validation
func NewExportConfig(exportPath string, timeRange TimeRange) (*ExportConfig, error) {
	if exportPath == "" {
		return nil, errors.New("export path cannot be empty")
	}
	return &ExportConfig{
		ExportPath: exportPath,
		TimeRange:  timeRange,
	}, nil
}
