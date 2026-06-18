package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

const requestIDHeader = "X-Request-ID"

type requestIDContextKey struct{}
type requestLogContextKey struct{}

type requestLogContext struct {
	mu    sync.Mutex
	level slog.Level
	attrs []slog.Attr
}

func RequestIDFromContext(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(requestIDContextKey{}).(string)
	return requestID, ok && requestID != ""
}

func AddRequestLogField(ctx context.Context, name string, value any) {
	logCtx, ok := ctx.Value(requestLogContextKey{}).(*requestLogContext)
	if !ok || !approvedRequestLogField(name) {
		return
	}

	logCtx.mu.Lock()
	defer logCtx.mu.Unlock()
	logCtx.attrs = append(logCtx.attrs, slog.Any(name, value))
}

func AddRequestLogFields(ctx context.Context, attrs ...slog.Attr) {
	for _, attr := range attrs {
		AddRequestLogField(ctx, attr.Key, attr.Value.Any())
	}
}

func RaiseRequestLogLevel(ctx context.Context, level slog.Level) {
	logCtx, ok := ctx.Value(requestLogContextKey{}).(*requestLogContext)
	if !ok {
		return
	}

	logCtx.mu.Lock()
	defer logCtx.mu.Unlock()
	if level > logCtx.level {
		logCtx.level = level
	}
}

func requestLoggingMiddleware(next http.Handler, logger *slog.Logger, now func() time.Time) http.Handler {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	if now == nil {
		now = time.Now
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := now()
		requestID := canonicalRequestID(r.Header.Get(requestIDHeader))
		if requestID == "" {
			var err error
			requestID, err = newRequestID()
			if err != nil {
				http.Error(w, "create request id", http.StatusInternalServerError)
				return
			}
		}

		logCtx := &requestLogContext{level: slog.LevelInfo}
		ctx := context.WithValue(r.Context(), requestIDContextKey{}, requestID)
		ctx = context.WithValue(ctx, requestLogContextKey{}, logCtx)
		r = r.WithContext(ctx)

		rec := &responseCapture{ResponseWriter: w, status: http.StatusOK}
		rec.Header().Set(requestIDHeader, requestID)

		defer func() {
			if recovered := recover(); recovered != nil {
				RaiseRequestLogLevel(r.Context(), slog.LevelError)
				AddRequestLogField(r.Context(), "outcome", "server_error")
				AddRequestLogField(r.Context(), "error_code", "panic")
				AddRequestLogField(r.Context(), "stack", string(debug.Stack()))
				if !rec.wroteHeader {
					writeAPIError(rec, http.StatusInternalServerError, "internal_error", "Internal server error.")
				}
			}

			if rec.status >= http.StatusInternalServerError {
				RaiseRequestLogLevel(r.Context(), slog.LevelError)
				if !hasRequestLogField(r.Context(), "stack") {
					AddRequestLogField(r.Context(), "stack", string(debug.Stack()))
				}
			}
			if !hasRequestLogField(r.Context(), "outcome") {
				AddRequestLogField(r.Context(), "outcome", outcomeForStatus(rec.status))
			}

			duration := now().Sub(start)
			attrs := []slog.Attr{
				slog.String("request_id", requestID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rec.status),
				slog.Int64("response_bytes", rec.bytes),
				slog.Float64("duration_ms", float64(duration.Microseconds())/1000),
				slog.String("user_agent", r.UserAgent()),
			}

			logCtx.mu.Lock()
			level := logCtx.level
			attrs = append(attrs, logCtx.attrs...)
			logCtx.mu.Unlock()

			logger.LogAttrs(context.Background(), level, "request completed", attrs...)
		}()

		next.ServeHTTP(rec, r)
	})
}

func hasRequestLogField(ctx context.Context, name string) bool {
	logCtx, ok := ctx.Value(requestLogContextKey{}).(*requestLogContext)
	if !ok {
		return false
	}

	logCtx.mu.Lock()
	defer logCtx.mu.Unlock()
	for _, attr := range logCtx.attrs {
		if attr.Key == name {
			return true
		}
	}
	return false
}

func outcomeForStatus(status int) string {
	switch {
	case status >= http.StatusInternalServerError:
		return "server_error"
	case status == http.StatusConflict:
		return "conflict"
	case status == http.StatusNotFound:
		return "not_found"
	case status == http.StatusForbidden || status == http.StatusUnauthorized:
		return "forbidden"
	case status >= http.StatusBadRequest:
		return "client_error"
	default:
		return "success"
	}
}

func safeInternalError(err error) string {
	if err == nil {
		return ""
	}
	return "redacted internal error"
}

type responseCapture struct {
	http.ResponseWriter
	status      int
	bytes       int64
	wroteHeader bool
}

func (w *responseCapture) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *responseCapture) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.status = status
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseCapture) Write(body []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(body)
	w.bytes += int64(n)
	return n, err
}

func canonicalRequestID(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if len(value) != 36 {
		return ""
	}
	for i, c := range value {
		switch i {
		case 8, 13, 18, 23:
			if c != '-' {
				return ""
			}
		default:
			if !isHex(c) {
				return ""
			}
		}
	}
	return value
}

func isHex(c rune) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')
}

func newRequestID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	var dst [36]byte
	hex.Encode(dst[0:8], b[0:4])
	dst[8] = '-'
	hex.Encode(dst[9:13], b[4:6])
	dst[13] = '-'
	hex.Encode(dst[14:18], b[6:8])
	dst[18] = '-'
	hex.Encode(dst[19:23], b[8:10])
	dst[23] = '-'
	hex.Encode(dst[24:36], b[10:16])
	return string(dst[:]), nil
}

func approvedRequestLogField(name string) bool {
	switch strings.TrimSpace(name) {
	case "actor_type", "error_code", "internal_error", "judgment_id", "jump_id", "open_month", "open_year", "operation", "outcome", "player_id", "route", "stack":
		return true
	default:
		return false
	}
}
