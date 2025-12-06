package main_test

import (
	"strings"
	"testing"

	main "github.com/cppcho/slack2md/cmd/slack2md"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T)
		wantErr  bool
		errMsg   string
		validate func(t *testing.T, config *main.Config)
	}{
		{
			name: "valid config with all required fields",
			setup: func(t *testing.T) {
				t.Setenv("SLACK_BOT_TOKEN", "xoxb-test-token")
				t.Setenv("SLACK_EXPORT_PATH", "/tmp/slack-export")
			},
			wantErr: false,
			validate: func(t *testing.T, config *main.Config) {
				if config.BotToken != "xoxb-test-token" {
					t.Errorf("BotToken = %q, want %q", config.BotToken, "xoxb-test-token")
				}
				if config.ExportPath != "/tmp/slack-export" {
					t.Errorf("ExportPath = %q, want %q", config.ExportPath, "/tmp/slack-export")
				}
				if config.DaysBack != 7 {
					t.Errorf("DaysBack = %d, want 7", config.DaysBack)
				}
				if len(config.ChannelIDs) != 0 {
					t.Errorf("ChannelIDs should be empty when not provided, got %v", config.ChannelIDs)
				}
			},
		},
		{
			name: "valid config with optional app token",
			setup: func(t *testing.T) {
				t.Setenv("SLACK_BOT_TOKEN", "xoxb-test-token")
				t.Setenv("SLACK_APP_TOKEN", "xapp-test-token")
				t.Setenv("SLACK_EXPORT_PATH", "/tmp/slack-export")
			},
			wantErr: false,
			validate: func(t *testing.T, config *main.Config) {
				if config.AppToken != "xapp-test-token" {
					t.Errorf("AppToken = %q, want %q", config.AppToken, "xapp-test-token")
				}
			},
		},
		{
			name: "missing SLACK_BOT_TOKEN",
			setup: func(t *testing.T) {
				t.Setenv("SLACK_EXPORT_PATH", "/tmp/slack-export")
			},
			wantErr: true,
			errMsg:  "SLACK_BOT_TOKEN environment variable is required",
		},
		{
			name: "missing SLACK_EXPORT_PATH",
			setup: func(t *testing.T) {
				t.Setenv("SLACK_BOT_TOKEN", "xoxb-test-token")
			},
			wantErr: true,
			errMsg:  "SLACK_EXPORT_PATH environment variable is required",
		},
		{
			name: "empty SLACK_BOT_TOKEN",
			setup: func(t *testing.T) {
				t.Setenv("SLACK_BOT_TOKEN", "")
				t.Setenv("SLACK_EXPORT_PATH", "/tmp/slack-export")
			},
			wantErr: true,
			errMsg:  "SLACK_BOT_TOKEN environment variable is required",
		},
		{
			name: "empty SLACK_EXPORT_PATH",
			setup: func(t *testing.T) {
				t.Setenv("SLACK_BOT_TOKEN", "xoxb-test-token")
				t.Setenv("SLACK_EXPORT_PATH", "")
			},
			wantErr: true,
			errMsg:  "SLACK_EXPORT_PATH environment variable is required",
		},
		{
			name: "single channel ID",
			setup: func(t *testing.T) {
				t.Setenv("SLACK_BOT_TOKEN", "xoxb-test-token")
				t.Setenv("SLACK_EXPORT_PATH", "/tmp/slack-export")
				t.Setenv("SLACK_CHANNEL_IDS", "C1234567890")
			},
			wantErr: false,
			validate: func(t *testing.T, config *main.Config) {
				if len(config.ChannelIDs) != 1 {
					t.Errorf("expected 1 channel ID, got %d", len(config.ChannelIDs))
					return
				}
				if config.ChannelIDs[0] != "C1234567890" {
					t.Errorf("ChannelIDs[0] = %q, want %q", config.ChannelIDs[0], "C1234567890")
				}
			},
		},
		{
			name: "multiple comma-separated channel IDs",
			setup: func(t *testing.T) {
				t.Setenv("SLACK_BOT_TOKEN", "xoxb-test-token")
				t.Setenv("SLACK_EXPORT_PATH", "/tmp/slack-export")
				t.Setenv("SLACK_CHANNEL_IDS", "C1234567890,C0987654321,C1111111111")
			},
			wantErr: false,
			validate: func(t *testing.T, config *main.Config) {
				expected := []string{"C1234567890", "C0987654321", "C1111111111"}
				if len(config.ChannelIDs) != len(expected) {
					t.Errorf("expected %d channel IDs, got %d", len(expected), len(config.ChannelIDs))
					return
				}
				for i, exp := range expected {
					if config.ChannelIDs[i] != exp {
						t.Errorf("ChannelIDs[%d] = %q, want %q", i, config.ChannelIDs[i], exp)
					}
				}
			},
		},
		{
			name: "channel IDs with whitespace",
			setup: func(t *testing.T) {
				t.Setenv("SLACK_BOT_TOKEN", "xoxb-test-token")
				t.Setenv("SLACK_EXPORT_PATH", "/tmp/slack-export")
				t.Setenv("SLACK_CHANNEL_IDS", " C1234567890 , C0987654321 , C1111111111 ")
			},
			wantErr: false,
			validate: func(t *testing.T, config *main.Config) {
				expected := []string{"C1234567890", "C0987654321", "C1111111111"}
				if len(config.ChannelIDs) != len(expected) {
					t.Errorf("expected %d channel IDs, got %d", len(expected), len(config.ChannelIDs))
					return
				}
				for i, exp := range expected {
					if config.ChannelIDs[i] != exp {
						t.Errorf("ChannelIDs[%d] = %q, want %q (whitespace not trimmed)", i, config.ChannelIDs[i], exp)
					}
				}
			},
		},
		{
			name: "channel IDs with trailing comma",
			setup: func(t *testing.T) {
				t.Setenv("SLACK_BOT_TOKEN", "xoxb-test-token")
				t.Setenv("SLACK_EXPORT_PATH", "/tmp/slack-export")
				t.Setenv("SLACK_CHANNEL_IDS", "C1234567890,C0987654321,")
			},
			wantErr: false,
			validate: func(t *testing.T, config *main.Config) {
				expected := []string{"C1234567890", "C0987654321"}
				if len(config.ChannelIDs) != len(expected) {
					t.Errorf("expected %d channel IDs, got %d", len(expected), len(config.ChannelIDs))
					return
				}
				for i, exp := range expected {
					if config.ChannelIDs[i] != exp {
						t.Errorf("ChannelIDs[%d] = %q, want %q", i, config.ChannelIDs[i], exp)
					}
				}
			},
		},
		{
			name: "channel IDs with leading comma",
			setup: func(t *testing.T) {
				t.Setenv("SLACK_BOT_TOKEN", "xoxb-test-token")
				t.Setenv("SLACK_EXPORT_PATH", "/tmp/slack-export")
				t.Setenv("SLACK_CHANNEL_IDS", ",C1234567890,C0987654321")
			},
			wantErr: false,
			validate: func(t *testing.T, config *main.Config) {
				expected := []string{"C1234567890", "C0987654321"}
				if len(config.ChannelIDs) != len(expected) {
					t.Errorf("expected %d channel IDs, got %d", len(expected), len(config.ChannelIDs))
					return
				}
				for i, exp := range expected {
					if config.ChannelIDs[i] != exp {
						t.Errorf("ChannelIDs[%d] = %q, want %q", i, config.ChannelIDs[i], exp)
					}
				}
			},
		},
		{
			name: "channel IDs with multiple consecutive commas",
			setup: func(t *testing.T) {
				t.Setenv("SLACK_BOT_TOKEN", "xoxb-test-token")
				t.Setenv("SLACK_EXPORT_PATH", "/tmp/slack-export")
				t.Setenv("SLACK_CHANNEL_IDS", "C1234567890,,,C0987654321")
			},
			wantErr: false,
			validate: func(t *testing.T, config *main.Config) {
				expected := []string{"C1234567890", "C0987654321"}
				if len(config.ChannelIDs) != len(expected) {
					t.Errorf("expected %d channel IDs, got %d", len(expected), len(config.ChannelIDs))
					return
				}
				for i, exp := range expected {
					if config.ChannelIDs[i] != exp {
						t.Errorf("ChannelIDs[%d] = %q, want %q", i, config.ChannelIDs[i], exp)
					}
				}
			},
		},
		{
			name: "empty channel IDs string (valid for auto-discovery)",
			setup: func(t *testing.T) {
				t.Setenv("SLACK_BOT_TOKEN", "xoxb-test-token")
				t.Setenv("SLACK_EXPORT_PATH", "/tmp/slack-export")
				t.Setenv("SLACK_CHANNEL_IDS", "")
			},
			wantErr: false,
			validate: func(t *testing.T, config *main.Config) {
				if len(config.ChannelIDs) != 0 {
					t.Errorf("expected empty ChannelIDs for auto-discovery, got %v", config.ChannelIDs)
				}
			},
		},
		{
			name: "export path with spaces",
			setup: func(t *testing.T) {
				t.Setenv("SLACK_BOT_TOKEN", "xoxb-test-token")
				t.Setenv("SLACK_EXPORT_PATH", "/tmp/slack export/my files")
			},
			wantErr: false,
			validate: func(t *testing.T, config *main.Config) {
				if config.ExportPath != "/tmp/slack export/my files" {
					t.Errorf("ExportPath = %q, want %q", config.ExportPath, "/tmp/slack export/my files")
				}
			},
		},
		{
			name: "export path with special characters",
			setup: func(t *testing.T) {
				t.Setenv("SLACK_BOT_TOKEN", "xoxb-test-token")
				t.Setenv("SLACK_EXPORT_PATH", "/tmp/slack-export_2024@home")
			},
			wantErr: false,
			validate: func(t *testing.T, config *main.Config) {
				if config.ExportPath != "/tmp/slack-export_2024@home" {
					t.Errorf("ExportPath = %q, want %q", config.ExportPath, "/tmp/slack-export_2024@home")
				}
			},
		},
		{
			name: "bot token with different format",
			setup: func(t *testing.T) {
				t.Setenv("SLACK_BOT_TOKEN", "xoxb-1234567890-test-token")
				t.Setenv("SLACK_EXPORT_PATH", "/tmp/slack-export")
			},
			wantErr: false,
			validate: func(t *testing.T, config *main.Config) {
				expected := "xoxb-1234567890-test-token"
				if config.BotToken != expected {
					t.Errorf("BotToken = %q, want %q", config.BotToken, expected)
				}
			},
		},
		{
			name: "app token with xapp prefix",
			setup: func(t *testing.T) {
				t.Setenv("SLACK_BOT_TOKEN", "xoxb-test-token")
				t.Setenv("SLACK_APP_TOKEN", "xapp-1-A1234567890-test-token")
				t.Setenv("SLACK_EXPORT_PATH", "/tmp/slack-export")
			},
			wantErr: false,
			validate: func(t *testing.T, config *main.Config) {
				expected := "xapp-1-A1234567890-test-token"
				if config.AppToken != expected {
					t.Errorf("AppToken = %q, want %q", config.AppToken, expected)
				}
			},
		},
		{
			name: "defaults - DaysBack is set to 7",
			setup: func(t *testing.T) {
				t.Setenv("SLACK_BOT_TOKEN", "xoxb-test-token")
				t.Setenv("SLACK_EXPORT_PATH", "/tmp/slack-export")
			},
			wantErr: false,
			validate: func(t *testing.T, config *main.Config) {
				if config.DaysBack != 7 {
					t.Errorf("DaysBack = %d, want 7 (default)", config.DaysBack)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			tt.setup(t)

			// Call LoadConfig
			config, err := main.LoadConfig()

			// Check error expectations
			if tt.wantErr {
				if err == nil {
					t.Errorf("main.LoadConfig() expected error, got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("main.LoadConfig() error = %q, want error containing %q", err.Error(), tt.errMsg)
				}
				return
			}

			// No error expected
			if err != nil {
				t.Errorf("main.LoadConfig() unexpected error: %v", err)
				return
			}

			// Validate config
			if tt.validate != nil {
				tt.validate(t, config)
			}
		})
	}
}

// TestLoadConfig_NilPointerSafety ensures LoadConfig returns a valid config pointer
func TestLoadConfig_NilPointerSafety(t *testing.T) {
	t.Setenv("SLACK_BOT_TOKEN", "xoxb-test-token")
	t.Setenv("SLACK_EXPORT_PATH", "/tmp/slack-export")

	config, err := main.LoadConfig()
	if err != nil {
		t.Fatalf("main.LoadConfig() unexpected error: %v", err)
	}

	if config == nil {
		t.Fatal("main.LoadConfig() returned nil config pointer")
	}
}
