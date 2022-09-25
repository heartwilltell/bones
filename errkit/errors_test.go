package errkit

import "testing"

func TestError_Error(t *testing.T) {
	type tcase struct {
		err  Error
		want string
	}

	tests := map[string]tcase{
		"ErrMigrationFailed": {err: ErrMigrationFailed, want: "failed to migrate database schema"},
		"ErrAlreadyExist":    {err: ErrAlreadyExist, want: "record already exist"},
		"ErrTxFailed":        {err: ErrTxFailed, want: "transaction failed"},
		"ErrTxBegin":         {err: ErrTxBegin, want: "failed to begin transaction"},
		"ErrTxCommit":        {err: ErrTxCommit, want: "failed to commit transaction"},
		"ErrTxRollback":      {err: ErrTxRollback, want: "failed to rollback transaction"},
		"ErrNotFound":        {err: ErrNotFound, want: "record not found"},
		"Custom":             {err: Error("test error"), want: "test error"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tc.err.Error(); got != tc.want {
				t.Errorf("Error() = %v, want %v", got, tc.want)
			}
		})
	}
}
