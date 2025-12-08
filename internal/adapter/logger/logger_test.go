package logger

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func captureOutput(f func()) string {
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

func TestLoggerImpl_Info(t *testing.T) {
	logger := NewLogger(INFO)
	output := captureOutput(func() {
		logger.Info("test message %s", "arg")
	})

	if !strings.Contains(output, "[INFO] test message arg") {
		t.Errorf("Expected INFO log, got: %s", output)
	}
}

func TestLoggerImpl_InfoWithHigherLevel(t *testing.T) {
	logger := NewLogger(ERROR)
	output := captureOutput(func() {
		logger.Info("test message")
	})

	if output != "" {
		t.Errorf("Expected no output when log level is ERROR, got: %s", output)
	}
}

func TestLoggerImpl_Error(t *testing.T) {
	logger := NewLogger(ERROR)
	output := captureOutput(func() {
		logger.Error("error message %d", 123)
	})

	if !strings.Contains(output, "[ERROR] error message 123") {
		t.Errorf("Expected ERROR log, got: %s", output)
	}
}

func TestLoggerImpl_Debug(t *testing.T) {
	logger := NewLogger(DEBUG)
	output := captureOutput(func() {
		logger.Debug("debug message")
	})

	if !strings.Contains(output, "[DEBUG] debug message") {
		t.Errorf("Expected DEBUG log, got: %s", output)
	}
}

func TestLoggerImpl_Warn(t *testing.T) {
	logger := NewLogger(WARN)
	output := captureOutput(func() {
		logger.Warn("warning message")
	})

	if !strings.Contains(output, "[WARN] warning message") {
		t.Errorf("Expected WARN log, got: %s", output)
	}
}
