package sqliteconn

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/heartwilltell/hc"
	_ "github.com/mattn/go-sqlite3"
)

// Compilation time check that Conn implements the hc.HealthChecker.
var _ hc.HealthChecker = (*Conn)(nil)

// JournalMode represents SQLite journaling mode.
// https://www.sqlite.org/pragma.html#pragma_journal_mode
//
// Note that the JournalMode for an InMemory database is either Memory or Off
// and can not be changed to a different value.
//
// An attempt to change the JournalMode of an InMemory database to any setting
// other than Memory or Off is ignored.
//
// Note also that the JournalMode cannot be changed while a transaction is active.
type JournalMode byte

const (
	// Delete journaling mode is the normal behavior. In the Delete mode, the rollback
	// journal is deleted at the conclusion of each transaction. Indeed, the delete
	// operation is the action that causes the transaction to commit.
	// See the document titled Atomic Commit In SQLite for additional detail:
	// https://www.sqlite.org/atomiccommit.html
	Delete JournalMode = iota

	// Truncate journaling mode commits transactions by truncating the rollback journal to
	// zero-length instead of deleting it. On many systems, truncating a file is much faster
	// than deleting the file since the containing directory does not need to be changed.
	Truncate

	// Persist journaling mode prevents the rollback journal from being deleted at the end of
	// each transaction. Instead, the header of the journal is overwritten with zeros.
	// This will prevent other database connections from rolling the journal back.
	// The Persist journaling mode is useful as an optimization on platforms where deleting or
	// truncating a file is much more expensive than overwriting the first block of a file with zeros.
	Persist

	// Memory journaling mode stores the rollback journal in volatile RAM.
	// This saves disk I/O but at the expense of database safety and integrity.
	// If the application using SQLite crashes in the middle of a transaction when
	// the Memory journaling mode is set, then the database file will very likely go corrupt.
	Memory

	// WAL journaling mode uses a write-ahead log instead of a rollback journal to implement transactions.
	// The WAL journaling mode is persistent; after being set it stays in effect across multiple database
	// connections and after closing and reopening the database.
	// A database in WAL journaling mode can only be accessed by SQLite version 3.7.0 (2010-07-21) or later.
	WAL

	// Off journaling mode disables the rollback journal completely.
	// No rollback journal is ever created and hence there is never a rollback journal to delete.
	// The Off journaling mode disables the atomic commit and rollback capabilities of SQLite.
	// The ROLLBACK command no longer works; it behaves in an undefined way.
	// Applications must avoid using the ROLLBACK command when the journal mode is Off.
	// If the application crashes in the middle of a transaction when the Off journaling mode is set,
	// then the database file will very likely go corrupt.
	Off
)

func (m JournalMode) String() string {
	modes := map[JournalMode]string{
		Delete:   "DELETE",
		Truncate: "TRUNCATE",
		Persist:  "PERSIST",
		Memory:   "MEMORY",
		WAL:      "WAL",
		Off:      "OFF",
	}

	return modes[m]
}

// AccessMode represents SQLite access mode.
// https://www.sqlite.org/c3ref/open.html
type AccessMode byte

func (m AccessMode) String() string {
	modes := map[AccessMode]string{
		ReadWriteCreate: "rwc",
		ReadOnly:        "ro",
		ReadWrite:       "rw",
		InMemory:        "memory",
	}

	return modes[m]
}

const (
	// ReadWriteCreate is a mode in which the database is opened for reading and writing,
	// and is created if it does not already exist.
	ReadWriteCreate AccessMode = iota

	// ReadOnly is a mode in which the database is opened in read-only mode.
	// If the database does not already exist, an error is returned.
	ReadOnly

	// ReadWrite is a mode in which database is opened for reading and writing if possible,
	// or reading only if the file is write protected by the operating system.
	// In either case the database must already exist, otherwise an error is returned.
	// For historical reasons, if opening in read-write mode fails due to OS-level permissions,
	// an attempt is made to open it in read-only mode.
	ReadWrite

	// InMemory is a mode in which database will be opened as an in-memory database.
	// The database is named by the "filename" argument for the purposes of cache-sharing,
	// if shared cache mode is enabled, but the "filename" is otherwise ignored.
	InMemory
)

// Option represents an optional functions which configures the ConnOptions.
type Option func(o *ConnOptions)

// WithAccessMode enables SQLite to use picked access mode.
func WithAccessMode(mode AccessMode) Option {
	return func(o *ConnOptions) { o.accessMode = mode }
}

// WithJournalMode sets SQLite to use picked journal mode.
func WithJournalMode(mode JournalMode) Option {
	return func(o *ConnOptions) { o.journalingMode = mode }
}

// ConnOptions holds a set of options which will be used to
// establish the connection with SQLite.
type ConnOptions struct {
	// path holds an absolute path to the database file.
	path string

	// journalingMode represents SQLite journaling mode.
	// https://www.sqlite.org/pragma.html#pragma_journal_mode
	journalingMode JournalMode

	// accessMode represents SQLite access mode.
	// https://www.sqlite.org/c3ref/open.html
	accessMode AccessMode
}

func (o *ConnOptions) connString() (string, error) {
	params := make([]string, 0, 2)

	switch o.accessMode {
	case ReadWriteCreate, InMemory, ReadOnly, ReadWrite:
		params = append(params, "mode="+o.accessMode.String())
	default:
		return "", errors.New("unsupported access mode")
	}

	switch o.journalingMode {
	case Delete, Truncate, Persist, WAL, Memory, Off:
		params = append(params, "_journal="+o.journalingMode.String())
	default:
		return "", errors.New("unsupported journal mode")
	}

	var b strings.Builder

	b.WriteString("file:" + o.path)

	if len(params) > 0 {
		b.WriteString("?")
	}

	for i, p := range params {
		b.WriteString(p)

		if i != len(p)-1 {
			b.WriteString("&")
		}
	}

	return b.String(), nil
}

// Conn represents connection to SQLite.
// Wraps the pointer to the standard sql.DB struct.
type Conn struct{ *sql.DB }

// New returns a pointer to a new instance of Conn with a pointer to sql.DB struct.
func New(path string, options ...Option) (*Conn, error) {
	absPath, absPathErr := filepath.Abs(path)
	if absPathErr != nil {
		return nil, fmt.Errorf("sqlite: determite absolute path to the database: %w", absPathErr)
	}

	connOptions := ConnOptions{
		path: absPath,
	}

	for _, option := range options {
		option(&connOptions)
	}

	connString, connStringErr := connOptions.connString()
	if connStringErr != nil {
		return nil, fmt.Errorf("sqlite: %w", connStringErr)
	}

	db, openErr := sql.Open("sqlite3", connString)
	if openErr != nil {
		return nil, fmt.Errorf("sqlite: open database: %w", openErr)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("sqlite: connect to database: %w", err)
	}

	return &Conn{DB: db}, nil
}

// Health implements hc.HealthChecker interface.
func (c *Conn) Health(ctx context.Context) error {
	if err := c.PingContext(ctx); err != nil {
		return fmt.Errorf("sqlite: health check: %w", err)
	}

	return nil
}
