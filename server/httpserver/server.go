package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/heartwilltell/bones/server"

	"github.com/VictoriaMetrics/metrics"
	"github.com/go-chi/chi/v5"
	"github.com/heartwilltell/hc"
	"github.com/heartwilltell/log"
)

const (
	// readTimeout represents default read timeout for the http.Server.
	readTimeout = 2 * time.Second

	// writeTimeout represents default write timeout for the http.Server.
	writeTimeout = 3 * time.Second

	// idleTimeout represents default idle timeout for the http.Server.
	idleTimeout = 10 * time.Second

	// shutdownTimeout represents server default shutdown timeout.
	shutdownTimeout = 5 * time.Second
)

// Server holds all app logic in form of http.Handler interfaces
// along with routing logic and HTTP transport logic.
type Server struct {
	hc     hc.HealthChecker
	log    log.Logger
	router chi.Router
	server *http.Server

	settings struct {
		healthCheckDisabled               bool
		healthCheckAccessLogsEnabled      bool
		healthCheckEndpointMetricsEnabled bool
		metricsDisabled                   bool
		metricsAccessLogsEnabled          bool
		metricsEndpointMetricsEnabled     bool
	}
}

// New return a new instance of Server struct.
func New(addr string, options ...Option) (*Server, error) {
	router := chi.NewRouter()

	s := Server{
		log: log.NewNopLog(),
		hc:  nil,

		router: router,
		server: &http.Server{
			Addr:         addr,
			Handler:      router,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
		},
	}

	// Apply all server options to the Server struct.
	for _, opt := range options {
		opt(&s)
	}

	s.setup()

	return &s, nil
}

func (s *Server) Mount(route string, handler http.Handler, middlewares ...Middleware) {
	s.router.Route(route, func(r chi.Router) {
		r.Use(middlewares...)
		r.Mount("/", handler)
	})
}

func (s *Server) RedeclareMetricsEndpoint(handler http.Handler, middlewares ...Middleware) {
	s.router.Route("/metrics", func(r chi.Router) {
		r.Use(middlewares...)
		r.Mount("/", handler)
	})
}

func (s *Server) RedeclareHealthEndpoint(handler http.Handler, middlewares ...Middleware) {
	s.router.Route("/health", func(r chi.Router) {
		r.Use(middlewares...)
		r.Mount("/", handler)
	})
}

func (s *Server) metrics(w http.ResponseWriter, _ *http.Request) {
	metrics.WritePrometheus(w, true)
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	if s.hc == nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
	}

	if err := s.hc.Health(r.Context()); err != nil {
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

// Serve listen to incoming connections and serves each request.
func (s *Server) Serve(ctx context.Context) error {
	if s.server.Addr == "" {
		return server.ErrInvalidAddress
	}

	// handle shutdown signal in the background
	go s.handleShutdown(ctx)

	s.log.Info("Server started to listen on: %s", s.server.Addr)

	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server failed: %w", err)
	}

	s.log.Info("Bye!")

	return nil
}

// handleShutdown blocks until select statement receives a signal from
// ctx.Done, after that new context.WithTimeout will be created and passed to
// http.Server Shutdown method.
//
// If Shutdown method returns non nil error, program will panic immediately.
func (s *Server) handleShutdown(ctx context.Context) {
	<-ctx.Done()

	s.log.Info("Shutting down the server!")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	s.server.SetKeepAlivesEnabled(false)

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		s.log.Error("Force exit! Failed to shutdown the server gracefully: %s", err.Error())
		panic(err)
	}
}

func (s *Server) setup() {
	s.setupHealthEndpoint()
	s.setupMetricsEndpoint()
}

func (s *Server) setupHealthEndpoint() {
	if !s.settings.healthCheckDisabled {
		s.router.Group(func(r chi.Router) {
			if s.settings.healthCheckAccessLogsEnabled {
				r.Use(LoggingMiddleware(s.log))
			}

			if s.settings.healthCheckEndpointMetricsEnabled {
				r.Use(MetricsMiddleware())
			}

			r.Head("/health", s.healthCheck)
			r.Get("/health", s.healthCheck)
		})
	}
}

func (s *Server) setupMetricsEndpoint() {
	if !s.settings.metricsDisabled {
		s.router.Group(func(r chi.Router) {
			if s.settings.metricsAccessLogsEnabled {
				r.Use(LoggingMiddleware(s.log))
			}

			if s.settings.metricsEndpointMetricsEnabled {
				r.Use(MetricsMiddleware())
			}

			r.Get("/metrics", s.metrics)
		})
	}
}

// Middleware represents an HTTP server middleware.
type Middleware = func(next http.Handler) http.Handler
