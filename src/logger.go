package main

import (
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

func configureLogging(debug bool) {
	debugMode.Store(debug)
}

func fileLogger(name string) *slog.Logger {
	level := slog.LevelInfo
	if debugMode.Load() {
		level = slog.LevelDebug
	}
	handler := slog.NewJSONHandler(&customLogWriter{}, &slog.HandlerOptions{
		Level: level,
	})
	logger := slog.New(handler)
	slog.SetLogLoggerLevel(level)
	return logger.With("component", name)
}
