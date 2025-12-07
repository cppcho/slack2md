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

// LoggerImpl implements the Logger interface
type LoggerImpl struct {
	level LogLevel
}

// NewLogger creates a new Logger with the specified minimum log level
func NewLogger(level LogLevel) *LoggerImpl {
	return &LoggerImpl{level: level}
}

// Info logs an informational message
func (l *LoggerImpl) Info(msg string, args ...interface{}) {
	if l.level <= INFO {
		fmt.Printf("[INFO] "+msg+"\n", args...)
	}
}

// Warn logs a warning message
func (l *LoggerImpl) Warn(msg string, args ...interface{}) {
	if l.level <= WARN {
		fmt.Printf("[WARN] "+msg+"\n", args...)
	}
}

// Error logs an error message
func (l *LoggerImpl) Error(msg string, args ...interface{}) {
	if l.level <= ERROR {
		fmt.Printf("[ERROR] "+msg+"\n", args...)
	}
}

// Debug logs a debug message
func (l *LoggerImpl) Debug(msg string, args ...interface{}) {
	if l.level <= DEBUG {
		fmt.Printf("[DEBUG] "+msg+"\n", args...)
	}
}
