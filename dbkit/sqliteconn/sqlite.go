package sqliteconn

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/heartwilltell/log"

	"github.com/heartwilltell/hc"
	_ "github.com/mattn/go-sqlite3"
)

// Compilation time check that Conn implements
// the hc.HealthChecker.
var _ hc.HealthChecker = (*Conn)(nil)

const (
	schemaVersionTableName   = "schema_version"
	migrationRollbackDivider = "---- create above / drop below ----"
)

// Conn represents connection to SQLite.
// Wraps the pointer to the standard sql.DB struct.
type Conn struct{ *sql.DB }

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

// Health implements hc.HealthChecker interface.
func (c *Conn) Health(ctx context.Context) error {
	if err := c.PingContext(ctx); err != nil {
		return fmt.Errorf("sqlite: health check failed: %w", err)
	}

	return nil
}

// MigrationOption
type MigrationOption func(m *Migrator)

// MigrateWithBackup
func MigrateWithBackup(filePath string) MigrationOption {
	return func(m *Migrator) {
		m.enableBackup = true
		m.databasePath = filePath
	}
}

// MigrateWithLogs
func MigrateWithLogs(logger log.Logger) MigrationOption {
	return func(m *Migrator) { m.log = logger }
}

// Migration represents a database migration.
type Migration struct {
	Number   uint
	Up, Down string
}

// Migrator loads and applies database migrations to SQLite.
type Migrator struct {
	conn *Conn
	log  log.Logger

	// migrationsPath holds a path to folder with migrations.
	migrationsPath string

	// databasePath holds a path to the SQLite database file.
	databasePath string

	// versionTableName holds the name of the schema version table.
	versionTableName string

	// enableBackup shows whether backup of the database is enabled.
	enableBackup bool

	migrationsLoaded sync.Once

	// migrations holds a set of migrations.
	migrations []Migration
}

// Migrations returns a pointer to a new instance of Migrator.
func Migrations(conn *Conn, migrations fs.FS, options ...MigrationOption) (*Migrator, error) {
	m := Migrator{
		conn:             conn,
		log:              log.NewNopLog(),
		migrationsPath:   ".",
		databasePath:     ".",
		versionTableName: schemaVersionTableName,
		enableBackup:     false,
	}

	for _, option := range options {
		option(&m)
	}

	entries, readErr := fs.ReadDir(migrations, m.migrationsPath)
	if readErr != nil {
		return nil, fmt.Errorf("sqlite: failed to load migrations: %w", readErr)
	}

	m.migrations = make([]Migration, 0, len(entries))

	for _, entry := range entries {
		info, infoErr := entry.Info()
		if infoErr != nil {
			return nil, fmt.Errorf("sqlite: failed to get file info: %w", infoErr)
		}

		if strings.HasSuffix(info.Name(), ".sql") {
			m.log.Info("Loading '%s'", info.Name())

			migration, readFileErr := fs.ReadFile(migrations, info.Name())
			if readFileErr != nil {
				return nil, fmt.Errorf("sqlite: failed to load migration file '%s': %w", info.Name(), readFileErr)
			}

			parts := strings.SplitN(string(migration), migrationRollbackDivider, 2)

			m.migrations = append(m.migrations, Migration{
				Up:   strings.TrimSpace(parts[0]),
				Down: strings.TrimSpace(parts[1]),
			})
		}

		m.log.Info("Loading finished")
	}

	return &m, nil
}

func (m *Migrator) Migrate(ctx context.Context) error {
	m.log.Info("Starting migration process")

	if m.enableBackup {
		m.log.Info("Creating database backup")

		if err := m.backup(); err != nil {
			return fmt.Errorf("failed to create database backup: %w", err)
		}

		m.log.Info("Database backup has been created")
	}

	m.log.Info("Migrating")

	if err := m.migrate(ctx); err != nil {
		m.log.Error("Migration failed: %s", err.Error())

		m.log.Info("Removing broken database file '%s'", m.databasePath)

		if removeErr := os.Remove(m.databasePath); removeErr != nil {
			return errors.Join(err, fmt.Errorf("failed to delete broken database file: %w", removeErr))
		}

		dir, file := path.Split(m.databasePath)

		if renameErr := os.Rename(path.Join(dir, fmt.Sprintf("backup-%s", file)), m.databasePath); renameErr != nil {
			return errors.Join(err, fmt.Errorf("failed to rename backup file: %w", renameErr))
		}

		return err
	}

	m.log.Info("Migration complete")
	m.log.Info("Removing backup")

	dir, file := path.Split(m.databasePath)

	if err := os.Remove(path.Join(dir, fmt.Sprintf("backup-%s", file))); err != nil {
		return errors.Join(err, fmt.Errorf("failed to rename enableBackup file: %w", err))
	}

	return nil
}

func (m *Migrator) migrate(ctx context.Context) (mErr error) {
	if err := m.conn.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	tx, txErr := m.conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if txErr != nil {
		return fmt.Errorf("failed to begin transaction: %w", txErr)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			mErr = errors.Join(mErr, fmt.Errorf("failed to rollback transaction: %w", err))
		}
	}()

	for _, migration := range m.migrations {
		if _, err := tx.ExecContext(ctx, migration.Up); err != nil {
			return fmt.Errorf("failed to apply migration: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (m *Migrator) backup() (bErr error) {
	stat, statErr := os.Stat(m.databasePath)
	if statErr != nil {
		return fmt.Errorf("failed to get file information: %w", statErr)
	}

	if stat.IsDir() {
		return fmt.Errorf("given path is a directory istead of file: %s", m.databasePath)
	}

	srcDir, _ := path.Split(m.databasePath)

	src, srcOpenErr := os.Open(m.databasePath)
	if srcOpenErr != nil {
		return fmt.Errorf("failed to open database file: %w", srcOpenErr)
	}
	defer func() {
		if err := src.Close(); err != nil {
			bErr = errors.Join(bErr, fmt.Errorf("failed to close database source file: %w", err))
		}
	}()

	dst, dstOpenErr := os.Create(path.Join(srcDir, fmt.Sprintf("backup-%s", stat.Name())))
	if dstOpenErr != nil {
		return fmt.Errorf("failed to create database backup file: %w", dstOpenErr)
	}

	defer func() {
		if err := dst.Close(); err != nil {
			bErr = errors.Join(bErr, fmt.Errorf("failed to close database backup file: %w", err))
		}
	}()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	return nil
}
