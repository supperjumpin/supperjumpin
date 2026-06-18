package main

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestRequiredDatabaseURLReturnsSUPPERJUMPINDatabaseURL(t *testing.T) {
	t.Setenv("SUPPERJUMPIN_DATABASE_URL", "postgres://user:pass@primary:5432/supperjumpin?sslmode=disable")
	t.Setenv("DATABASE_URL", "postgres://user:pass@ambient:5432/supperjumpin?sslmode=disable")

	got, err := requiredDatabaseURL()
	if err != nil {
		t.Fatalf("expected database URL, got error: %v", err)
	}

	want := "postgres://user:pass@primary:5432/supperjumpin?sslmode=disable"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestRequiredDatabaseURLFailsWithoutSUPPERJUMPINDatabaseURL(t *testing.T) {
	t.Setenv("SUPPERJUMPIN_DATABASE_URL", "")
	t.Setenv("DATABASE_URL", "postgres://user:pass@ambient:5432/supperjumpin?sslmode=disable")

	got, err := requiredDatabaseURL()
	if err == nil {
		t.Fatalf("expected error, got database URL %q", got)
	}

	want := "SUPPERJUMPIN_DATABASE_URL is required for durable Supperjumpin API state"
	if err.Error() != want {
		t.Fatalf("expected error %q, got %q", want, err.Error())
	}
}

func TestNewLoggerFromEnvDefaultsToJSONInfo(t *testing.T) {
	t.Setenv("SUPPERJUMPIN_LOG_FORMAT", "")
	t.Setenv("SUPPERJUMPIN_LOG_LEVEL", "")

	var logs bytes.Buffer
	logger, err := newLoggerFromEnv(&logs)
	if err != nil {
		t.Fatalf("expected logger, got error: %v", err)
	}

	logger.Debug("hidden")
	logger.Info("visible")

	got := logs.String()
	if strings.Contains(got, "hidden") {
		t.Fatalf("expected debug log to be filtered, got %q", got)
	}
	if !strings.Contains(got, `"msg":"visible"`) {
		t.Fatalf("expected JSON info log, got %q", got)
	}
}

func TestNewLoggerFromEnvSupportsTextDebug(t *testing.T) {
	t.Setenv("SUPPERJUMPIN_LOG_FORMAT", "text")
	t.Setenv("SUPPERJUMPIN_LOG_LEVEL", "debug")

	var logs bytes.Buffer
	logger, err := newLoggerFromEnv(&logs)
	if err != nil {
		t.Fatalf("expected logger, got error: %v", err)
	}

	logger.Debug("visible")

	got := logs.String()
	if !strings.Contains(got, "level=DEBUG") || !strings.Contains(got, "msg=visible") {
		t.Fatalf("expected text debug log, got %q", got)
	}
}

func TestNewLoggerFromEnvRejectsInvalidFormat(t *testing.T) {
	t.Setenv("SUPPERJUMPIN_LOG_FORMAT", "xml")
	t.Setenv("SUPPERJUMPIN_LOG_LEVEL", "info")

	_, err := newLoggerFromEnv(&bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error")
	}

	want := `SUPPERJUMPIN_LOG_FORMAT must be json or text, got "xml"`
	if err.Error() != want {
		t.Fatalf("expected %q, got %q", want, err.Error())
	}
}

func TestLogLevelFromEnvRejectsInvalidLevel(t *testing.T) {
	t.Setenv("SUPPERJUMPIN_LOG_LEVEL", "trace")

	level, err := logLevelFromEnv()
	if err == nil {
		t.Fatalf("expected error, got level %v", level)
	}

	want := `SUPPERJUMPIN_LOG_LEVEL must be debug, info, warn, or error, got "trace"`
	if err.Error() != want {
		t.Fatalf("expected %q, got %q", want, err.Error())
	}
}

func TestLogLevelFromEnvParsesWarn(t *testing.T) {
	t.Setenv("SUPPERJUMPIN_LOG_LEVEL", "warn")

	level, err := logLevelFromEnv()
	if err != nil {
		t.Fatalf("expected level, got error: %v", err)
	}
	if level != slog.LevelWarn {
		t.Fatalf("expected warn, got %v", level)
	}
}
