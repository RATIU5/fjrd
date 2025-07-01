package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	default:
		return "info"
	}
}

func ParseLevel(s string) Level {
	switch strings.ToLower(s) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

func (l Level) ToSlog() slog.Level {
	switch l {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type Logger struct {
	*slog.Logger
	level Level
}

func New(level Level, output io.Writer) *Logger {
	if output == nil {
		output = os.Stderr
	}

	opts := &slog.HandlerOptions{
		Level: level.ToSlog(),
	}

	handler := slog.NewTextHandler(output, opts)
	logger := slog.New(handler)

	return &Logger{
		Logger: logger,
		level:  level,
	}
}

func NewJSON(level Level, output io.Writer) *Logger {
	if output == nil {
		output = os.Stderr
	}

	opts := &slog.HandlerOptions{
		Level: level.ToSlog(),
	}

	handler := slog.NewJSONHandler(output, opts)
	logger := slog.New(handler)

	return &Logger{
		Logger: logger,
		level:  level,
	}
}

func (l *Logger) Level() Level {
	return l.level
}

func (l *Logger) IsDebug() bool {
	return l.level <= LevelDebug
}

func (l *Logger) IsInfo() bool {
	return l.level <= LevelInfo
}

func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		Logger: l.Logger.With("component", component),
		level:  l.level,
	}
}

func (l *Logger) WithOperation(operation string) *Logger {
	return &Logger{
		Logger: l.Logger.With("operation", operation),
		level:  l.level,
	}
}

func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		Logger: l.Logger.With("error", err),
		level:  l.level,
	}
}