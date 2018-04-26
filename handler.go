package grpcsql

import (
	"net/http"
	"testing"
)

type Handler struct {
	t       *testing.T
	handler http.Handler
}

func NewHandler(t *testing.T, handler http.Handler) http.Handler {
	return &Handler{
		t:       t,
		handler: handler,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler.ServeHTTP(w, r)
}
