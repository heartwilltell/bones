package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/heartwilltell/bones/ctxkit"
	"github.com/heartwilltell/bones/errkit"
	"github.com/heartwilltell/log"
)

// LoggingMiddleware represents logging middleware.
func LoggingMiddleware(log log.Logger) Middleware {
	format := "%s %d %s Remote: %s %s"
	errFormat := format + " Error: %s"

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now().UTC()

			var hookedError error

			ctx := ctxkit.SetLogErrHook(r.Context(), func(err error) { hookedError = err })

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r.WithContext(ctx))
			status := ww.Status()

			if status >= http.StatusBadRequest {
				if hookedError != nil {
					log.Error(errFormat, r.Method, status, r.RequestURI, r.RemoteAddr, time.Since(start).String(), hookedError)

					errkit.Report(hookedError)
					return
				}

				log.Error(format, r.Method, status, r.RequestURI, r.RemoteAddr, time.Since(start).String())
			} else {
				log.Info(format, r.Method, status, r.RequestURI, r.RemoteAddr, time.Since(start).String())
			}
		}

		return http.HandlerFunc(fn)
	}
}
