package bones

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/pprof"
	"strings"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/go-chi/chi/v5"
	"github.com/heartwilltell/bones/middleware"
	"github.com/heartwilltell/bones/respond"
	"github.com/heartwilltell/hc"
	"github.com/heartwilltell/log"
	"golang.org/x/sync/errgroup"
)

const (
	// readTimeout represents default read timeout for the http.Server.
	readTimeout = 10 * time.Second

	// readHeaderTimeout represents default read header timeout for http.Server.
	readHeaderTimeout = 5 * time.Second

	// writeTimeout represents default write timeout for the http.Server.
	writeTimeout = 10 * time.Second

	// idleTimeout represents default idle timeout for the http.Server.
	idleTimeout = 90 * time.Second

	// shutdownTimeout represents server default shutdown timeout.
	shutdownTimeout = 5 * time.Second
)

// Error represents package level errors.
type Error string

func (e Error) Error() string { return string(e) }

// Middleware represents an http.Handler middleware.
type Middleware = func(next http.Handler) http.Handler

// Option implements functional options pattern for Server type.
// Represents a function which receive a pointer to the config struct
// and changes it default values to the given ones.
//
// See the Server.configure function to understand the configuration behaviour.
// Option functions should only be passed to Server constructor function New.
type Option[T any] func(o T)

// WithLogger sets the server logger.
func WithLogger(l log.Logger) Option[*config] { return func(c *config) { c.logger = l } }

// WithHealthCheck turns on the health check endpoint.
// Receives the following option to configure the endpoint:
// - SetHealthChecker - to change the HealthChecker implementation.
// - SetHealthEndpointRoute - to set the endpoint route.
// - SetHealthEndpointAccessLog - to enable access log for endpoint.
// - SetHealthEndpointMetricsCollection - to enable metrics collection for endpoint.
func WithHealthCheck(options ...Option[*healthConfig]) Option[*config] {
	return func(c *config) {
		c.health.enable = true

		for _, opt := range options {
			opt(&c.health)
		}
	}
}

// SetHealthChecker represents an optional function for WithHealthCheck function.
// If passed to the WithHealthCheck, will set the config.health.healthChecker.
func SetHealthChecker(hc hc.HealthChecker) Option[*healthConfig] {
	return func(c *healthConfig) {
		// To not shoot in the leg. There are already a nop checker.
		if hc == nil {
			return
		}

		c.healthChecker = hc
	}
}

// SetHealthEndpointRoute represents an optional function for WithHealthCheck function.
// If passed to the WithHealthCheck, will set the config.health.route.
func SetHealthEndpointRoute(route string) Option[*healthConfig] {
	return func(c *healthConfig) { c.route = route }
}

// SetHealthEndpointAccessLog represents an optional function for WithHealthCheck function.
// If passed to the WithHealthCheck, will set the config.health.accessLogsEnabled to true.
func SetHealthEndpointAccessLog() Option[*healthConfig] {
	return func(c *healthConfig) { c.accessLogsEnabled = true }
}

// SetHealthEndpointMetricsCollection represents an optional function for WithHealthCheck function.
// If passed to the WithHealthCheck, will set the config.health.metricsForEndpointEnabled to true.
func SetHealthEndpointMetricsCollection() Option[*healthConfig] {
	return func(c *healthConfig) { c.metricsForEndpointEnabled = true }
}

// WithMetrics turns on the metrics endpoint.
// Receives the following option to configure the endpoint:
// - SetMetricsEndpointRoute - to set the endpoint route.
// - SetMetricsEndpointAccessLog - to enable access log for endpoint.
// - SetMetricsEndpointMetricsCollection - to enable metrics collection for endpoint.
func WithMetrics(options ...Option[*metricsConfig]) Option[*config] {
	return func(c *config) {
		c.metrics.enable = true

		for _, opt := range options {
			opt(&c.metrics)
		}
	}
}

// SetMetricsEndpointRoute represents an optional function for WithMetrics function.
// If passed to the WithMetrics, will set the config.health.route.
func SetMetricsEndpointRoute(route string) Option[*metricsConfig] {
	return func(c *metricsConfig) { c.route = route }
}

// SetMetricsEndpointAccessLog represents an optional function for WithMetrics function.
// If passed to the WithMetrics, will set the config.health.accessLogsEnabled to true.
func SetMetricsEndpointAccessLog() Option[*metricsConfig] {
	return func(c *metricsConfig) { c.accessLogsEnabled = true }
}

