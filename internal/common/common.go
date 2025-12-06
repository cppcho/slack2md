package common

import "fmt"

// PrintBanner prints a formatted banner with the given tool name
func PrintBanner(toolName string) {
	fmt.Printf("=== %s ===\n", toolName)
}

// Success prints a success message
func Success(message string) {
	fmt.Printf("✓ %s\n", message)
}

// Error prints an error message
func Error(message string) {
	fmt.Printf("✗ %s\n", message)
}
