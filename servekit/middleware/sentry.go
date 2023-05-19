package middleware

// func Sentry() Middleware {
// 	return func(next http.Handler) http.Handler {
// 		fn := func(w http.ResponseWriter, r *http.Request) {
// 			var sendErr error
//
// 			ctx := ctxkit.Set[func(err error)](r.Context(), "ctx.sentry", func(err error) { sendErr = err })
// 			next.ServeHTTP(w, r.WithContext(ctx))
//
// 			if sendErr != nil {
//
// 			}
// 		}
//
// 		return sentryhttp.New().HandleFunc(fn)
// 	}
// }
