package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// LogLevel represents the logging level
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
	SILENT
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	case SILENT:
		return "SILENT"
	default:
		return "UNKNOWN"
	}
}

// Logger provides structured logging to stderr
type Logger struct {
	level LogLevel
}

// Global logger instance
var globalLogger *Logger

// init initializes the global logger based on environment variables
func init() {
	level := getLogLevelFromEnv()
	globalLogger = &Logger{level: level}
}

// getLogLevelFromEnv determines log level from environment variables
func getLogLevelFromEnv() LogLevel {
	// Check for specific log level setting
	if levelStr := os.Getenv("ATLASSIAN_ASSETS_LOG_LEVEL"); levelStr != "" {
		switch strings.ToUpper(levelStr) {
		case "DEBUG":
			return DEBUG
		case "INFO":
			return INFO
		case "WARNING", "WARN":
			return WARNING
		case "ERROR":
			return ERROR
		case "SILENT", "NONE":
			return SILENT
		}
	}

	// Check for legacy debug flag
	if debugStr := os.Getenv("ATLASSIAN_ASSETS_DEBUG"); debugStr != "" {
		if strings.ToLower(debugStr) == "true" || debugStr == "1" {
			return DEBUG
		}
	}

	// Default to WARNING level for production use
	return WARNING
}

// SetLevel sets the global logging level
func SetLevel(level LogLevel) {
	globalLogger.level = level
}

// GetLevel returns the current logging level
func GetLevel() LogLevel {
	return globalLogger.level
}

// shouldLog checks if a message should be logged based on current level
func (l *Logger) shouldLog(level LogLevel) bool {
	return l.level <= level
}

// logf formats and logs a message to stderr if the level is appropriate
func (l *Logger) logf(level LogLevel, format string, args ...interface{}) {
	if !l.shouldLog(level) {
		return
	}

	prefix := fmt.Sprintf("[%s] ", level.String())
	message := fmt.Sprintf(format, args...)
	
	// Always log to stderr to avoid contaminating stdout
	fmt.Fprintf(os.Stderr, "%s%s\n", prefix, message)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.logf(DEBUG, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.logf(INFO, format, args...)
}

// Warning logs a warning message
func (l *Logger) Warning(format string, args ...interface{}) {
	l.logf(WARNING, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.logf(ERROR, format, args...)
}

// Fatal logs an error message and exits the program
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.logf(ERROR, format, args...)
	os.Exit(1)
}

// Global convenience functions
func Debug(format string, args ...interface{}) {
	globalLogger.Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	globalLogger.Info(format, args...)
}

func Warning(format string, args ...interface{}) {
	globalLogger.Warning(format, args...)
}

func Error(format string, args ...interface{}) {
	globalLogger.Error(format, args...)
}

func Fatal(format string, args ...interface{}) {
	globalLogger.Fatal(format, args...)
}

// IsDebugEnabled returns true if debug logging is enabled
func IsDebugEnabled() bool {
	return globalLogger.level <= DEBUG
}

// IsInfoEnabled returns true if info logging is enabled
func IsInfoEnabled() bool {
	return globalLogger.level <= INFO
}

// SetupStandardLogger configures the standard Go logger to use stderr
func SetupStandardLogger() {
	// Redirect standard log package to stderr with our prefix
	log.SetOutput(os.Stderr)
	log.SetPrefix("[ERROR] ")
	log.SetFlags(log.LstdFlags)
}