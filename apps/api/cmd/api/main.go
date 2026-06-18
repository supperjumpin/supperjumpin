package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/supperjumpin/supperjumpin/apps/api/internal/httpapi"
)

func main() {
	logger, err := newLoggerFromEnv(os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	auth := httpapi.AuthVerifierChain{}
	if token := os.Getenv("SUPPERJUMPIN_DEV_AUTH_TOKEN"); token != "" {
		auth = append(auth, httpapi.StaticAuthVerifier{token: {
			Provider: "local-dev",
			Subject:  envOrDefault("SUPPERJUMPIN_DEV_AUTH_SUBJECT", "dev-player"),
			Email:    envOrDefault("SUPPERJUMPIN_DEV_AUTH_EMAIL", "player@example.com"),
		}})
	}

	databaseURL, err := requiredDatabaseURL()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	store, err := httpapi.NewPostgresStore(context.Background(), databaseURL)
	if err != nil {
		logger.Error("connect to Postgres", "err", err)
		os.Exit(1)
	}
	defer store.Close()

	server := httpapi.NewServer(httpapi.ServerConfig{
		Auth:         auth,
		Store:        store,
		Now:          time.Now,
		JumpPlanning: store,
		Judgment:     store,
		PublicRead:   store,
		Open:         store,
		CaptionEdit:  store,
		JumpRetract:  store,
		Logger:       logger,
	})
	logger.Info("Supperjumpin API listening", "port", port)
	if err := http.ListenAndServe(":"+port, server); err != nil {
		logger.Error("serve API", "err", err)
		os.Exit(1)
	}
}

func newLoggerFromEnv(w io.Writer) (*slog.Logger, error) {
	level, err := logLevelFromEnv()
	if err != nil {
		return nil, err
	}

	format := strings.ToLower(strings.TrimSpace(envOrDefault("SUPPERJUMPIN_LOG_FORMAT", "json")))
	opts := &slog.HandlerOptions{Level: level}
	switch format {
	case "json":
		return slog.New(slog.NewJSONHandler(w, opts)), nil
	case "text":
		return slog.New(slog.NewTextHandler(w, opts)), nil
	default:
		return nil, fmt.Errorf("SUPPERJUMPIN_LOG_FORMAT must be json or text, got %q", format)
	}
}

func logLevelFromEnv() (slog.Level, error) {
	level := strings.ToLower(strings.TrimSpace(envOrDefault("SUPPERJUMPIN_LOG_LEVEL", "info")))
	switch level {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("SUPPERJUMPIN_LOG_LEVEL must be debug, info, warn, or error, got %q", level)
	}
}

func envOrDefault(name string, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func requiredDatabaseURL() (string, error) {
	databaseURL := os.Getenv("SUPPERJUMPIN_DATABASE_URL")
	if databaseURL == "" {
		return "", fmt.Errorf("SUPPERJUMPIN_DATABASE_URL is required for durable Supperjumpin API state")
	}
	return databaseURL, nil
}
