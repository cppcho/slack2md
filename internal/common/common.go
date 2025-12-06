package common

import "fmt"

// Success prints a success message
func Success(message string) {
	fmt.Printf("✓ %s\n", message)
}

// Error prints an error message
func Error(message string) {
	fmt.Printf("✗ %s\n", message)
}
