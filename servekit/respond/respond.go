package respond

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"

	"github.com/heartwilltell/bones/ctxkit"
	"github.com/heartwilltell/bones/errkit"
)

var (
	// errResponderInit is a guard to set the ErrorResponder only once to avoid
	// accidentally reassigned errResponder which is used by default.
	errResponderInit sync.Once

	// errResponder represents the default implementation of ErrorResponder func.
	errResponder ErrorResponder = func(w http.ResponseWriter, err error) {
		if errors.Is(err, errkit.ErrAlreadyExists) {
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return
		}

		if errors.Is(err, errkit.ErrNotFound) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		if errors.Is(err, errkit.ErrUnauthenticated) {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		if errors.Is(err, errkit.ErrUnauthorized) {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		if errors.Is(err, errkit.ErrInvalidArgument) {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if errors.Is(err, errkit.ErrUnavailable) {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
)

// WithErrorResponder sets the given responder as errResponder.
func WithErrorResponder(responder ErrorResponder) {
	errResponderInit.Do(func() { errResponder = responder })
}

// ErrorResponder represents a function which should be called to respond with an error on HTTP call.
type ErrorResponder func(w http.ResponseWriter, err error)

// Status writes an HTTP status to the w http.ResponseWriter.
func Status(w http.ResponseWriter, _ *http.Request, status int) {
	w.WriteHeader(status)
}

// Error tries to map err to errkit.Error and based on result
// writes standard http error to response writer.
func Error(w http.ResponseWriter, r *http.Request, err error) {
	// Get log hook from the context to set an error which
	// will be logged along with access log line.
	if hook := ctxkit.GetLogErrHook(r.Context()); hook != nil {
		hook(err)
	}

	// Call the default error responder.
	errResponder(w, err)
}

// JSON tries to encode v into json representation and write it to response writer.
func JSON(w http.ResponseWriter, r *http.Request, status int, v any) {
	coder := json.NewEncoder(w)
	coder.SetEscapeHTML(true)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := coder.Encode(v); err != nil {
		// Get log hook from the context to set an error which
		// will be logged along with access log line.
		if hook := ctxkit.GetLogErrHook(r.Context()); hook != nil {
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
		if hook := ctxkit.GetLogErrHook(r.Context()); hook != nil {
			hook(err)
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
