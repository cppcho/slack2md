package common_test

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/cppcho/slack2md/internal/common"
)

// captureStdout captures stdout during function execution
func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestSuccess(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		validate func(t *testing.T, output string)
	}{
		{
			name:    "simple success message",
			message: "Operation completed",
			validate: func(t *testing.T, output string) {
				if !strings.Contains(output, "✓") {
					t.Error("expected output to contain checkmark symbol")
				}
				if !strings.Contains(output, "Operation completed") {
					t.Error("expected output to contain message")
				}
			},
		},
		{
			name:    "success with details",
			message: "Exported 100 messages from #general",
			validate: func(t *testing.T, output string) {
				if !strings.Contains(output, "✓") {
					t.Error("expected output to contain checkmark symbol")
				}
				if !strings.Contains(output, "Exported 100 messages from #general") {
					t.Error("expected output to contain message")
				}
			},
		},
		{
			name:    "empty message",
			message: "",
			validate: func(t *testing.T, output string) {
				if !strings.Contains(output, "✓") {
					t.Error("expected output to contain checkmark symbol")
				}
				expected := "✓ \n"
				if output != expected {
					t.Errorf("expected output %q, got %q", expected, output)
				}
			},
		},
		{
			name:    "message with newlines",
			message: "Line 1\nLine 2",
			validate: func(t *testing.T, output string) {
				if !strings.Contains(output, "✓") {
					t.Error("expected output to contain checkmark symbol")
				}
				if !strings.Contains(output, "Line 1\nLine 2") {
					t.Error("expected output to preserve newlines")
				}
			},
		},
		{
			name:    "message with special characters",
			message: "Success: 100% complete!",
			validate: func(t *testing.T, output string) {
				if !strings.Contains(output, "Success: 100% complete!") {
					t.Error("expected output to contain message with special chars")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStdout(func() {
				common.Success(tt.message)
			})

			if tt.validate != nil {
				tt.validate(t, output)
			}
		})
	}
}

func TestError(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		validate func(t *testing.T, output string)
	}{
		{
			name:    "simple error message",
			message: "Connection failed",
			validate: func(t *testing.T, output string) {
				if !strings.Contains(output, "✗") {
					t.Error("expected output to contain X mark symbol")
				}
				if !strings.Contains(output, "Connection failed") {
					t.Error("expected output to contain message")
				}
			},
		},
		{
			name:    "error with details",
			message: "Failed to export: channel not found",
			validate: func(t *testing.T, output string) {
				if !strings.Contains(output, "✗") {
					t.Error("expected output to contain X mark symbol")
				}
				if !strings.Contains(output, "Failed to export: channel not found") {
					t.Error("expected output to contain message")
				}
			},
		},
		{
			name:    "empty error message",
			message: "",
			validate: func(t *testing.T, output string) {
				if !strings.Contains(output, "✗") {
					t.Error("expected output to contain X mark symbol")
				}
				expected := "✗ \n"
				if output != expected {
					t.Errorf("expected output %q, got %q", expected, output)
				}
			},
		},
		{
			name:    "error with technical details",
			message: "Error: status code 404, endpoint /api/channels",
			validate: func(t *testing.T, output string) {
				if !strings.Contains(output, "Error: status code 404") {
					t.Error("expected output to contain error details")
				}
			},
		},
		{
			name:    "error with file path",
			message: "Cannot write to /tmp/export/general.md: permission denied",
			validate: func(t *testing.T, output string) {
				if !strings.Contains(output, "/tmp/export/general.md") {
					t.Error("expected output to contain file path")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStdout(func() {
				common.Error(tt.message)
			})

			if tt.validate != nil {
				tt.validate(t, output)
			}
		})
	}
}

// TestOutputFormatConsistency verifies all output functions follow consistent formatting
func TestOutputFormatConsistency(t *testing.T) {
	tests := []struct {
		name       string
		function   func()
		wantPrefix string
	}{
		{
			name: "Success format",
			function: func() {
				common.Success("test")
			},
			wantPrefix: "✓",
		},
		{
			name: "Error format",
			function: func() {
				common.Error("test")
			},
			wantPrefix: "✗",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStdout(tt.function)

			if !strings.HasPrefix(output, tt.wantPrefix) {
				t.Errorf("expected output to start with %q, got %q", tt.wantPrefix, output)
			}

			if !strings.HasSuffix(output, "\n") {
				t.Error("expected output to end with newline")
			}
		})
	}
}

// TestMultipleOutputCalls verifies functions can be called multiple times
func TestMultipleOutputCalls(t *testing.T) {
	output := captureStdout(func() {
		common.Success("First operation")
		common.Success("Second operation")
		common.Error("Something went wrong")
	})

	expectedParts := []string{
		"✓ First operation",
		"✓ Second operation",
		"✗ Something went wrong",
	}

	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("expected output to contain %q, got %q", part, output)
		}
	}
}
