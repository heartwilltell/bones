package ctxkit

import (
	"context"
	"testing"

	"github.com/maxatome/go-testdeep/td"
)

func TestGetLogErrHook(t *testing.T) {
	want := func(error) {}
	ctx := context.WithValue(context.Background(), logErrHook, want)
	got := GetLogErrHook(ctx)
	td.Cmp(t, got, td.Shallow(want))
}

func TestGetRequestID(t *testing.T) {
	want := "testid"
	ctx := context.WithValue(context.Background(), requestID, want)
	got := GetRequestID(ctx)
	td.Cmp(t, got, want)
}

func TestSetLogErrHook(t *testing.T) {
	want := func(error) {}
	ctx := SetLogErrHook(context.Background(), want)
	got, _ := ctx.Value(logErrHook).(func(error))
	td.Cmp(t, got, td.Shallow(want))
}

func TestSetRequestID(t *testing.T) {
	want := "testid"
	ctx := SetRequestID(context.Background(), want)
	got := ctx.Value(requestID)
	td.Cmp(t, got, want)
}
