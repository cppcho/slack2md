package main_test

import (
	"testing"

	main "github.com/cppcho/slack2md/cmd/slack2md"
)

// TestLoadConfig_NilPointerSafety ensures LoadConfig returns a valid config pointer
func TestLoadConfig(t *testing.T) {
	config, err := main.LoadConfig()
	if err != nil {
		t.Fatalf("main.LoadConfig() unexpected error: %v", err)
	}
	if config == nil {
		t.Fatal("main.LoadConfig() returned nil config pointer")
	}
}