// SetMetricsEndpointMetricsCollection represents an optional function for WithMetrics function.
// If passed to the WithMetrics, will set the config.health.metricsForEndpointEnabled to true.
func SetMetricsEndpointMetricsCollection() Option[*metricsConfig] {
	return func(c *metricsConfig) { c.metricsForEndpointEnabled = true }
}

// WithGlobalMiddlewares sets given middlewares as router-wide middlewares.
// Means that they will be applied to each server endpoint.
func WithGlobalMiddlewares(m ...Middleware) Option[*config] {
	return func(c *config) { c.globalMiddlewares = append(c.globalMiddlewares, m...) }
}

// WithReadTimeout sets the http.Server ReadTimeout.
func WithReadTimeout(t time.Duration) Option[*config] {
	return func(c *config) { c.readTimeout = t }
}

// WithWriteTimeout sets the http.Server WriteTimeout.
func WithWriteTimeout(t time.Duration) Option[*config] {
	return func(s *config) { s.writeTimeout = t }
}

// WithIdleTimeout sets the http.Server IdleTimeout.
func WithIdleTimeout(t time.Duration) Option[*config] {
	return func(s *config) { s.idleTimeout = t }
}

// Server holds all app logic in form of http.Handler interfaces
// along with routing logic and HTTP transport logic.
type Server struct {
	health hc.HealthChecker
	logger log.Logger
	router chi.Router
	server *http.Server
}

// New return a new instance of Server struct.
func New(addr string, options ...Option[*config]) (*Server, error) {
	router := chi.NewRouter()

	s := Server{
		logger: log.NewNopLog(),
		health: hc.NewNopChecker(),
		router: router,
		server: &http.Server{
			Addr:    addr,
			Handler: router,
		},
	}

	// To not keep a lot of unnecessary stuff on the Server struct instance,
	// all the options will be applied to config struct, which will be used
	// as the source of truth to apply the final configuration to the Server.
	if err := s.configure(options...); err != nil {
		return nil, fmt.Errorf("failed to apply server cofiguration: %w", err)
	}

	return &s, nil
}

func (s *Server) Mount(route string, handler http.Handler, middlewares ...Middleware) {
	s.router.Route(route, func(r chi.Router) {
		r.Use(middlewares...)
		r.Mount("/", handler)
	})
}

