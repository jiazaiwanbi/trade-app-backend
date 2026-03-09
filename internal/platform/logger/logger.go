package logger

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"
)

type Logger struct {
	out io.Writer
	mu  sync.Mutex
}

type entry struct {
	Timestamp string         `json:"timestamp"`
	Level     string         `json:"level"`
	Message   string         `json:"message"`
	Fields    map[string]any `json:"fields,omitempty"`
}

func New() *Logger {
	return &Logger{out: os.Stdout}
}

func (l *Logger) Info(message string, fields map[string]any) {
	l.log("INFO", message, fields)
}

func (l *Logger) Error(message string, fields map[string]any) {
	l.log("ERROR", message, fields)
}

func (l *Logger) InfoContext(ctx context.Context, message string, fields map[string]any) {
	l.Info(message, withContextFields(ctx, fields))
}

func (l *Logger) ErrorContext(ctx context.Context, message string, fields map[string]any) {
	l.Error(message, withContextFields(ctx, fields))
}

func (l *Logger) log(level string, message string, fields map[string]any) {
	payload := entry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     level,
		Message:   message,
		Fields:    fields,
	}

	encoded, _ := json.Marshal(payload)

	l.mu.Lock()
	defer l.mu.Unlock()

	_, _ = l.out.Write(append(encoded, '\n'))
}

func withContextFields(ctx context.Context, fields map[string]any) map[string]any {
	merged := map[string]any{}
	for key, value := range fields {
		merged[key] = value
	}

	if requestID, ok := RequestIDFromContext(ctx); ok {
		merged["request_id"] = requestID
	}

	return merged
}
