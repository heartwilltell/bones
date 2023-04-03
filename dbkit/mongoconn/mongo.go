package mongoconn

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Conn wraps the connection to the MongoDB.
type Conn struct{ *mongo.Client }

// New returns a pointer to a new instance of Conn struct.
func New(ctx context.Context, addr string, clientOptions ...*options.ClientOptions) (*Conn, error) {
	connOptions := options.MergeClientOptions(clientOptions...).
		ApplyURI(addr)

	client, err := mongo.Connect(ctx, connOptions)
	if err != nil {
		return nil, fmt.Errorf("mongo: connection failed: %w", err)
	}

	return &Conn{Client: client}, nil
}

func (c *Conn) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Disconnect(ctx); err != nil {
		return err
	}

	return nil
}

// HealthCheck implements the health.Checker interface for MongoDB connection.
func (c *Conn) HealthCheck(ctx context.Context) error {
	prefs, err := readpref.New(readpref.PrimaryPreferredMode)
	if err != nil {
		return fmt.Errorf("mongo: failed to create read preference")
	}

	if err := c.Client.Ping(ctx, prefs); err != nil {
		return fmt.Errorf("mongo: failed to ping database: %w", err)
	}

	return nil
}
