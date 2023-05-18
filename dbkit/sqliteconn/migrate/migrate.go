package migrate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/heartwilltell/bones/dbkit/sqliteconn"
	"github.com/heartwilltell/log"
)

const (
	schemaVersionTableName   = "schema_version"
	migrationRollbackDivider = "---- create above / drop below ----"
)

// MigrationOption implements functional options pattern for Migrator type.
// Represents a function which receive a pointer to the Migrator struct
// and changes it default values to the given ones.
type MigrationOption func(m *Migrator)

// MigrateWithBackup sets whether backup of the database enabled.
func MigrateWithBackup(databasePath string) MigrationOption {
	return func(m *Migrator) {
		m.backupEnabled = true
		m.databasePath = databasePath
	}
}

// MigrateWithLogger sets the Migrator logger.
func MigrateWithLogger(logger log.Logger) MigrationOption {
	return func(m *Migrator) { m.log = logger }
}

func MigrateWithMigrationPath(migrationsPath string) MigrationOption {
	return func(m *Migrator) { m.migrationsPath = migrationsPath }
}

// Migration represents a database migration.
type Migration struct {
	Number   uint
	Up, Down string
}

// Migrator loads and applies database migrations to SQLite.
type Migrator struct {
	log log.Logger

	// migrations holds a filesystem with migration files.
	migrations fs.FS

	// migrationsPath holds path to the folder with migrations.
	migrationsPath string

	// databasePath holds an absolute path to the SQLite database file.
	databasePath string

	// versionTableName holds the name of the schema version table.
	versionTableName string

	// backupEnabled shows whether backup of the database is enabled.
	backupEnabled   bool
	backupKeepAfter bool
}

// Migrations returns a pointer to a new instance of Migrator.
func Migrations(migrations fs.FS, options ...MigrationOption) *Migrator {
	m := Migrator{
		log:              log.NewNopLog(),
		migrations:       migrations,
		migrationsPath:   ".",
		databasePath:     ".",
		versionTableName: schemaVersionTableName,
		backupEnabled:    false,
		backupKeepAfter:  false,
	}

	for _, option := range options {
		option(&m)
	}

	return &m
}

func (m *Migrator) Migrate(ctx context.Context, conn *sqliteconn.Conn) error {
	migrations, loadErr := m.loadMigrations()
	if loadErr != nil {
		return loadErr
	}

	m.log.Info("Starting migration process")

	if m.backupEnabled {
		m.log.Info("Creating database backup")

		if err := m.backup(); err != nil {
			return fmt.Errorf("failed to create database backup: %w", err)
		}

		m.log.Info("Database backup has been created")
	}

	m.log.Info("Migrating")

	if err := m.migrate(ctx, conn, migrations); err != nil {
		m.log.Error("Migration failed: %s", err.Error())

		if m.backupEnabled {
			m.log.Info("Removing broken database file '%s'", m.databasePath)

			if removeErr := os.Remove(m.databasePath); removeErr != nil {
				return errors.Join(err, fmt.Errorf("failed to delete broken database file: %w", removeErr))
			}

			dir, file := path.Split(m.databasePath)

			if renameErr := os.Rename(path.Join(dir, fmt.Sprintf("backup-%s", file)), m.databasePath); renameErr != nil {
				return errors.Join(err, fmt.Errorf("failed to rename backup file: %w", renameErr))
			}
		}

		return err
	}

	m.log.Info("Migration complete")

	if m.backupEnabled && !m.backupKeepAfter {
		m.log.Info("Removing backup")

		dir, file := path.Split(m.databasePath)

		if err := os.Remove(path.Join(dir, fmt.Sprintf("backup-%s", file))); err != nil {
			return errors.Join(err, fmt.Errorf("failed to rename backupEnabled file: %w", err))
		}
	}

	return nil
}

func (m *Migrator) migrate(ctx context.Context, conn *sqliteconn.Conn, migrations []Migration) (mErr error) {
	if err := conn.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	tx, txErr := conn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if txErr != nil {
		return fmt.Errorf("failed to begin transaction: %w", txErr)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			mErr = errors.Join(mErr, fmt.Errorf("failed to rollback transaction: %w", err))
		}
	}()

	for _, migration := range migrations {
		if _, err := tx.ExecContext(ctx, migration.Up); err != nil {
			return fmt.Errorf("failed to apply migration: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (m *Migrator) loadMigrations() ([]Migration, error) {
	entries, readErr := fs.ReadDir(m.migrations, m.migrationsPath)
	if readErr != nil {
		return nil, fmt.Errorf("sqlite: failed to load migrations: %w", readErr)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	migrations := make([]Migration, 0, len(entries))

	for i, entry := range entries {
		info, infoErr := entry.Info()
		if infoErr != nil {
			return nil, fmt.Errorf("sqlite: failed to get file info: %w", infoErr)
		}

		if strings.HasSuffix(info.Name(), ".sql") {
			m.log.Info("Loading '%s'", info.Name())

			migration, readFileErr := fs.ReadFile(m.migrations, info.Name())
			if readFileErr != nil {
				return nil, fmt.Errorf("sqlite: failed to load migration file '%s': %w", info.Name(), readFileErr)
			}

			parts := strings.SplitN(string(migration), migrationRollbackDivider, 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("sqlite: invalid migration file '%s': divider '%s' is absent", info.Name(), migrationRollbackDivider)
			}

			migrations = append(migrations, Migration{
				Number: func() uint {
					if i == 0 {
						return uint(i + 1)
					}

					return uint(i)
				}(),
				Up:   strings.TrimSpace(parts[0]),
				Down: strings.TrimSpace(parts[1]),
			})
		}

		m.log.Info("Loading finished")
	}

	return migrations, nil
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
