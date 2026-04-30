package middleware

import (
	"log/slog"
	"net/http"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	pkglogger "github.com/manuelgomezsw/loopi-api/pkg/logger"
)

// RequestLogger returns a middleware that logs each HTTP request as a
// structured slog record. It propagates a per-request logger (with the
// request_id already attached) into the request context so that service
// and domain code can retrieve it via logger.FromContext(ctx).
func RequestLogger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			reqID := chimiddleware.GetReqID(r.Context())

			reqLog := log.With("request_id", reqID)
			ctx := pkglogger.WithContext(r.Context(), reqLog)

			ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r.WithContext(ctx))

			reqLog.InfoContext(ctx, "request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"latency_ms", time.Since(start).Milliseconds(),
				"remote_ip", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)
		})
	}
}
