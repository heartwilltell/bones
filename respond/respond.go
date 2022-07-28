package respond

import (
	"encoding/json"
	"net/http"

	"github.com/heartwilltell/bones/ctxutil"
)

func Error(w http.ResponseWriter, r *http.Request, err error) {

	// Get log hook from the context to set an error which
	// will be logged along with access log line.
	if hook, ok := ctxutil.GetErrorLogHook(r.Context()); ok {
		hook(err)
	}
}

// JSON tries to encode v into json representation and write it to
// response writer.
func JSON(w http.ResponseWriter, r *http.Request, status int, v any) {
	coder := json.NewEncoder(w)
	coder.SetEscapeHTML(true)

	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := coder.Encode(v); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		// Get log hook from the context to set an error which
		// will be logged along with access log line.
		if hook, ok := ctxutil.GetErrorLogHook(r.Context()); ok {
			hook(err)
		}

		return
	}
}
