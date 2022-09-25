package errkit

// This is compiling time check for interface implementation.
var _ error = (Error)("")

const (
	// ErrUnknown indicates unknown error.
	ErrUnknown Error = "unknown error"

	// ErrInvalidArgument indicates that client specified an invalid argument.
	ErrInvalidArgument Error = "invalid argument"

	// ErrNotFound indicates that requested entity was not found.
	ErrNotFound Error = "now found"

	// ErrAlreadyExists indicates an attempt to create an entity
	// which is failed because such entity already exists.
	ErrAlreadyExists Error = "already exists"

	// ErrUnauthenticated indicates the request does not have valid
	// authentication credentials to perform the operation.
	ErrUnauthenticated Error = "authentication failed"

	// ErrUnauthorized indicates the caller does not have permission to
	// execute the specified operation. It must not be used if the caller
	// cannot be identified (use ErrUnauthenticated instead for those errors).
	ErrUnauthorized Error = "permission denied"

	// ErrUnavailable indicates that the service is currently unavailable.
	// This kind of error is retryable. Caller should retry with a backoff.
	ErrUnavailable Error = "temporarily unavailable"

	// ErrConnFailed shows that connection to resource failed.
	ErrConnFailed Error = "connection failed"

	// ErrTxFailed shows that database transaction failed.
	ErrTxFailed Error = "transaction failed"

	// ErrTxBegin shows that database transaction failed to begin.
	ErrTxBegin Error = "failed to begin transaction"

	// ErrTxCommit shows that commit of database transaction failed.
	ErrTxCommit Error = "failed to commit transaction"

	// ErrTxRollback shows that rollback of database transaction failed.
	ErrTxRollback Error = "failed to rollback transaction"

	// ErrMigrationFailed shows that database migration failed.
	ErrMigrationFailed Error = "failed to migrate database schema"
)

// Error type represents package level errors.
type Error string

func (e Error) Error() string { return string(e) }
