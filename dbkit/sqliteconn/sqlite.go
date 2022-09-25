package sqliteconn

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/heartwilltell/hc"
	_ "github.com/mattn/go-sqlite3"
)

// Compilation time check that Conn implements
// the hc.HealthChecker.
var _ hc.HealthChecker = (*Conn)(nil)

// Conn represents connection to SQLite.
// Wraps the pointer to the standard sql.DB struct.
type Conn struct{ *sql.DB }

// Health implements hc.HealthChecker interface.
func (c *Conn) Health(ctx context.Context) error {
	if err := c.PingContext(ctx); err != nil {
		return fmt.Errorf("sqlite: health check failed: %w", err)
	}
	return nil
}

// New returns a pointer to a new instance of Conn with a pointer to sql.DB struct.
func New(path string) (*Conn, error) {
	db, openErr := sql.Open("sqlite3", path)
	if openErr != nil {
		return nil, fmt.Errorf("sqlite: failed to open database: %w", openErr)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("sqlite: failed to connect to database: %w", err)
	}

	return &Conn{db}, nil
}
