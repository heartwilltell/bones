package bones

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/go-chi/chi/v5"
	"github.com/heartwilltell/bones/respond"
	"github.com/heartwilltell/hc"
	"github.com/heartwilltell/log"
)

// Middleware represents an HTTP server middleware.
type Middleware = func(next http.Handler) http.Handler

const (
	// readTimeout represents default read timeout for the http.Server.
	readTimeout = 2 * time.Second

	// readHeaderTimeout represents default read header timeout for http.Server.
	readHeaderTimeout = 5 * time.Second

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
	config serverConfig
}

// New return a new instance of Server struct.
func New(addr string, options ...Option) (*Server, error) {
	router := chi.NewRouter()

	s := Server{
		log: log.NewNopLog(),
		hc:  hc.NewNopChecker(),

		router: router,
		server: &http.Server{
			Addr:              addr,
			Handler:           router,
			ReadHeaderTimeout: readHeaderTimeout,
			ReadTimeout:       readTimeout,
			WriteTimeout:      writeTimeout,
			IdleTimeout:       idleTimeout,
		},

		config: defaultConfig(),
	}

	// Apply all server options to the Server struct.
	for _, opt := range options {
		opt(&s)
	}

	if err := s.applyConfig(); err != nil {
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
		return errInvalidAddress
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

func (s *Server) ServeTLS(ctx context.Context, cert, key string) error {
	if s.server.Addr == "" {
		return errInvalidAddress
	}

	// handle shutdown signal in the background
	go s.handleShutdown(ctx)

	s.log.Info("Server started to listen on: %s", s.server.Addr)

	if err := s.server.ListenAndServeTLS(cert, key); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server failed: %w", err)
	}

	s.log.Info("Bye!")

	return nil
}

func (s *Server) metrics(w http.ResponseWriter, _ *http.Request) {
	metrics.WritePrometheus(w, true)
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	if s.hc == nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := s.hc.Health(r.Context()); err != nil {
		respond.Error(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
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
