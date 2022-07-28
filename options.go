package bones

import (
	"fmt"
	"net/http/pprof"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/heartwilltell/bones/middlewares"
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

// serverConfig holds Server configuration.
type serverConfig struct {
	// hc - holds configuration for health endpoint.
	hc telemetryConfig

	// metrics - holds configuration for metrics endpoint.
	metrics telemetryConfig

	// profiler holds configuration fot profiler endpoint.
	profiler telemetryConfig
}

// telemetryConfig represents configuration of builtin telemetry routes,
// like: health-check, metrics, profiler, etc.
type telemetryConfig struct {
	enable                    bool
	accessLogsEnabled         bool
	metricsForEndpointEnabled bool
	route                     string
}

// defaultConfig returns serverConfig that holds
// default Server configuration.
func defaultConfig() serverConfig {
	config := serverConfig{
		hc: telemetryConfig{
			enable:                    true,
			accessLogsEnabled:         false,
			metricsForEndpointEnabled: false,
			route:                     "/health",
		},

		metrics: telemetryConfig{
			enable:                    true,
			accessLogsEnabled:         false,
			metricsForEndpointEnabled: false,
			route:                     "/metrics",
		},

		profiler: telemetryConfig{
			enable:                    false,
			accessLogsEnabled:         false,
			metricsForEndpointEnabled: false,
			route:                     "/profiler",
		},
	}

	return config
}

func (s *Server) applyConfig() error {
	if s.config.hc.enable {
		if s.config.hc.route == "" {
			return fmt.Errorf("empty healt-check route")
		}

		if !strings.HasPrefix("/", s.config.hc.route) {
			return fmt.Errorf("invalid healt-check route: %s (route should start with '/' slash)", s.config.hc.route)
		}

		s.router.Group(func(g chi.Router) {
			if s.config.hc.accessLogsEnabled {
				g.Use(middlewares.LoggingMiddleware(s.log))
			}

			if s.config.hc.metricsForEndpointEnabled {
				g.Use(middlewares.MetricsMiddleware())
			}

			g.Get(s.config.hc.route, s.healthCheck)
		})
	}

	if s.config.metrics.enable {
		if s.config.metrics.route == "" {
			return fmt.Errorf("empty metrics route")
		}

		if !strings.HasPrefix("/", s.config.metrics.route) {
			return fmt.Errorf("invalid metrics route: %s (route should start with '/' slash)", s.config.metrics.route)
		}

		s.router.Group(func(g chi.Router) {
			if s.config.metrics.accessLogsEnabled {
				g.Use(middlewares.LoggingMiddleware(s.log))
			}

			if s.config.metrics.metricsForEndpointEnabled {
				g.Use(middlewares.MetricsMiddleware())
			}

			g.Get(s.config.metrics.route, s.metrics)
		})
	}

	if s.config.profiler.enable {
		s.router.Group(func(g chi.Router) {
			if s.config.profiler.accessLogsEnabled {
				g.Use(middlewares.LoggingMiddleware(s.log))
			}

			g.Route("/debug/pprof", func(profiler chi.Router) {
				profiler.HandleFunc("/", pprof.Index)
				profiler.HandleFunc("/cmdline", pprof.Cmdline)
				profiler.HandleFunc("/profile", pprof.Profile)
				profiler.HandleFunc("/symbol", pprof.Symbol)
				profiler.HandleFunc("/trace", pprof.Trace)
			})
		})
	}

	return nil
}
