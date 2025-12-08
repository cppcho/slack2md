package logger

import "fmt"

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// Logger implements the interfaces.Logger interface
type Logger struct {
	level LogLevel
}

// NewLogger creates a new Logger with the specified minimum log level
func NewLogger(level LogLevel) *Logger {
	return &Logger{level: level}
}

// Info logs an informational message
func (l *Logger) Info(msg string, args ...interface{}) {
	if l.level <= INFO {
		fmt.Printf("[INFO] "+msg+"\n", args...)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...interface{}) {
	if l.level <= WARN {
		fmt.Printf("[WARN] "+msg+"\n", args...)
	}
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...interface{}) {
	if l.level <= ERROR {
		fmt.Printf("[ERROR] "+msg+"\n", args...)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...interface{}) {
	if l.level <= DEBUG {
		fmt.Printf("[DEBUG] "+msg+"\n", args...)
	}
}
