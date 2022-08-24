package errutil

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
	ErrAlreadyExists Error = "already exist"

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
)

// Error type represents package level errors.
type Error string

func (e Error) Error() string { return string(e) }
