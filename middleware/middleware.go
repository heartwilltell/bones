package mw

import "net/http"

// Middleware represents an HTTP server middleware.
type Middleware = func(next http.Handler) http.Handler
