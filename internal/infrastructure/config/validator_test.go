package config

import (
	"strings"
	"testing"
)

func TestConfigValidator_Validate_Success(t *testing.T) {
	validator := NewConfigValidator()
	config := &EnvConfig{
		BotToken:   "xoxb-test-token",
		AppToken:   "xapp-test-token",
		ExportPath: "/tmp/export",
		DaysBack:   7,
	}

	err := validator.Validate(config)
	if err != nil {
		t.Errorf("Expected validation to pass, got error: %v", err)
	}
}

func TestConfigValidator_Validate_MissingBotToken(t *testing.T) {
	validator := NewConfigValidator()
	config := &EnvConfig{
		ExportPath: "/tmp/export",
		DaysBack:   7,
	}

	err := validator.Validate(config)
	if err == nil {
		t.Error("Expected validation to fail for missing bot token")
	}
	if !strings.Contains(err.Error(), "SLACK_BOT_TOKEN is required") {
		t.Errorf("Expected error about missing bot token, got: %v", err)
	}
}

func TestConfigValidator_Validate_InvalidBotTokenPrefix(t *testing.T) {
	validator := NewConfigValidator()
	config := &EnvConfig{
		BotToken:   "invalid-prefix-token",
		ExportPath: "/tmp/export",
		DaysBack:   7,
	}

	err := validator.Validate(config)
	if err == nil {
		t.Error("Expected validation to fail for invalid bot token prefix")
	}
	if !strings.Contains(err.Error(), "must start with 'xoxb-'") {
		t.Errorf("Expected error about bot token prefix, got: %v", err)
	}
}

func TestConfigValidator_Validate_InvalidAppTokenPrefix(t *testing.T) {
	validator := NewConfigValidator()
	config := &EnvConfig{
		BotToken:   "xoxb-test-token",
		AppToken:   "invalid-prefix",
		ExportPath: "/tmp/export",
		DaysBack:   7,
	}

	err := validator.Validate(config)
	if err == nil {
		t.Error("Expected validation to fail for invalid app token prefix")
	}
	if !strings.Contains(err.Error(), "must start with 'xapp-'") {
		t.Errorf("Expected error about app token prefix, got: %v", err)
	}
}

func TestConfigValidator_Validate_MissingExportPath(t *testing.T) {
	validator := NewConfigValidator()
	config := &EnvConfig{
		BotToken: "xoxb-test-token",
		DaysBack: 7,
	}

	err := validator.Validate(config)
	if err == nil {
		t.Error("Expected validation to fail for missing export path")
	}
	if !strings.Contains(err.Error(), "SLACK_EXPORT_PATH is required") {
		t.Errorf("Expected error about missing export path, got: %v", err)
	}
}

func TestConfigValidator_Validate_NegativeDaysBack(t *testing.T) {
	validator := NewConfigValidator()
	config := &EnvConfig{
		BotToken:   "xoxb-test-token",
		ExportPath: "/tmp/export",
		DaysBack:   -1,
	}

	err := validator.Validate(config)
	if err == nil {
		t.Error("Expected validation to fail for negative days back")
	}
	if !strings.Contains(err.Error(), "must be non-negative") {
		t.Errorf("Expected error about negative days back, got: %v", err)
	}
}

func TestConfigValidator_Validate_TooManyDaysBack(t *testing.T) {
	validator := NewConfigValidator()
	config := &EnvConfig{
		BotToken:   "xoxb-test-token",
		ExportPath: "/tmp/export",
		DaysBack:   400,
	}

	err := validator.Validate(config)
	if err == nil {
		t.Error("Expected validation to fail for days back > 365")
	}
	if !strings.Contains(err.Error(), "cannot exceed 365 days") {
		t.Errorf("Expected error about too many days back, got: %v", err)
	}
}

func TestConfigValidator_Validate_MultipleErrors(t *testing.T) {
	validator := NewConfigValidator()
	config := &EnvConfig{
		DaysBack: -1,
	}

	err := validator.Validate(config)
	if err == nil {
		t.Error("Expected validation to fail with multiple errors")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "SLACK_BOT_TOKEN is required") {
		t.Error("Expected error about missing bot token")
	}
	if !strings.Contains(errMsg, "SLACK_EXPORT_PATH is required") {
		t.Error("Expected error about missing export path")
	}
	if !strings.Contains(errMsg, "must be non-negative") {
		t.Error("Expected error about negative days back")
	}
}
