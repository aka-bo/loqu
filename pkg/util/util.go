package util

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-logr/glogr"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
)

type contextKey string

const (
	KeyRequestID                   = "x-request-id"
	contextKeyRequestID contextKey = KeyRequestID
)

// EnsureRequestID retrieves the value of the X-Request-Id header. If not found it will generate a new one and set the header accordingly.
// Returns the resulting id.
func EnsureRequestID(r *http.Request) string {
	id := r.Header.Get(KeyRequestID)
	if len(id) == 0 {
		id = NewRequestID()
		r.Header.Set(KeyRequestID, id)
	}

	return id
}

// NewRequestID generates a new request id
func NewRequestID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

// RequestContext generates a new context that includes a request id
func RequestContext(r *http.Request) context.Context {
	id := EnsureRequestID(r)
	return context.WithValue(r.Context(), contextKeyRequestID, id)
}

// GetRequestID retrieves the request id from the request context
func GetRequestID(r *http.Request) string {
	id, _ := r.Context().Value(contextKeyRequestID).(string)
	return id
}

// WithID returns a logr.Logger with a name and request id added to the log context
func WithID(name string, r *http.Request) logr.Logger {
	return glogr.New().WithName(name).WithValues("RequestID", GetRequestID(r))
}
