package middlewares

import (
	"net/http"

	"github.com/heartwilltell/bones"
	"github.com/heartwilltell/bones/ctxutil"
)

// RequestIDMiddleware tries to find different request IDs in
// request headers and set them to the request context.
func RequestIDMiddleware() bones.Middleware {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			if id := r.Header.Get("lq-request-id"); id != "" {
				ctx = ctxutil.SetRequestID(ctx, id)
			}

			if id := r.Header.Get("cf-ray"); id != "" {
				ctx = ctxutil.SetCFRequestID(ctx, id)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}
