package bones

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// listener represents an HTTP listener.
type listener struct {
	router chi.Router
	server *http.Server
}

func newListener(addr string) listener {
	router := chi.NewRouter()

	l := listener{
		router: router,
		server: &http.Server{
			Addr:              addr,
			Handler:           router,
			ReadHeaderTimeout: readHeaderTimeout,
			ReadTimeout:       readTimeout,
			WriteTimeout:      writeTimeout,
			IdleTimeout:       idleTimeout,
		},
	}

	return l
}
