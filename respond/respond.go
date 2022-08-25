package respond

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/heartwilltell/bones/bctx"
	"github.com/heartwilltell/bones/berr"
)

// Error tries to map err to bones.Error and based on result
// writes standard http error to response writer.
func Error(w http.ResponseWriter, r *http.Request, err error) {
	// Get log hook from the context to set an error which
	// will be logged along with access log line.
	if hook := bctx.Get[func(error)](r.Context(), bctx.ErrorLogHook); hook != nil {
		hook(err)
	}

	if errors.Is(err, berr.ErrAlreadyExists) {
		http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		return
	}

	if errors.Is(err, berr.ErrNotFound) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	if errors.Is(err, berr.ErrUnauthenticated) {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	if errors.Is(err, berr.ErrUnauthorized) {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	if errors.Is(err, berr.ErrInvalidArgument) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if errors.Is(err, berr.ErrUnavailable) {
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}

	if errors.Is(err, berr.ErrUnknown) {
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
		if hook := bctx.Get[func(error)](r.Context(), bctx.ErrorLogHook); hook != nil {
			hook(err)
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// TEXT tries to write v to response writer.
func TEXT(w http.ResponseWriter, r *http.Request, status int, v []byte) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)

	// return if v is empty
	if len(v) == 0 {
		return
	}

	if _, err := w.Write(v); err != nil {
		// Get log hook from the context to set an error which
		// will be logged along with access log line.
		if hook := bctx.Get[func(error)](r.Context(), bctx.ErrorLogHook); hook != nil {
			hook(err)
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
