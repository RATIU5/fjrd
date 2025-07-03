package logger

import (
	"context"
	"fmt"
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

	handler := NewCustomHandler(output, opts)
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

type CustomHandler struct {
	textHandler slog.Handler
	output      io.Writer
	level       slog.Level
}

func NewCustomHandler(output io.Writer, opts *slog.HandlerOptions) *CustomHandler {
	if output == nil {
		output = os.Stderr
	}
	
	return &CustomHandler{
		textHandler: slog.NewTextHandler(output, opts),
		output:      output,
		level:       opts.Level.Level(),
	}
}

func (h *CustomHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *CustomHandler) Handle(ctx context.Context, record slog.Record) error {
	if record.Level == slog.LevelInfo {
		// For Info level, output only the message
		_, err := fmt.Fprintln(h.output, record.Message)
		return err
	}
	
	// For all other levels, use the standard text handler
	return h.textHandler.Handle(ctx, record)
}

func (h *CustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &CustomHandler{
		textHandler: h.textHandler.WithAttrs(attrs),
		output:      h.output,
		level:       h.level,
	}
}

func (h *CustomHandler) WithGroup(name string) slog.Handler {
	return &CustomHandler{
		textHandler: h.textHandler.WithGroup(name),
		output:      h.output,
		level:       h.level,
	}
}