package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

type contextKey struct{}

// gcpHandler wraps a JSON slog.Handler and appends a "severity" field
// so Google Cloud Logging parses log levels correctly.
type gcpHandler struct {
	inner slog.Handler
}

func newGCPHandler(opts *slog.HandlerOptions) *gcpHandler {
	return &gcpHandler{inner: slog.NewJSONHandler(os.Stdout, opts)}
}

func (h *gcpHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *gcpHandler) Handle(ctx context.Context, r slog.Record) error {
	r.AddAttrs(slog.String("severity", strings.ToUpper(r.Level.String())))
	return h.inner.Handle(ctx, r)
}

func (h *gcpHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &gcpHandler{inner: h.inner.WithAttrs(attrs)}
}

func (h *gcpHandler) WithGroup(name string) slog.Handler {
	return &gcpHandler{inner: h.inner.WithGroup(name)}
}

// New creates a configured *slog.Logger.
// format="json" produces GCP-compatible JSON (production).
// format="text" produces human-readable output (development).
func New(level, format string) *slog.Logger {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(strings.ToUpper(level))); err != nil {
		lvl = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: lvl}

	var handler slog.Handler
	if format == "json" {
		handler = newGCPHandler(opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

// WithContext returns a context carrying the given logger.
func WithContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, l)
}

// FromContext returns the logger stored in ctx.
// Falls back to slog.Default() if none is present.
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(contextKey{}).(*slog.Logger); ok && l != nil {
		return l
	}
	return slog.Default()
}
