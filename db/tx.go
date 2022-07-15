package db

import "context"

// Doer represents interface that can wrapp function calls in
// closure that represents service level transaction.
type Doer interface {
	// Do executes service transaction.
	Do(ctx context.Context, f func(ctx context.Context) error) error
}

// Tx represents service level transaction.
type Tx interface {
	// Begin starts a pseudo nested transaction.
	Begin(ctx context.Context) (Tx, error)
	// Commit commits the transaction.
	Commit(ctx context.Context) error
	// Rollback rolls back the transaction.
	Rollback(ctx context.Context) error
}
