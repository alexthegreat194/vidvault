package main

import (
	"log/slog"
	"os"
	"sync/atomic"
)

var debugMode atomic.Bool

func configureLogging(debug bool) {
	debugMode.Store(debug)
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(handler))
}

func fileLogger(name string) *slog.Logger {
	return slog.Default().With("component", name)
}

func logDebug(logger *slog.Logger, msg string, args ...any) {
	if debugMode.Load() {
		logger.Debug(msg, args...)
	}
}
