package ctxkit

import (
	"context"
)

const (
	// LogErrHook represents a key for context by which
	// the error log hook can be received from the context.
	logErrHook key = "ctx.log-error-hook"

	// RequestID represents a key for context by which
	// the request ID can be received from the context.
	requestID key = "ctx.request-id"
)

// key represents a context key with custom type.
type key string

// SetLogErrHook sets the hook function to the context.
func SetLogErrHook(ctx context.Context, hook func(err error)) context.Context {
	return context.WithValue(ctx, logErrHook, hook)
}

// GetLogErrHook gets the hook function from the context which sets the error to server error log.
// If searched values is absent in context, then nil wil be returned.
func GetLogErrHook(ctx context.Context) func(error) {
	if hook, ok := ctx.Value(logErrHook).(func(error)); ok {
		return hook
	}

	return nil
}

// SetRequestID sets the request ID to the context.
func SetRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestID, id)
}

// GetRequestID gets the request ID from the context.
// If searched values is absent in context, then empty string wil be returned.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestID).(string); ok {
		return id
	}

	return ""
}
