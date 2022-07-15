package db

// This is compiling time check for interface implementation.
var _ error = (Error)("")

// Enumeration of database related errors.
const (
	// ErrNotFound shows that no records has been found.
	ErrNotFound Error = "record not found"

	// ErrAlreadyExist shows that record already exist in database.
	ErrAlreadyExist Error = "record already exist"

	// ErrTxFailed shows that database transaction failed.
	ErrTxFailed Error = "transaction failed"

	// ErrTxBegin shows that begin of database transaction is failed.
	ErrTxBegin Error = "failed to begin transaction"

	// ErrTxCommit shows that commit of database transaction failed.
	ErrTxCommit Error = "failed to commit transaction"

	// ErrTxRollback shows that rollback of database transaction failed.
	ErrTxRollback Error = "failed to rollback transaction"

	// ErrConnFailed shows that database connection failed.
	ErrConnFailed Error = "failed to connect to database"

	// ErrMigrationFailed shows that database migration failed.
	ErrMigrationFailed Error = "failed to migrate database schema"
)

type Error string

func (e Error) Error() string { return string(e) }
