package main

import "net/http"

type Handler struct{}

func (h *Handler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	_ = http.ListenAndServe("localhost:8080", &Handler{}) //nolint:gosec // G114: a probe, never started by the test
}
