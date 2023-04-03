package middleware

import (
	"fmt"
	"net/http"
)

func ForseHTTPSMiddleware(addr string) Middleware {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			url := fmt.Sprintf("https://%s%s", addr, r.URL.Path)
			http.Redirect(w, r, url, http.StatusMovedPermanently)
		}

		return http.HandlerFunc(fn)
	}
}
