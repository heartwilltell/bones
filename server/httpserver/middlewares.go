package httpserver

import (
	"fmt"
	"net/http"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/heartwilltell/bones/ctxutil"
	"github.com/heartwilltell/log"
)

// LoggingMiddleware represents logging middleware.
func LoggingMiddleware(log log.Logger) Middleware {
	format := "%s %d %s Remote: %s %s Request ID: %s"
	errFormat := format + " Error: %s"

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now().UTC()

			var hookedError error

			ctx := ctxutil.SetErrorLogHook(r.Context(), hookedError)
			rid := ctx.Value(middleware.RequestIDKey)

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r.WithContext(ctx))
			status := ww.Status()

			if status >= http.StatusBadRequest {
				if hookedError != nil {
					log.Error(errFormat, r.Method, status, r.RequestURI, r.RemoteAddr, time.Since(start).String(), rid, hookedError)
					return
				}

				log.Error(format, r.Method, status, r.RequestURI, r.RemoteAddr, time.Since(start).String(), rid)
			} else {
				log.Info(format, r.Method, status, r.RequestURI, r.RemoteAddr, time.Since(start).String(), rid)
			}
		}

		return http.HandlerFunc(fn)
	}
}

// MetricsMiddleware represents HTTP metrics collecting middlewares.
func MetricsMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)

			httpReqDur := fmt.Sprintf(`http_request_duration{method="%s", route="%s", code="%d"}`, r.Method, r.URL.Path, ww.Status())
			metrics.GetOrCreateSummaryExt(httpReqDur, 5*time.Minute, []float64{0.95, 0.99}).UpdateDuration(start)

			httpReqTotal := fmt.Sprintf(`http_requests_total{method="%s", route="%s", code="%d"}`, r.Method, r.URL.Path, ww.Status())
			metrics.GetOrCreateCounter(httpReqTotal).Inc()
		}

		return http.HandlerFunc(fn)
	}
}
