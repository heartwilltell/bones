package mw

import (
	"net/http"
	"time"

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

			ctx := ctxutil.Set(r.Context(), ctxutil.ErrorLogHook, func(err error) { hookedError = err })
			rid := ctxutil.Get(ctx, ctxutil.RequestID)

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
