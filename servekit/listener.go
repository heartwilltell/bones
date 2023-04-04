package servekit

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
	"github.com/heartwilltell/bones/servekit/middleware"
	"github.com/heartwilltell/bones/servekit/respond"
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

// Middleware represents a http.Handler middleware.
type Middleware = func(next http.Handler) http.Handler

// ListenerHTTP all app logic in form of http.Handler interfaces
// along with routing logic and HTTP transport logic.
type ListenerHTTP struct {
	health hc.HealthChecker
	logger log.Logger
	router chi.Router
	server *http.Server
}

// New return a new instance of ListenerHTTP struct.
func New(addr string, options ...Option[*config]) (*ListenerHTTP, error) {
	router := chi.NewRouter()

	s := ListenerHTTP{
		logger: log.NewNopLog(),
		health: hc.NewNopChecker(),
		router: router,
		server: &http.Server{
			Addr:              addr,
			Handler:           router,
			ReadTimeout:       readTimeout,
			ReadHeaderTimeout: readHeaderTimeout,
			WriteTimeout:      writeTimeout,
			IdleTimeout:       idleTimeout,
		},
	}

	// To not keep a lot of unnecessary stuff on the ListenerHTTP struct instance,
	// all the options will be applied to config struct, which will be used
	// as the source of truth to apply the final configuration to the ListenerHTTP.
	if err := s.configure(options...); err != nil {
		return nil, fmt.Errorf("failed to apply cofiguration: %w", err)
	}

	return &s, nil
}

func (l *ListenerHTTP) Mount(route string, handler http.Handler, middlewares ...Middleware) {
	l.router.Route(route, func(r chi.Router) {
		r.Use(middlewares...)
		r.Mount("/", handler)
	})
}

// Serve listen to incoming connections and serves each request.
func (l *ListenerHTTP) Serve(ctx context.Context) error {
	if l.server.Addr == "" {
		return fmt.Errorf("invalid listener address: %s", l.server.Addr)
	}

	g, serveCtx := errgroup.WithContext(ctx)

	// handle shutdown signal in the background
	g.Go(func() error { return l.handleShutdown(serveCtx) })

	g.Go(func() error {
		l.logger.Info("ListenerHTTP started to listen on: %s", l.server.Addr)

		if err := l.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listener failed: %w", err)
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		l.logger.Error("Force exit! %s", err.Error())
		panic(err)
	}

	l.logger.Info("Bye!")

	return nil
}

func (l *ListenerHTTP) ServeTLS(ctx context.Context, cert, key string) error {
	if l.server.Addr == "" {
		return fmt.Errorf("invalid listener address: %s", l.server.Addr)
	}

	g, sctx := errgroup.WithContext(ctx)

	// handle shutdown signal in the background
	g.Go(func() error { return l.handleShutdown(sctx) })

	g.Go(func() error {
		l.logger.Info("ListenerHTTP started to listen on: %s", l.server.Addr)

		if err := l.server.ListenAndServeTLS(cert, key); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listener failed: %w", err)
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		l.logger.Error("Force exit! %s", err.Error())
		panic(err)
	}

	l.logger.Info("Bye!")

	return nil
}

func (l *ListenerHTTP) metrics(w http.ResponseWriter, _ *http.Request) {
	metrics.WritePrometheus(w, true)
}

func (l *ListenerHTTP) healthCheck(w http.ResponseWriter, r *http.Request) {
	if l.health == nil {
		respond.Status(w, r, http.StatusOK)
		return
	}

	if err := l.health.Health(r.Context()); err != nil {
		respond.Error(w, r, err)
		return
	}

	respond.Status(w, r, http.StatusOK)
}

// handleShutdown blocks until select statement receives a signal from
// ctx.Done, after that new context.WithTimeout will be created and passed to
// http.Server Shutdown method.
//
// If Shutdown method returns non nil error, program will panic immediately.
func (l *ListenerHTTP) handleShutdown(ctx context.Context) error {
	<-ctx.Done()

	l.logger.Info("Shutting down the listener!")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := l.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shutdown the listener gracefully: %w", err)
	}

	return nil
}

func (l *ListenerHTTP) configure(options ...Option[*config]) error {
	// Initialize the default settings.
	cfg := config{
		logger: log.NewNopLog(),

		readTimeout:       readTimeout,
		readHeaderTimeout: readHeaderTimeout,
		writeTimeout:      writeTimeout,
		idleTimeout:       idleTimeout,

		globalMiddlewares: make([]Middleware, 0),

		health: HealthEndpointConfig{
			enable:                    false,
			accessLogsEnabled:         false,
			metricsForEndpointEnabled: false,
			route:                     "/health",
			healthChecker:             hc.NewNopChecker(),
		},

		metrics: MetricsEndpointConfig{
			enable:                    false,
			accessLogsEnabled:         false,
			metricsForEndpointEnabled: false,
			route:                     "/metrics",
		},

		profiler: ProfilerEndpointConfig{
			enable:            false,
			accessLogsEnabled: false,
			route:             "/profiler",
		},
	}

	// Apply all server options to the config struct.
	for _, opt := range options {
		opt(&cfg)
	}

	// Apply logger settings.
	l.logger = cfg.logger

	// Apply health checker settings.
	l.health = cfg.health.healthChecker

	// Apply server timeouts.
	l.server.ReadTimeout = cfg.readTimeout
	l.server.ReadHeaderTimeout = cfg.readHeaderTimeout
	l.server.WriteTimeout = cfg.writeTimeout
	l.server.IdleTimeout = cfg.idleTimeout

	// Apply router-wide middleware.
	l.router.Use(cfg.globalMiddlewares...)

	if cfg.health.enable {
		if cfg.health.route == "" {
			return fmt.Errorf("invalid healt-check route: %s (should not be empty)", cfg.health.route)
		}

		if !strings.HasPrefix(cfg.health.route, "/") {
			return fmt.Errorf("invalid healt-check route: %s (route should start with '/' slash)", cfg.health.route)
		}

		l.router.Group(func(g chi.Router) {
			if cfg.health.accessLogsEnabled {
				g.Use(middleware.LoggingMiddleware(l.logger))
			}

			if cfg.health.metricsForEndpointEnabled {
				g.Use(middleware.MetricsMiddleware())
			}

			g.Get(cfg.health.route, l.healthCheck)
		})
	}

	if cfg.metrics.enable {
		if cfg.metrics.route == "" {
			return fmt.Errorf("invalid metrics route: %s (should not be empty)", cfg.metrics.route)
		}

		if !strings.HasPrefix(cfg.metrics.route, "/") {
			return fmt.Errorf("invalid metrics route: %s (route should start with '/' slash)", cfg.metrics.route)
		}

		l.router.Group(func(g chi.Router) {
			if cfg.metrics.accessLogsEnabled {
				g.Use(middleware.LoggingMiddleware(l.logger))
			}

			if cfg.metrics.metricsForEndpointEnabled {
				g.Use(middleware.MetricsMiddleware())
			}

			g.Get(cfg.metrics.route, l.metrics)
		})
	}

	if cfg.profiler.enable {
		l.router.Group(func(g chi.Router) {
			if cfg.profiler.accessLogsEnabled {
				g.Use(middleware.LoggingMiddleware(l.logger))
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

// config holds ListenerHTTP configuration.
type config struct {
	// logger represents a log instance which will used by ListenerHTTP.
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
	health HealthEndpointConfig

	// metrics holds configuration for metrics endpoint.
	metrics MetricsEndpointConfig

	// profiler holds configuration fot profiler endpoint.
	profiler ProfilerEndpointConfig
}
