package middleware

import (
	"net/http"

	"github.com/heartwilltell/bones/ctxkit"
	"github.com/heartwilltell/bones/idkit"
)

// RequestIDMiddleware tries to find request IDs in
// request headers and set it to the request context.
func RequestIDMiddleware(header ...string) Middleware {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			if len(header) != 0 {
				if rid := r.Header.Get(header[0]); rid != "" {
					ctx = ctxkit.SetRequestID(ctx, rid)
				}
			} else {
				ctx = ctxkit.SetRequestID(ctx, idkit.ULID())
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}
