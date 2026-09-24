package main

import "net/http"

type handler struct{}

func (h *handler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func Run() error { return http.ListenAndServe("localhost:8080", &handler{}) }

func main() { _ = Run() }
