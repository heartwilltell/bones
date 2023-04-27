package servekit

import (
	"time"

	"github.com/heartwilltell/hc"
	"github.com/heartwilltell/log"
)

// Option implements functional options pattern for ListenerHTTP type.
// Represents a function which receive a pointer to the config struct
// and changes it default values to the given ones.
//
// See the ListenerHTTP.configure function to understand the configuration behaviour.
// Option functions should only be passed to ListenerHTTP constructor function New.
type Option[T any] func(o T)

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

// WithGlobalMiddlewares sets given middlewares as router-wide middlewares.
// Means that they will be applied to each server endpoint.
func WithGlobalMiddlewares(m ...Middleware) Option[*config] {
	return func(c *config) { c.globalMiddlewares = append(c.globalMiddlewares, m...) }
}

// WithLogger sets the server logger.
func WithLogger(l log.Logger) Option[*config] {
	return func(c *config) {
		if l != nil {
			c.logger = l
		}
	}
}

// WithMetrics turns on the metrics endpoint.
func WithMetrics(cfg MetricsEndpointConfig) Option[*config] {
	return func(c *config) {
		c.metrics.enable = true
		c.metrics.AccessLogsEnabled = cfg.AccessLogsEnabled
		c.metrics.MetricsForEndpointEnabled = cfg.MetricsForEndpointEnabled

		if cfg.Route != "" {
			c.metrics.Route = cfg.Route
		}
	}
}

// WithHealthCheck turns on the health check endpoint.
func WithHealthCheck(cfg HealthEndpointConfig) Option[*config] {
	return func(c *config) {
		c.health.enable = cfg.enable
		c.health.AccessLogsEnabled = cfg.AccessLogsEnabled
		c.health.MetricsForEndpointEnabled = cfg.MetricsForEndpointEnabled

		if cfg.Route != "" {
			c.metrics.Route = cfg.Route
		}

		if cfg.HealthChecker != nil {
			c.health.HealthChecker = cfg.HealthChecker
		}
	}
}

// WithProfiler turns on the profiler endpoint.
func WithProfiler(cfg ProfilerEndpointConfig) Option[*config] {
	return func(c *config) {
		c.profiler.enable = cfg.enable
		c.profiler.accessLogsEnabled = cfg.accessLogsEnabled

		if cfg.route != "" {
			c.profiler.route = cfg.route
		}
	}
}

// MetricsEndpointConfig represents configuration of builtin metrics route.
type MetricsEndpointConfig struct {
	Route                     string
	AccessLogsEnabled         bool
	MetricsForEndpointEnabled bool

	enable bool
}

// HealthEndpointConfig represents configuration of builtin health check route.
type HealthEndpointConfig struct {
	Route                     string
	HealthChecker             hc.HealthChecker
	AccessLogsEnabled         bool
	MetricsForEndpointEnabled bool

	enable bool
}

// ProfilerEndpointConfig represents configuration of builtin profiler route.
type ProfilerEndpointConfig struct {
	route             string
	accessLogsEnabled bool

	enable bool
}
