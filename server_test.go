package bones

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/heartwilltell/hc"
	"github.com/heartwilltell/log"
	"github.com/maxatome/go-testdeep/td"
)

func TestNew(t *testing.T) {
	type tcase struct {
		addr    string
		options []Option[*config]
		want    *Server
		wantErr error
	}

	tests := map[string]tcase{
		"OK": {
			addr:    ":8080",
			options: nil,
			want: &Server{
				health: hc.NewNopChecker(),
				logger: log.NewNopLog(),
				router: chi.NewRouter(),
				server: &http.Server{
					Addr:    ":8080",
					Handler: chi.NewRouter(),
				},
			},
			wantErr: nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := New(tc.addr, tc.options...)
			td.Cmp(t, got, tc.want)
			td.Cmp(t, err, tc.wantErr)
		})
	}
}
