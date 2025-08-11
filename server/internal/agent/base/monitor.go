package base

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// LogLevel represents the logging level
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Monitor provides monitoring and debugging capabilities for the agent
type Monitor struct {
	enabled  bool
	logLevel LogLevel
	logger   *log.Logger
}

// NewMonitor creates a new monitor instance
func NewMonitor(enabled bool, logLevel LogLevel, logger *log.Logger) *Monitor {
	if logger == nil {
		logger = log.Default()
	}

	return &Monitor{
		enabled:  enabled,
		logLevel: logLevel,
		logger:   logger,
	}
}

// Debug logs a debug message
func (m *Monitor) Debug(component string, message string, data map[string]interface{}) {
	m.log(LogLevelDebug, component, message, data)
}

// Info logs an info message
func (m *Monitor) Info(component string, message string, data map[string]interface{}) {
	m.log(LogLevelInfo, component, message, data)
}

// Warn logs a warning message
func (m *Monitor) Warn(component string, message string, data map[string]interface{}) {
	m.log(LogLevelWarn, component, message, data)
}

// Error logs an error message
func (m *Monitor) Error(component string, message string, err error) {
	data := map[string]interface{}{}
	if err != nil {
		data["error"] = err.Error()
	}
	m.log(LogLevelError, component, message, data)
}

// log logs a message with the specified level
func (m *Monitor) log(level LogLevel, component string, message string, data map[string]interface{}) {
	if !m.enabled || level < m.logLevel {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logMessage := fmt.Sprintf("[%s] [%s] [%s] %s", timestamp, level.String(), component, message)

	if len(data) > 0 {
		strData, _ := json.Marshal(data)
		logMessage += fmt.Sprintf(" - Data: %s", strData)
	}

	m.logger.Println(logMessage)
}

// SetLogLevel sets the logging level
func (m *Monitor) SetLogLevel(level LogLevel) {
	m.logLevel = level
}

// GetLogLevel returns the current logging level
func (m *Monitor) GetLogLevel() LogLevel {
	return m.logLevel
}
