package bctx

import (
	"context"
)

const (
	// LogErrHook represents a key for context by which
	// the error log hook can be received from the context.
	LogErrHook key = "ctx.log-error-hook"

	// RequestID represents a key for context by which
	// the request ID can be received from the context.
	RequestID key = "ctx.request-id"
)

type LogErrHookFunc func(error)

type (
	// key represents a context key with custom type.
	key string

	// value represents generic constraint for context value type.
	value interface{ LogErrHookFunc | string }
)

// Get gets value from context by context key.
// If searched values is absent in context, then
// zero value of specified type wil be returned.
func Get[T value](ctx context.Context, k key) T {
	if v, ok := ctx.Value(k).(T); ok {
		return v
	}

	return zero[T]()
}

// Set sets given value to context by specified key.
func Set[T value](ctx context.Context, k key, v T) context.Context {
	return context.WithValue(ctx, k, v)
}

// zero returns zero value for specified type.
func zero[T value]() T { var z T; return z }
