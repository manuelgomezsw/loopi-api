package base

import (
	"fmt"
	"log"
	"time"
)

// LogLevel represents the level of logging
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// Logger provides structured logging for use cases
type Logger struct {
	entityName string
	level      LogLevel
}

// NewLogger creates a new logger for a specific entity
func NewLogger(entityName string) *Logger {
	return &Logger{
		entityName: entityName,
		level:      LogLevelInfo, // Default level
	}
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// LogOperation logs a use case operation
func (l *Logger) LogOperation(operation string, status string, metadata map[string]interface{}) {
	if l.level > LogLevelInfo {
		return
	}

	logEntry := l.createLogEntry("OPERATION", operation, status, metadata)
	log.Println(logEntry)
}

// LogError logs an error in a use case operation
func (l *Logger) LogError(operation string, err error, metadata map[string]interface{}) {
	if l.level > LogLevelError {
		return
	}

	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["error"] = err.Error()

	logEntry := l.createLogEntry("ERROR", operation, "failed", metadata)
	log.Println(logEntry)
}

// LogWarning logs a warning in a use case operation
func (l *Logger) LogWarning(operation string, message string, metadata map[string]interface{}) {
	if l.level > LogLevelWarn {
		return
	}

	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["warning"] = message

	logEntry := l.createLogEntry("WARNING", operation, "warning", metadata)
	log.Println(logEntry)
}

// LogDebug logs debug information
func (l *Logger) LogDebug(operation string, message string, metadata map[string]interface{}) {
	if l.level > LogLevelDebug {
		return
	}

	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["debug"] = message

	logEntry := l.createLogEntry("DEBUG", operation, "debug", metadata)
	log.Println(logEntry)
}

// LogBusinessRule logs business rule validation/execution
func (l *Logger) LogBusinessRule(operation string, rule string, status string, metadata map[string]interface{}) {
	if l.level > LogLevelInfo {
		return
	}

	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["business_rule"] = rule

	logEntry := l.createLogEntry("BUSINESS_RULE", operation, status, metadata)
	log.Println(logEntry)
}

// LogValidation logs validation results
func (l *Logger) LogValidation(operation string, field string, status string, metadata map[string]interface{}) {
	if l.level > LogLevelInfo {
		return
	}

	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["validation_field"] = field

	logEntry := l.createLogEntry("VALIDATION", operation, status, metadata)
	log.Println(logEntry)
}

// LogPerformance logs performance metrics
func (l *Logger) LogPerformance(operation string, duration time.Duration, metadata map[string]interface{}) {
	if l.level > LogLevelInfo {
		return
	}

	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["duration_ms"] = duration.Milliseconds()
	metadata["duration_readable"] = duration.String()

	logEntry := l.createLogEntry("PERFORMANCE", operation, "completed", metadata)
	log.Println(logEntry)
}

// createLogEntry creates a structured log entry
func (l *Logger) createLogEntry(logType string, operation string, status string, metadata map[string]interface{}) string {
	timestamp := time.Now().Format("2006-01-02T15:04:05.000Z")

	entry := fmt.Sprintf("[%s] %s UseCase=%s Operation=%s Status=%s",
		timestamp, logType, l.entityName, operation, status)

	// Add metadata
	if metadata != nil && len(metadata) > 0 {
		metadataStr := " Metadata={"
		first := true
		for key, value := range metadata {
			if !first {
				metadataStr += ", "
			}
			metadataStr += fmt.Sprintf("%s=%v", key, value)
			first = false
		}
		metadataStr += "}"
		entry += metadataStr
	}

	return entry
}

// WithMetadata creates a logger with pre-set metadata
func (l *Logger) WithMetadata(metadata map[string]interface{}) *LoggerWithMetadata {
	return &LoggerWithMetadata{
		logger:   l,
		metadata: metadata,
	}
}

// LoggerWithMetadata is a logger with pre-set metadata
type LoggerWithMetadata struct {
	logger   *Logger
	metadata map[string]interface{}
}

// LogOperation logs an operation with pre-set metadata
func (lm *LoggerWithMetadata) LogOperation(operation string, status string, additionalMetadata map[string]interface{}) {
	metadata := lm.mergeMetadata(additionalMetadata)
	lm.logger.LogOperation(operation, status, metadata)
}

// LogError logs an error with pre-set metadata
func (lm *LoggerWithMetadata) LogError(operation string, err error, additionalMetadata map[string]interface{}) {
	metadata := lm.mergeMetadata(additionalMetadata)
	lm.logger.LogError(operation, err, metadata)
}

// mergeMetadata merges pre-set metadata with additional metadata
func (lm *LoggerWithMetadata) mergeMetadata(additional map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	// Copy pre-set metadata
	for k, v := range lm.metadata {
		merged[k] = v
	}

	// Add additional metadata (overwrites pre-set if same key)
	if additional != nil {
		for k, v := range additional {
			merged[k] = v
		}
	}

	return merged
}

// Performance timer utility
type PerformanceTimer struct {
	logger    *Logger
	operation string
	startTime time.Time
	metadata  map[string]interface{}
}

// StartTimer starts a performance timer
func (l *Logger) StartTimer(operation string, metadata map[string]interface{}) *PerformanceTimer {
	return &PerformanceTimer{
		logger:    l,
		operation: operation,
		startTime: time.Now(),
		metadata:  metadata,
	}
}

// Stop stops the timer and logs the duration
func (pt *PerformanceTimer) Stop() {
	duration := time.Since(pt.startTime)
	pt.logger.LogPerformance(pt.operation, duration, pt.metadata)
}
