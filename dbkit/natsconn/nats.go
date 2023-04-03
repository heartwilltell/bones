package natsconn

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
)

// Conn wraps connection with NATS.
type Conn struct{ *nats.Conn }

func (c *Conn) HealthCheck(_ context.Context) error {
	if c.Status() == nats.CONNECTED {
		return nil
	}

	return fmt.Errorf("nats: connection failed: %s", c.Status())
}

// New returns a pointer to a new instance of Conn.
func New(addr string) (*Conn, error) {
	conn, err := nats.Connect(addr)
	if err != nil {
		return nil, fmt.Errorf("nats: failed to connect to NATS: %w", err)
	}

	return &Conn{Conn: conn}, nil
}
