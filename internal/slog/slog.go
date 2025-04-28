package slog

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"
)

var Logger *slog.Logger

type ColoredHandler struct{}

func (h *ColoredHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h *ColoredHandler) Handle(ctx context.Context, r slog.Record) error {
	color := ""
	reset := "\033[0m"

	switch r.Level {
	case slog.LevelDebug:
		color = "\033[36m" // Cyan
	case slog.LevelInfo:
		color = "\033[32m" // Green
	case slog.LevelWarn:
		color = "\033[33m" // Yellow
	case slog.LevelError:
		color = "\033[31m" // Red
	default:
		color = ""
	}

	timestamp := r.Time.Format(time.RFC3339)
	msg := r.Message
	output := fmt.Sprintf("%s%s [%s] %s%s\n", color, timestamp, r.Level, msg, reset)
	_, err := os.Stdout.Write([]byte(output))
	return err
}

func (h *ColoredHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *ColoredHandler) WithGroup(name string) slog.Handler {
	return h
}

func Init() {
	Logger = slog.New(&ColoredHandler{})
}

func Info(msg string, args ...interface{}) {
	Logger.Info(fmt.Sprintf(msg, args...))
}

func Warn(msg string, args ...interface{}) {
	Logger.Warn(fmt.Sprintf(msg, args...))
}

func Error(msg string, args ...interface{}) {
	Logger.Error(fmt.Sprintf(msg, args...))
}

func Debug(msg string, args ...interface{}) {
	Logger.Debug(fmt.Sprintf(msg, args...))
}
