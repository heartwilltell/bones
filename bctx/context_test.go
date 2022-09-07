package bctx

import (
	"context"
	"reflect"
	"testing"

	"github.com/maxatome/go-testdeep/td"
)

func TestGet(t *testing.T) {
	t.Run("Get RequestID", func(t *testing.T) {
		want := "testid"
		ctx := context.WithValue(context.Background(), RequestID, want)

		got := Get[string](ctx, RequestID)
		if got != want {
			t.Errorf("expected := %s, got := %s", want, got)
		}
	})
}

func TestSet(t *testing.T) {
	t.Run("Set RequestID", func(t *testing.T) {
		want := "testid"

		ctx := Set[string](context.Background(), RequestID, want)

		got, ok := ctx.Value(RequestID).(string)
		if !ok {
			t.Errorf("expected velue of string type but got %s", reflect.TypeOf(got).Kind().String())
		}

		if reflect.TypeOf(got).Kind() != reflect.String {
			t.Errorf("expected velue of string type but got %s", reflect.TypeOf(got).Kind().String())
		}

		if got != want {
			t.Errorf("expected := %s got := %s", want, got)
		}
	})
}

func Test_zero(t *testing.T) {
	t.Run("Zero func(error)", func(t *testing.T) {
		var expected func(error)

		got := zero[func(error)]()

		td.Cmp(t, got, expected)
	})

	t.Run("Zero string", func(t *testing.T) {
		var expected string

		got := zero[string]()

		td.Cmp(t, got, expected)
	})
}
