package httpapi

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestRequestLoggingGeneratesRequestID(t *testing.T) {
	server, logs := newLoggingTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID, ok := RequestIDFromContext(r.Context())
		if !ok {
			t.Fatal("request ID missing from context")
		}
		if got := w.Header().Get(requestIDHeader); got != requestID {
			t.Fatalf("expected response header %q, got %q", requestID, got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/me", nil))

	requestID := rec.Header().Get(requestIDHeader)
	if !uuidPattern.MatchString(requestID) {
		t.Fatalf("expected generated UUID v4 request ID, got %q", requestID)
	}
	entry := decodeLogEntry(t, logs)
	if entry["request_id"] != requestID {
		t.Fatalf("expected log request_id %q, got %#v", requestID, entry["request_id"])
	}
}

func TestRequestLoggingPreservesIncomingCanonicalUUID(t *testing.T) {
	server, logs := newLoggingTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	want := "11111111-2222-4333-8444-555555555555"
	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req.Header.Set(requestIDHeader, want)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if got := rec.Header().Get(requestIDHeader); got != want {
		t.Fatalf("expected preserved request ID %q, got %q", want, got)
	}
	entry := decodeLogEntry(t, logs)
	if entry["request_id"] != want {
		t.Fatalf("expected log request_id %q, got %#v", want, entry["request_id"])
	}
}

func TestRequestLoggingReplacesInvalidIncomingRequestID(t *testing.T) {
	server, _ := newLoggingTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req.Header.Set(requestIDHeader, "not-a-uuid")
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	requestID := rec.Header().Get(requestIDHeader)
	if requestID == "not-a-uuid" || !uuidPattern.MatchString(requestID) {
		t.Fatalf("expected replacement UUID v4 request ID, got %q", requestID)
	}
}

func TestRequestLoggingCapturesResponseAndAccumulatorFields(t *testing.T) {
	server, logs := newLoggingTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AddRequestLogField(r.Context(), "jump_id", "jump_123")
		AddRequestLogField(r.Context(), "remote_addr", "127.0.0.1")
		AddRequestLogField(r.Context(), "status", http.StatusTeapot)
		RaiseRequestLogLevel(r.Context(), slog.LevelDebug)
		RaiseRequestLogLevel(r.Context(), slog.LevelWarn)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	}))

	req := httptest.NewRequest(http.MethodPost, "/v1/jumps", nil)
	req.Header.Set("User-Agent", "supperjumpin-test")
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	entry := decodeLogEntry(t, logs)
	if entry["level"] != "WARN" {
		t.Fatalf("expected WARN log level, got %#v", entry["level"])
	}
	if entry["status"] != float64(http.StatusCreated) {
		t.Fatalf("expected status 201, got %#v", entry["status"])
	}
	if entry["response_bytes"] != float64(len("created")) {
		t.Fatalf("expected response byte count, got %#v", entry["response_bytes"])
	}
	if _, ok := entry["duration_ms"].(float64); !ok {
		t.Fatalf("expected numeric duration_ms, got %#v", entry["duration_ms"])
	}
	if entry["user_agent"] != "supperjumpin-test" {
		t.Fatalf("expected user agent, got %#v", entry["user_agent"])
	}
	if _, ok := entry["remote_addr"]; ok {
		t.Fatalf("remote_addr should not be logged: %#v", entry)
	}
	if entry["status"] != float64(http.StatusCreated) {
		t.Fatalf("accumulated fields should not override canonical status, got %#v", entry["status"])
	}
	if entry["jump_id"] != "jump_123" {
		t.Fatalf("expected accumulated field, got %#v", entry["jump_id"])
	}
}

func TestRequestLoggingRecoversPanicBeforeResponse(t *testing.T) {
	server, logs := newLoggingTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/panic", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode panic response: %v", err)
	}
	if body["error"] != "internal_error" || body["message"] != "Internal server error." {
		t.Fatalf("expected generic internal_error response, got %#v", body)
	}

	entry := decodeLogEntry(t, logs)
	if entry["level"] != "ERROR" {
		t.Fatalf("expected ERROR log level, got %#v", entry["level"])
	}
	if entry["status"] != float64(http.StatusInternalServerError) {
		t.Fatalf("expected status 500, got %#v", entry["status"])
	}
	if _, ok := entry["panic"]; ok {
		t.Fatalf("panic value should not be logged: %#v", entry["panic"])
	}
	stack, ok := entry["stack"].(string)
	if !ok || !strings.Contains(stack, "runtime/debug.Stack") {
		t.Fatalf("expected stack trace, got %#v", entry["stack"])
	}
}

func TestResponseCaptureUnwrapsUnderlyingResponseWriter(t *testing.T) {
	underlying := httptest.NewRecorder()
	rec := &responseCapture{ResponseWriter: underlying, status: http.StatusOK}

	if got := rec.Unwrap(); got != underlying {
		t.Fatalf("expected underlying response writer")
	}
}

func newLoggingTestServer(next http.Handler) (http.Handler, *bytes.Buffer) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))
	now := fixedClock(time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC), time.Date(2026, 6, 17, 12, 0, 0, 1500000, time.UTC))
	return requestLoggingMiddleware(next, logger, now), &logs
}

func fixedClock(values ...time.Time) func() time.Time {
	call := 0
	return func() time.Time {
		if call >= len(values) {
			return values[len(values)-1]
		}
		value := values[call]
		call++
		return value
	}
}

func decodeLogEntry(t *testing.T, logs *bytes.Buffer) map[string]any {
	t.Helper()
	var entry map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(logs.Bytes()), &entry); err != nil {
		t.Fatalf("decode log entry %q: %v", logs.String(), err)
	}
	return entry
}
