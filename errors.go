package bones

// This is compiling time check for interface implementation.
var _ error = (Error)("")

// ErrInvalidAddress is returned when Server address is empty.
const ErrInvalidAddress Error = "invalid server address"

// Error type represents package level errors.
type Error string

func (e Error) Error() string { return string(e) }
