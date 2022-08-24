package mw

import (
	"fmt"
	"net/http"

	"github.com/VictoriaMetrics/metrics"
	"github.com/heartwilltell/log"
)

// RecoveryMiddleware represents mw which catches and recovers from panics
// Returns the HTTP 500 (Internal Server Error) status if possible.
func RecoveryMiddleware(log log.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovery := recover(); recovery != nil && recovery != http.ErrAbortHandler {
					log.Error("Recovered form PANIC: %v", recovery)

					panicsTotal := fmt.Sprintf(`server_panics_total{method="%s", route="%s"}`, r.Method, r.URL.Path)
					metrics.GetOrCreateCounter(panicsTotal).Inc()

					w.WriteHeader(http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
