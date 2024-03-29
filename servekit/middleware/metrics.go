package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// MetricsMiddleware represents HTTP metrics collecting middleware.
func MetricsMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ctx := chi.RouteContext(r.Context())
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			httpReqDur := fmt.Sprintf(`http_request_duration{method="%s", route="%s", code="%d"}`,
				r.Method, ctx.RoutePattern(), ww.Status(),
			)
			metrics.GetOrCreateSummaryExt(httpReqDur, 5*time.Minute, []float64{0.95, 0.99}).UpdateDuration(start)

			httpReqTotal := fmt.Sprintf(`http_requests_total{method="%s", route="%s", code="%d"}`,
				r.Method, ctx.RoutePattern(), ww.Status(),
			)
			metrics.GetOrCreateCounter(httpReqTotal).Inc()
		}

		return http.HandlerFunc(fn)
	}
}
