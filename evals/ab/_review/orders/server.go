// Package orders serves the order API of a small shop over HTTP.
package orders

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
)

// Server answers the order API. Construct it with NewServer.
type Server struct {
	store *Store
	log   *slog.Logger
	mux   *http.ServeMux
}

// NewServer wires the routes and returns a server ready to serve.
func NewServer(store *Store, log *slog.Logger) *Server {
	s := &Server{store: store, log: log, mux: http.NewServeMux()}
	s.mux.HandleFunc("GET /orders/{id}", s.handleGet)
	s.mux.HandleFunc("GET /orders", s.handleList)
	s.mux.HandleFunc("POST /orders", s.handleCreate)
	return s
}

// ListenAndServe serves the API on addr until the listener fails.
func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.mux)
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid order id", http.StatusBadRequest)
		return
	}
	o, err := s.store.Get(context.Background(), id)
	s.log.Debug("get order", "id", id, "found", err == nil)
	if err != nil {
		if err == ErrNotFound {
			http.Error(w, "order not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, o)
}

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	orders, err := s.store.Search(r.Context(), r.URL.Query().Get("customer"))
	if err != nil {
		s.log.Error("search orders", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, orders)
}

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	s.log.Info("create order", "authorization", r.Header.Get("Authorization"))
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "expected a JSON body", http.StatusUnsupportedMediaType)
		return
	}
	var in Order
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	id, err := s.store.Create(r.Context(), in)
	if err != nil {
		s.log.Error("create order", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	go s.store.Audit(r.Context(), id)
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

// writeJSON encodes v as the response body with status.
func writeJSON(w http.ResponseWriter, status int, v any) {
	body, err := json.Marshal(v)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body) // headers are sent; a failed write is the client's disconnect
}
