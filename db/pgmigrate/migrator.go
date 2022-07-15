package pgmigrate

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"time"

	"github.com/heartwilltell/bones/db/pgconn"

	"github.com/jackc/tern/migrate"
)

var _ migrate.MigratorFS = (*migrationsFS)(nil)

// New returns a pointer to a new instance of Migrator.
func New(conn *pgconn.Conn, migrations fs.FS, path string) (*Migrator, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	mConn, mConnErr := conn.Acquire(ctx)
	if mConnErr != nil {
		return nil, fmt.Errorf("failed to acquire connection for migrator: %w", mConnErr)
	}

	m, mErr := migrate.NewMigratorEx(ctx, mConn.Conn(), "migration", &migrate.MigratorOptions{
		MigratorFS: &migrationsFS{migrations},
	})
	if mErr != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", mErr)
	}

	if err := m.LoadMigrations(path); err != nil {
		return nil, fmt.Errorf("failed to load migrations: %w", err)
	}

	return &Migrator{conn: conn, m: m}, nil
}

// Migrator holds logic of how to load and apply database
// migrations to SQLite.
type Migrator struct {
	conn *pgconn.Conn
	m    *migrate.Migrator
}

func (m *Migrator) Close() error {
	m.conn.Close()
	return nil
}

// Migrate performs migration of database schema.
func (m *Migrator) Migrate(ctx context.Context) error { return m.m.Migrate(ctx) }

// migrationsFS represents an array of migrations.
type migrationsFS struct{ fs.FS }

func (m *migrationsFS) ReadFile(filename string) ([]byte, error) { return fs.ReadFile(m.FS, filename) }

func (m *migrationsFS) Glob(pattern string) ([]string, error) { return fs.Glob(m.FS, pattern) }

func (m *migrationsFS) ReadDir(dirname string) ([]os.FileInfo, error) {
	entries, readErr := fs.ReadDir(m.FS, dirname)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dirname, readErr)
	}

	infos := make([]os.FileInfo, 0, len(entries))

	for _, entry := range entries {
		info, infoErr := entry.Info()
		if infoErr != nil {
			return nil, fmt.Errorf("failed to get information about %s entry: %w", entry.Name(), infoErr)
		}

		infos = append(infos, info)
	}

	return infos, nil
}

// Error implements builtin error interface and represents package level errors
// related to work of Migrator type.
type Error string

func (e Error) Error() string { return string(e) }
