package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"
)

var debugMode atomic.Bool

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorGray   = "\033[90m"
)

type customLogWriter struct{}

func colorize(color string, text string) string {
	return color + text + colorReset
}

func colorizeLevel(level string) string {
	switch level {
	case "DEBUG":
		return colorize(colorBlue, level)
	case "INFO":
		return colorize(colorGreen, level)
	case "WARN":
		return colorize(colorYellow, level)
	case "ERROR":
		return colorize(colorRed, level)
	}
	return level
}

func (*customLogWriter) Write(p []byte) (n int, err error) {
	type logEntry struct {
		Time      string         `json:"time,omitempty"`
		Level     string         `json:"level,omitempty"`
		Message   string         `json:"message,omitempty"`
		Component string         `json:"component,omitempty"`
		Attrs     map[string]any `json:"attrs,omitempty"`
	}

	var raw map[string]any
	if err := json.Unmarshal(p, &raw); err != nil {
		return 0, err
	}

	entry := logEntry{
		Attrs: make(map[string]any),
	}

	if t, ok := raw["time"].(string); ok {
		entry.Time = t
		delete(raw, "time")
	}
	if l, ok := raw["level"].(string); ok {
		entry.Level = l
		delete(raw, "level")
	}
	if m, ok := raw["msg"].(string); ok {
		entry.Message = m
		delete(raw, "msg")
	} else if m, ok := raw["message"].(string); ok {
		entry.Message = m
		delete(raw, "message")
	}
	if component, ok := raw["component"].(string); ok {
		entry.Component = component
		delete(raw, "component")
	}

	attrs := ""
	for key, value := range raw {
		entry.Attrs[key] = value
		attrs += key + "=" + fmt.Sprintf("%v", value) + " "
	}
	displayTime := entry.Time
	if parsed, err := time.Parse(time.RFC3339Nano, entry.Time); err == nil {
		displayTime = parsed.Format(time.TimeOnly)
	} else if parsed, err := time.Parse(time.TimeOnly, entry.Time); err == nil {
		displayTime = parsed.Format(time.TimeOnly)
	}

	fmt.Println(
		displayTime,
		colorizeLevel(entry.Level),
		colorize(colorGray, fmt.Sprintf("[%s]",
			entry.Component)),
		fmt.Sprintf("'%s'", entry.Message),
		attrs)
	return len(p), nil
}

// levelFilterHandler gates DEBUG records on [debugMode]. Package-level loggers are
// constructed during init (before [configureLogging] runs), so the handler must not
// bake in a static min level.
type levelFilterHandler struct {
	inner slog.Handler
}

func (h *levelFilterHandler) Enabled(ctx context.Context, level slog.Level) bool {
	if level >= slog.LevelInfo {
		return h.inner.Enabled(ctx, level)
	}
	return debugMode.Load() && h.inner.Enabled(ctx, level)
}

func (h *levelFilterHandler) Handle(ctx context.Context, r slog.Record) error {
	if !h.Enabled(ctx, r.Level) {
		return nil
	}
	return h.inner.Handle(ctx, r)
}

func (h *levelFilterHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &levelFilterHandler{inner: h.inner.WithAttrs(attrs)}
}

func (h *levelFilterHandler) WithGroup(name string) slog.Handler {
	return &levelFilterHandler{inner: h.inner.WithGroup(name)}
}

func configureLogging(debug bool) {
	debugMode.Store(debug)
	if debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}
}

func fileLogger(name string) *slog.Logger {
	inner := slog.NewJSONHandler(&customLogWriter{}, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(&levelFilterHandler{inner: inner})
	return logger.With("component", name)
}
