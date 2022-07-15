package grpcserver

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/heartwilltell/hc"
	"github.com/heartwilltell/log"
	"google.golang.org/grpc"
)

type UnaryHandler interface {
	Mount(server *grpc.Server, interceptors ...grpc.UnaryServerInterceptor)
}

type StreamHandler interface {
	Mount(server *grpc.Server, interceptors ...grpc.StreamServerInterceptor)
}

// Server represents gRPC server.
type Server struct {
	log log.Logger
	hc  hc.HealthChecker

	listener net.Listener
	server   *grpc.Server
}

func New(addr string, options ...Option) (*Server, error) {
	listener, listenErr := net.Listen("tcp", addr)
	if listenErr != nil {
		return nil, fmt.Errorf("failed to create listener: %w", listenErr)
	}

	s := Server{
		log:      log.NewNopLog(),
		hc:       nil,
		listener: listener,
		server:   grpc.NewServer(),
	}

	// Apply all server options to the Server struct.
	for _, opt := range options {
		opt(&s)
	}

	return &s, nil
}

func (s *Server) Mount(handler UnaryHandler, interceptors ...grpc.UnaryServerInterceptor) {
	handler.Mount(s.server, interceptors...)
}

func (s *Server) MountStream(handler StreamHandler, interceptors ...grpc.StreamServerInterceptor) {
	handler.Mount(s.server, interceptors...)
}

// Serve listen to incoming connections and serves each request.
func (s *Server) Serve(ctx context.Context) error {
	// Handles the shutdown signal in the background.
	go s.handleShutdown(ctx)

	s.log.Info("Server started to listen on: %s", s.listener.Addr().String())

	if err := s.server.Serve(s.listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		return fmt.Errorf("server failed: %w", err)
	}

	s.log.Info("Bye!")

	return nil
}

// handleShutdown blocks until select statement receives a signal from
// ctx.Done, after that server's GracefulStop method will be called.
func (s *Server) handleShutdown(ctx context.Context) {
	<-ctx.Done()

	s.log.Info("Shutting down the server!")
	s.server.GracefulStop()
}
