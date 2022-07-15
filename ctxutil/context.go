package ctxutil

import (
	"context"

	"github.com/heartwilltell/bones/db"
)

// key represents a context key with custom type.
type key string

const (
	// ErrorLogHook represents key for context by which
	// the error log hook can be received from the context.
	ErrorLogHook key = "ctx.error-log-hook"

	// ServiceTx represents key for context by which
	// the service transaction can be received from the context.
	ServiceTx key = "ctx.service-transaction"
)

// SetErrorLogHook sets the error log hook to context.
func SetErrorLogHook(ctx context.Context, hookedErr error) context.Context {
	return context.WithValue(ctx, ErrorLogHook, func(err error) { hookedErr = err })
}

// GetErrorLogHook gets the error log hook to context.
func GetErrorLogHook(ctx context.Context) (func(err error), bool) {
	if hook, ok := ctx.Value(ErrorLogHook).(func(error)); ok {
		return hook, true
	}

	return nil, false
}

// SetServiceTx
func SetServiceTx(ctx context.Context, tx db.Tx) context.Context {
	return context.WithValue(ctx, ServiceTx, tx)
}

// GetServiceTx
func GetServiceTx(ctx context.Context) (db.Tx, bool) {
	if tx, ok := ctx.Value(ServiceTx).(db.Tx); ok {
		return tx, true
	}

	return nil, false
}
