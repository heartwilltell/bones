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

func WithCredentials(username, password string) Option {
	return func(o *redis.Options) {
		o.Username = username
		o.Password = password
	}
}

func WithCredentialsProvider(provider func() (string, string)) Option {
	return func(o *redis.Options) { o.CredentialsProvider = provider }
}

// New returns a pointer to a new instance of Conn struct.
func New(addr string, options ...Option) (*Conn, error) {
	connOptions := redis.Options{
		Network:               "",
		Addr:                  addr,
		ClientName:            "",
		Dialer:                nil,
		OnConnect:             nil,
		DB:                    0,
		MaxRetries:            0,
		MinRetryBackoff:       0,
		MaxRetryBackoff:       0,
		DialTimeout:           0,
		ReadTimeout:           0,
		WriteTimeout:          0,
		ContextTimeoutEnabled: false,
		PoolFIFO:              false,
		PoolSize:              0,
		PoolTimeout:           0,
		MinIdleConns:          0,
		MaxIdleConns:          0,
		ConnMaxIdleTime:       0,
		ConnMaxLifetime:       0,
		TLSConfig:             nil,
		Limiter:               nil,
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
