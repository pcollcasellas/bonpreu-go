package utils

import (
	"fmt"
	"log"
	"time"
)

// Logger provides logging functionality
type Logger struct {
	prefix string
}

// NewLogger creates a new logger instance
func NewLogger(prefix string) *Logger {
	return &Logger{prefix: prefix}
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	log.Printf("[INFO] %s: %s", l.prefix, message)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	log.Printf("[ERROR] %s: %s", l.prefix, message)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	log.Printf("[DEBUG] %s: %s", l.prefix, message)
}

// LogDuration logs the duration of an operation
func (l *Logger) LogDuration(operation string, start time.Time) {
	duration := time.Since(start)
	l.Info("%s completed in %v", operation, duration)
}
