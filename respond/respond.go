package respond

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/heartwilltell/bones"
	"github.com/heartwilltell/bones/ctxutil"
)

// Error tries to map err to bones.Error and based on result
// writes standard http error to response writer.
func Error(w http.ResponseWriter, r *http.Request, err error) {
	// Get log hook from the context to set an error which
	// will be logged along with access log line.
	if hook, ok := ctxutil.GetErrorLogHook(r.Context()); ok {
		hook(err)
	}

	if errors.Is(err, bones.ErrAlreadyExists) {
		http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		return
	}

	if errors.Is(err, bones.ErrNotFound) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	if errors.Is(err, bones.ErrUnauthenticated) {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	if errors.Is(err, bones.ErrUnauthorized) {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	if errors.Is(err, bones.ErrInvalidArgument) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if errors.Is(err, bones.ErrUnavailable) {
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}

	if errors.Is(err, bones.ErrUnknown) {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// JSON tries to encode v into json representation and write it to
// response writer.
func JSON(w http.ResponseWriter, r *http.Request, status int, v any) {
	coder := json.NewEncoder(w)
	coder.SetEscapeHTML(true)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := coder.Encode(v); err != nil {
		// Get log hook from the context to set an error which
		// will be logged along with access log line.
		if hook, ok := ctxutil.GetErrorLogHook(r.Context()); ok {
			hook(err)
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

// TEXT tries to write v to response writer.
func TEXT(w http.ResponseWriter, r *http.Request, status int, v string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)

	// return in v is empty
	if v == "" {
		return
	}

	if _, err := w.Write([]byte(v)); err != nil {
		// Get log hook from the context to set an error which
		// will be logged along with access log line.
		if hook, ok := ctxutil.GetErrorLogHook(r.Context()); ok {
			hook(err)
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
