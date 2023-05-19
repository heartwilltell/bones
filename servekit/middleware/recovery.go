package middleware

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/VictoriaMetrics/metrics"
	"github.com/heartwilltell/bones/ctxkit"
	"github.com/heartwilltell/bones/errkit"
	"github.com/heartwilltell/log"
)

// RecoveryMiddleware represents middleware which catches and recovers from panics.
// Returns the HTTP 500 (Internal Server Error) status if possible.
func RecoveryMiddleware(log log.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovery := recover(); recovery != nil && isAbortHandlerError(recovery) {
					panicErr := recoveryValueToError(recovery)

					log.Error("Recovered form PANIC: %s", panicErr)

					if hook := ctxkit.GetLogErrHook(r.Context()); hook != nil {
						hook(recoveryValueToError(recovery))
					}

					m := fmt.Sprintf(`server_panics_total{method="%s", route="%s"}`, r.Method, r.URL.Path)
					metrics.GetOrCreateCounter(m).Inc()

					errkit.Report(panicErr)

					w.WriteHeader(http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

func isAbortHandlerError(recovery any) bool {
	if recoveryErr, ok := recovery.(error); ok && errors.Is(recoveryErr, http.ErrAbortHandler) {
		return true
	}

	return false
}

func recoveryValueToError(recovery any) error {
	if recoveryErr, ok := recovery.(error); ok {
		return recoveryErr
	}

	return fmt.Errorf("recover value: %+v", recovery)
}
