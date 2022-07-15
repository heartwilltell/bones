package grpcserver

import (
	"github.com/heartwilltell/log"
)

// Option represents functional options pattern for Server type, a function
// which receive a pointer to Server struct.
// Option functions can only be passed to Server constructor function New
// and can change the defaults of Server struct.
type Option func(server *Server)

// WithLogger sets the server logger.
func WithLogger(l log.Logger) Option { return func(s *Server) { s.log = l } }
