package main

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
)

func configureLogging(ctx context.Context, opts *cliOptions) error {
	logFilePath := filepath.Join(opts.logDir, "log")
	w, logFileErr := os.OpenFile(
		logFilePath,
		os.O_CREATE|os.O_TRUNC|os.O_WRONLY,
		0644,
	)

	if logFileErr != nil {
		return logFileErr
	}

	logLevel := slog.LevelInfo
	if opts.debug {
		logLevel = slog.LevelDebug
	}

	slog.SetLogLoggerLevel(logLevel)
	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{Level: logLevel})
	logger := slog.New(handler)
	slog.SetDefault(logger)
	slog.InfoContext(
		ctx, "Logging configured",
		"filepath", logFilePath,
		"logLevel", logLevel,
	)
	return nil
}
