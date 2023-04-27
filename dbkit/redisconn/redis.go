package redisconn

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Conn wraps connection with the Redis.
type Conn struct{ *redis.Client }

// Option modifies the redis.Options.
type Option func(o *redis.Options)

func WithClientName(name string) Option {
	return func(o *redis.Options) { o.ClientName = name }
}

// New returns a pointer to a new instance of Conn struct.
func New(addr string, options ...Option) (*Conn, error) {
	connOptions := redis.Options{
		Addr: addr,
	}

	for _, option := range options {
		option(&connOptions)
	}

	client := redis.NewClient(&connOptions)

	return &Conn{Client: client}, nil
}

func (c *Conn) HealthCheck(ctx context.Context) error {
	if s := c.Client.Ping(ctx); s.Err() != nil {
		return fmt.Errorf("redis: healthcheck failed: %w", s.Err())
	}

	return nil
}
