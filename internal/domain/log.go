package domain

import (
	"time"

	"github.com/google/uuid"
)

type LogLevel string

const (
	LogLevelDebug   LogLevel = "debug"
	LogLevelInfo    LogLevel = "info"
	LogLevelWarning LogLevel = "warning"
	LogLevelError   LogLevel = "error"
)

type Log struct {
	ID        uuid.UUID      `json:"id"`
	UserID    uuid.UUID      `json:"user_id"`
	Level     LogLevel       `json:"level"`
	Message   string         `json:"message"`
	Metadata  map[string]any `json:"metadata"`
	Timestamp time.Time      `json:"timestamp"`
	Source    string         `json:"source"`
	Tags      []string       `json:"tags"`
}
