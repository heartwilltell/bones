package bones

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
)

func Test(t *testing.T) {
	r := chi.NewRouter()
	r.Get("/book/{id}/pages", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.RequestURI)
		id := chi.URLParam(r, "id")
		fmt.Println(id)
	})

	s := http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		if err := s.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	rest, err := http.DefaultClient.Get("http://localhost:8080/book//pages")
	fmt.Println(rest.StatusCode, err)
}
