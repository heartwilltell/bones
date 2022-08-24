package ctxutil

import (
	"context"
)

// key represents a context key with custom type.
type key string

// value represents generic constraint for context value type.
type value interface {
	func(error) | string
}

const (
	// ErrorLogHook represents a key for context by which
	// the error log hook can be received from the context.
	ErrorLogHook key = "ctx.error-log-hook"

	// RequestID represents a key for context by which
	// the request ID can be received from the context.
	RequestID key = "ctx.request-id"
)

func Get[T value](ctx context.Context, k key) T {
	if v, ok := ctx.Value(k).(T); ok {
		return v
	}

	return nil
}

func Set[T value](ctx context.Context, k key, v T) context.Context {
	return context.WithValue(ctx, k, v)
}
