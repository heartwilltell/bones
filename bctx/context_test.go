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
		td.Cmp(t, got, want)
	})

	t.Run("Get LogErrHook", func(t *testing.T) {
		want := LogErrHookFunc(func(err error) {})
		ctx := context.WithValue(context.Background(), LogErrHook, want)

		got := Get[LogErrHookFunc](ctx, LogErrHook)
		td.Cmp(t, got, want, td.Ptr())
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

	t.Run("Set LogErrHook", func(t *testing.T) {
		want := LogErrHookFunc(func(err error) {})

		ctx := Set[LogErrHookFunc](context.Background(), LogErrHook, want)

		got, ok := ctx.Value(LogErrHook).(LogErrHookFunc)
		if !ok {
			t.Errorf("expected velue of string type but got %s", reflect.TypeOf(got).Kind().String())
		}

		if reflect.TypeOf(got).Kind() != reflect.Func {
			t.Errorf("expected velue of string type but got %s", reflect.TypeOf(got).Kind().String())
		}

		if !reflect.DeepEqual(want, got) {
			t.Errorf("expected := %p, got := %p", want, got)
		}
	})
}

func Test_zero(t *testing.T) {
	t.Run("Zero func(error)", func(t *testing.T) {
		var expected func(error)

		got := zero[LogErrHookFunc]()

		td.Cmp(t, got, expected)
	})

	t.Run("Zero string", func(t *testing.T) {
		var expected string

		got := zero[string]()

		td.Cmp(t, got, expected)
	})
}
