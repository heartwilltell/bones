package ctxutil

import (
	"context"

	"github.com/heartwilltell/bones/db"
)

// key represents a context key with custom type.
type key string

const (
	// errorLogHook represents key for context by which
	// the error log hook can be received from the context.
	errorLogHook key = "ctx.error-log-hook"

	// serviceTx represents key for context by which
	// the service transaction can be received from the context.
	serviceTx key = "ctx.service-tx"

	// requestID  represents key for context by which
	// the request ID can be received from the context.
	requestID key = "ctx.request-id"
)

// SetErrorLogHook sets the error log hook to the context.
func SetErrorLogHook(ctx context.Context, hook func(err error)) context.Context {
	return context.WithValue(ctx, errorLogHook, hook)
}

// GetErrorLogHook gets the error log hook to the context.
func GetErrorLogHook(ctx context.Context) (func(err error), bool) {
	if hook, ok := ctx.Value(errorLogHook).(func(error)); ok {
		return hook, true
	}

	return nil, false
}

// SetServiceTx sets the service transaction to the context.
func SetServiceTx(ctx context.Context, tx db.Tx) context.Context {
	return context.WithValue(ctx, serviceTx, tx)
}

// GetServiceTx gets the service transaction from the context.
func GetServiceTx(ctx context.Context) (db.Tx, bool) {
	if tx, ok := ctx.Value(serviceTx).(db.Tx); ok {
		return tx, true
	}

	return nil, false
}

// SetRequestID sets the request ID to the context.
func SetRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestID, id)
}

// GetRequestID gets the request ID from the context.
func GetRequestID(ctx context.Context) (string, bool) {
	if id, ok := ctx.Value(requestID).(string); ok {
		return id, true
	}

	return "", false
}
