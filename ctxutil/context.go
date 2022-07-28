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
	ServiceTx key = "ctx.service-tx"

	// RequestID  represents key for context by which
	// the request ID can be received from the context.
	RequestID key = "ctx.request-id"

	// CFRequestID represents key for context by which
	// the request ID can be received from the context.
	CFRequestID key = "ctx.cloudflare-request-id"
)

// SetErrorLogHook sets the error log hook to the context.
func SetErrorLogHook(ctx context.Context, hookedErr error) context.Context {
	return context.WithValue(ctx, ErrorLogHook, func(err error) { hookedErr = err })
}

// GetErrorLogHook gets the error log hook to the context.
func GetErrorLogHook(ctx context.Context) (func(err error), bool) {
	if hook, ok := ctx.Value(ErrorLogHook).(func(error)); ok {
		return hook, true
	}

	return nil, false
}

// SetServiceTx sets the service transaction to the context.
func SetServiceTx(ctx context.Context, tx db.Tx) context.Context {
	return context.WithValue(ctx, ServiceTx, tx)
}

// GetServiceTx gets the service transaction from the context.
func GetServiceTx(ctx context.Context) (db.Tx, bool) {
	if tx, ok := ctx.Value(ServiceTx).(db.Tx); ok {
		return tx, true
	}

	return nil, false
}

// SetRequestID sets the request ID to the context.
func SetRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, RequestID, id)
}

// GetRequestID gets the request ID from the context.
func GetRequestID(ctx context.Context) (string, bool) {
	if id, ok := ctx.Value(RequestID).(string); ok {
		return id, true
	}

	return "", false
}

// SetCFRequestID sets the Cloudflare request ID to the context.
func SetCFRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, CFRequestID, id)
}

// GetCFRequestID gets the Cloudflare request ID from the context.
func GetCFRequestID(ctx context.Context) (string, bool) {
	if id, ok := ctx.Value(CFRequestID).(string); ok {
		return id, true
	}

	return "", false
}
