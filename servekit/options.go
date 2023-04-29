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

// WithHealthCheck turns on the health check endpoint.
// Receives the following option to configure the endpoint:
// - SetHealthChecker - to change the HealthChecker implementation.
// - HealthCheckRoute - to set the endpoint route.
// - HealthCheckAccessLog - to enable access log for endpoint.
// - HealthCheckMetricsForEndpoint - to enable metrics collection for endpoint.
func WithHealthCheck(options ...Option[*HealthEndpointConfig]) Option[*config] {
	return func(c *config) {
		c.health.enable = true

		for _, opt := range options {
			opt(&c.health)
		}
	}
}

// SetHealthChecker represents an optional function for WithHealthCheck function.
// If passed to the WithHealthCheck, will set the config.health.healthChecker.
func SetHealthChecker(hc hc.HealthChecker) Option[*HealthEndpointConfig] {
	return func(c *HealthEndpointConfig) {
		// To not shoot in the leg. There are already a nop checker.
		if hc == nil {
			return
		}

		c.HealthChecker = hc
	}
}

// HealthCheckRoute represents an optional function for WithHealthCheck function.
// If passed to the WithHealthCheck, will set the config.health.route.
func HealthCheckRoute(route string) Option[*HealthEndpointConfig] {
	return func(c *HealthEndpointConfig) { c.Route = route }
}

// HealthCheckAccessLog represents an optional function for WithHealthCheck function.
// If passed to the WithHealthCheck, will set the config.health.accessLogsEnabled to true.
func HealthCheckAccessLog(enable bool) Option[*HealthEndpointConfig] {
	return func(c *HealthEndpointConfig) { c.AccessLogsEnabled = enable }
}

// HealthCheckMetricsForEndpoint represents an optional function for WithHealthCheck function.
// If passed to the WithHealthCheck, will set the config.health.metricsForEndpointEnabled to true.
func HealthCheckMetricsForEndpoint(enable bool) Option[*HealthEndpointConfig] {
	return func(c *HealthEndpointConfig) { c.MetricsForEndpointEnabled = enable }
}

// WithMetrics turns on the metrics endpoint.
// Receives the following option to configure the endpoint:
// - MetricsRoute - to set the endpoint route.
// - MetricsAccessLog - to enable access log for endpoint.
// - MetricsMetricsForEndpoint - to enable metrics collection for endpoint.
func WithMetrics(options ...Option[*MetricsEndpointConfig]) Option[*config] {
	return func(c *config) {
		c.metrics.enable = true

		for _, opt := range options {
			opt(&c.metrics)
		}
	}
}

// MetricsRoute represents an optional function for WithMetrics function.
// If passed to the WithMetrics, will set the config.health.route.
func MetricsRoute(route string) Option[*MetricsEndpointConfig] {
	return func(c *MetricsEndpointConfig) { c.Route = route }
}

// MetricsAccessLog represents an optional function for WithMetrics function.
// If passed to the WithMetrics, will set the config.health.accessLogsEnabled to true.
func MetricsAccessLog(enable bool) Option[*MetricsEndpointConfig] {
	return func(c *MetricsEndpointConfig) { c.AccessLogsEnabled = enable }
}

// MetricsMetricsForEndpoint represents an optional function for WithMetrics function.
// If passed to the WithMetrics, will set the config.health.metricsForEndpointEnabled to true.
func MetricsMetricsForEndpoint(enable bool) Option[*MetricsEndpointConfig] {
	return func(c *MetricsEndpointConfig) { c.MetricsForEndpointEnabled = enable }
}

// WithProfiler turns on the profiler endpoint.
func WithProfiler(cfg ProfilerEndpointConfig) Option[*config] {
	return func(c *config) {
		c.profiler.enable = true
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
