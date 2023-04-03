package middleware

import (
	"fmt"
	"log"
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestName(t *testing.T) {
	m := chi.NewMux()
	m.Get("/a/b/{id}", func(w http.ResponseWriter, r *http.Request) {
		ctx := chi.RouteContext(r.Context())

		fmt.Printf("%s\n", ctx.RoutePattern())
		fmt.Printf("%s\n", r.Proto)
		fmt.Printf("%s\n", r.RequestURI)
	})

	go func() { log.Fatal(http.ListenAndServe(":8080", m)) }()

	_, err := http.DefaultClient.Get("http://:8080/a/b/1")
	if err != nil {
		t.Fatal(err)
	}
}
