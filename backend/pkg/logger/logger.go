package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
)

type LogLevel string

const (
	InfoLevel  LogLevel = "INFO"
	ErrorLevel LogLevel = "ERROR"
	DebugLevel LogLevel = "DEBUG"
)

type LogEntry struct {
	Timestamp     string   `json:"timestamp"`
	Level         LogLevel `json:"level"`
	Message       string   `json:"message"`
	CorrelationID string   `json:"correlation_id,omitempty"`
	UserID        string   `json:"user_id,omitempty"`
	Details       any      `json:"details,omitempty"`
}

func Log(level LogLevel, message string, correlationID string, userID string, details any) {
	entry := LogEntry{
		Timestamp:     time.Now().Format(time.RFC3339),
		Level:         level,
		Message:       message,
		CorrelationID: correlationID,
		UserID:        userID,
		Details:       details,
	}

	data, _ := json.Marshal(entry)
	fmt.Fprintln(os.Stdout, string(data))
}

func Info(message string, correlationID string, details any) {
	Log(InfoLevel, message, correlationID, "", details)
}

func Error(message string, correlationID string, details any) {
	Log(ErrorLevel, message, correlationID, "", details)
}

func NewCorrelationID() string {
	return uuid.New().String()
}