// Serve listen to incoming connections and serves each request.
func (s *Server) Serve(ctx context.Context) error {
	if s.server.Addr == "" {
		return fmt.Errorf("invalid server address: %s", s.server.Addr)
	}

	g, sctx := errgroup.WithContext(ctx)

	// handle shutdown signal in the background
	g.Go(func() error { return s.handleShutdown(sctx) })

	g.Go(func() error {
		s.logger.Info("Server started to listen on: %s", s.server.Addr)

		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server failed: %w", err)
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		s.logger.Error("Force exit! %s", err.Error())
		panic(err)
	}

	s.logger.Info("Bye!")

	return nil
}

func (s *Server) ServeTLS(ctx context.Context, cert, key string) error {
	if s.server.Addr == "" {
		return fmt.Errorf("invalid server address: %s", s.server.Addr)
	}

	g, sctx := errgroup.WithContext(ctx)

	// handle shutdown signal in the background
	g.Go(func() error { return s.handleShutdown(sctx) })

	g.Go(func() error {
		s.logger.Info("Server started to listen on: %s", s.server.Addr)

		if err := s.server.ListenAndServeTLS(cert, key); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server failed: %w", err)
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		s.logger.Error("Force exit! %s", err.Error())
		panic(err)
	}

	s.logger.Info("Bye!")

	return nil
}

func (s *Server) metrics(w http.ResponseWriter, _ *http.Request) {
	metrics.WritePrometheus(w, true)
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	if s.health == nil {
		respond.TEXT(w, r, http.StatusOK, nil)
		return
	}

	if err := s.health.Health(r.Context()); err != nil {
		respond.Error(w, r, err)
		return
	}

	respond.TEXT(w, r, http.StatusOK, nil)
}

// handleShutdown blocks until select statement receives a signal from
// ctx.Done, after that new context.WithTimeout will be created and passed to
// http.Server Shutdown method.
//
// If Shutdown method returns non nil error, program will panic immediately.
func (s *Server) handleShutdown(ctx context.Context) error {
	<-ctx.Done()

	s.logger.Info("Shutting down the server!")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	s.server.SetKeepAlivesEnabled(false)

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shutdown the server gracefully: %w", err)
	}

	return nil
}

func (s *Server) configure(options ...Option[*config]) error {
	// Initialize the default settings.
	cfg := defaultConfig()

	// Apply all server options to the config struct.
	for _, opt := range options {
		opt(&cfg)
	}

	// Apply logger settings.
	s.logger = cfg.logger

	// Apply health checker settings.
	s.health = cfg.health.healthChecker

	// Apply server timeouts.
	s.server.ReadTimeout = cfg.readTimeout
	s.server.ReadHeaderTimeout = cfg.readHeaderTimeout
	s.server.WriteTimeout = cfg.writeTimeout
	s.server.IdleTimeout = cfg.idleTimeout

	// Apply router-wide middlewares.
	s.router.Use(cfg.globalMiddlewares...)

	if cfg.health.enable {
		if cfg.health.route == "" {
			return fmt.Errorf("invalid healt-check route: %s (should not be empty)", cfg.health.route)
		}

		if !strings.HasPrefix(cfg.health.route, "/") {
			return fmt.Errorf("invalid healt-check route: %s (route should start with '/' slash)", cfg.health.route)
		}

		s.router.Group(func(g chi.Router) {
			if cfg.health.accessLogsEnabled {
				g.Use(middleware.LoggingMiddleware(s.logger))
			}

			if cfg.health.metricsForEndpointEnabled {
				g.Use(middleware.MetricsMiddleware())
			}

			g.Get(cfg.health.route, s.healthCheck)
		})
	}

	if cfg.metrics.enable {
		if cfg.metrics.route == "" {
			return fmt.Errorf("invalid metrics route: %s (should not be empty)", cfg.metrics.route)
		}

		if !strings.HasPrefix(cfg.metrics.route, "/") {
			return fmt.Errorf("invalid metrics route: %s (route should start with '/' slash)", cfg.metrics.route)
		}

		s.router.Group(func(g chi.Router) {
			if cfg.metrics.accessLogsEnabled {
				g.Use(middleware.LoggingMiddleware(s.logger))
			}

			if cfg.metrics.metricsForEndpointEnabled {
				g.Use(middleware.MetricsMiddleware())
			}

			g.Get(cfg.metrics.route, s.metrics)
		})
	}

	if cfg.profiler.enable {
		s.router.Group(func(g chi.Router) {
			if cfg.profiler.accessLogsEnabled {
				g.Use(middleware.LoggingMiddleware(s.logger))
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

// defaultConfig returns an instance os config with default values.
func defaultConfig() config {
	cfg := config{
		logger: log.NewNopLog(),

		readTimeout:       readTimeout,
		readHeaderTimeout: readHeaderTimeout,
		writeTimeout:      writeTimeout,
		idleTimeout:       idleTimeout,

		globalMiddlewares: make([]Middleware, 0, 0),

		health: healthConfig{
			enable:                    false,
			accessLogsEnabled:         false,
			metricsForEndpointEnabled: false,
			route:                     "/health",
			healthChecker:             hc.NewNopChecker(),
		},

		metrics: metricsConfig{
			enable:                    false,
			accessLogsEnabled:         false,
			metricsForEndpointEnabled: false,
			route:                     "/metrics",
		},

		profiler: profilerConfig{
			enable:            false,
			accessLogsEnabled: false,
			route:             "/profiler",
		},
	}

	return cfg
}

// config holds Server configuration.
type config struct {
	// logger represents a log instance which will used by Server.
	logger log.Logger

	// readTimeout represents the http.Server ReadTimeout.
	readTimeout time.Duration

	// readHeaderTimeout represents the http.Server ReadHeaderTimeout.
	readHeaderTimeout time.Duration

	// writeTimeout represents the http.Server WriteTimeout.
	writeTimeout time.Duration

	// idleTimeout represents the http.Server IdleTimeout.
	idleTimeout time.Duration

	globalMiddlewares []Middleware

	// health holds configuration for health endpoint.
	health healthConfig

	// metrics holds configuration for metrics endpoint.
	metrics metricsConfig

	// profiler holds configuration fot profiler endpoint.
	profiler profilerConfig
}

// metricsConfig represents configuration of builtin metrics route.
type metricsConfig struct {
	enable                    bool
	accessLogsEnabled         bool
	metricsForEndpointEnabled bool
	route                     string
}

// healthConfig represents configuration of builtin health check route.
type healthConfig struct {
	enable                    bool
	healthChecker             hc.HealthChecker
	accessLogsEnabled         bool
	metricsForEndpointEnabled bool
	route                     string
}

// profilerConfig represents configuration of builtin profiler route.
type profilerConfig struct {
	enable            bool
	accessLogsEnabled bool
	route             string
}
