package middleware

import (
	"net/http"

	"github.com/heartwilltell/bones/bctx"
)

// RequestIDMiddleware tries to find different request IDs in
// request headers and set them to the request context.
func RequestIDMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			if id := r.Header.Get("lq-request-id"); id != "" {
				ctx = bctx.Set(ctx, bctx.RequestID, id)
			}

			if id := r.Header.Get("cf-ray"); id != "" {
				ctx = bctx.Set(ctx, bctx.RequestID, id)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}
