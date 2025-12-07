package config

import (
	"fmt"
	"strings"
)

// ConfigValidator validates configuration
type ConfigValidator struct{}

// NewConfigValidator creates a new ConfigValidator
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{}
}

// Validate validates the configuration and returns detailed errors
func (v *ConfigValidator) Validate(config *EnvConfig) error {
	var errs []string

	// Validate BotToken
	if config.BotToken == "" {
		errs = append(errs, "SLACK_BOT_TOKEN is required")
	} else if !strings.HasPrefix(config.BotToken, "xoxb-") {
		errs = append(errs, "SLACK_BOT_TOKEN must start with 'xoxb-'")
	}

	// Validate AppToken (if provided)
	if config.AppToken != "" && !strings.HasPrefix(config.AppToken, "xapp-") {
		errs = append(errs, "SLACK_APP_TOKEN must start with 'xapp-' when provided")
	}

	// Validate ExportPath
	if config.ExportPath == "" {
		errs = append(errs, "SLACK_EXPORT_PATH is required")
	}

	// Validate DaysBack
	if config.DaysBack < 0 {
		errs = append(errs, "DAYS_BACK must be non-negative")
	}
	if config.DaysBack > 365 {
		errs = append(errs, "DAYS_BACK cannot exceed 365 days")
	}

	if len(errs) > 0 {
		return fmt.Errorf("configuration validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}

	return nil
}
