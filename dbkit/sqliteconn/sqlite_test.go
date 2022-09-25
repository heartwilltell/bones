package sqliteconn

import (
	"context"
	"os"
	"path"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestNew(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		expName := "test.db"

		dbpath := path.Join(t.TempDir(), expName)

		conn, err := New(dbpath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		defer func() {
			if err := os.Remove(dbpath); err != nil {
				t.Log("failed to delete test file:", dbpath)
			}
		}()

		if err := conn.Health(context.Background()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if err := conn.Close(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		f, openErr := os.Open(dbpath)
		if openErr != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		stat, statErr := f.Stat()
		if statErr != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if stat.Name() != expName {
			t.Errorf("expected name := %s, got := %s", expName, stat.Name())
		}

		if stat.Mode() != os.FileMode(0644) {
			t.Errorf("expected mode := %s, got := %s", os.FileMode(0644), stat.Mode())
		}
	})
}
