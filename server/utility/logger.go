package utility

import (
	"log/slog"
	"os"
)

type LoggerInterface interface {
	Debug(message string, args ...any)
	Info(message string, args ...any)
	Error(err error, args ...any)
	Warn(message string, args ...any)
	Fatal(message string, args ...any)
}

type LoggerInstance struct {
	logger *slog.Logger
}

// Global logger instance
var Logger *LoggerInstance

// Basic logger
func InitLogger(debug bool) *LoggerInstance {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	io_handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	logger := slog.New(io_handler)
	Logger = &LoggerInstance{logger: logger}
	return Logger
}

func (l *LoggerInstance) WithComponent(component string) *LoggerInstance {
	return &LoggerInstance{logger: l.logger.With("component", component)}
}

func (l *LoggerInstance) WithSessionID(sessionID string) *LoggerInstance {
	return &LoggerInstance{logger: l.logger.With("session_id", sessionID)}
}

func (l *LoggerInstance) Debug(message string, args ...any) {
	l.logger.Debug(message, args...)
}

func (l *LoggerInstance) Info(message string, args ...any) {
	l.logger.Info(message, args...)
}

func (l *LoggerInstance) Error(err error, args ...any) {
	l.logger.Error(err.Error(), args...)
}

func (l *LoggerInstance) Warn(message string, args ...any) {
	l.logger.Warn(message, args...)
}

// Fatal logs the message and exits the application with status code 1
func (l *LoggerInstance) Fatal(message string, args ...any) {
	l.logger.Error(message, args...)
	os.Exit(1)
}
