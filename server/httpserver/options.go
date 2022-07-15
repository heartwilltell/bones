package httpserver

import (
	"time"

	"github.com/heartwilltell/log"
)

// Option represents functional options pattern for Server type, a function
// which receive a pointer to Server struct.
// Option functions can only be passed to Server constructor function New
// and can change the defaults of Server struct.
type Option func(server *Server)

// WithLogger sets the server logger.
func WithLogger(l log.Logger) Option { return func(s *Server) { s.log = l } }

// WithMiddlewares sets given middlewares as router wide middlewares.
func WithMiddlewares(m ...Middleware) Option { return func(s *Server) { s.router.Use(m...) } }

// WithReadTimeout sets the http.Server ReadTimeout.
func WithReadTimeout(t time.Duration) Option { return func(s *Server) { s.server.ReadTimeout = t } }

// WithWriteTimeout sets the http.Server WriteTimeout.
func WithWriteTimeout(t time.Duration) Option { return func(s *Server) { s.server.WriteTimeout = t } }

// WithIdleTimeout sets the http.Server IdleTimeout.
func WithIdleTimeout(t time.Duration) Option { return func(s *Server) { s.server.IdleTimeout = t } }
