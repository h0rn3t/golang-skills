// Package handlers is the order module's HTTP surface: decode, structural
// checks, identity, a call into the use case, and the error-to-status map.
package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"example.com/shop/internal/order"
	"example.com/shop/internal/order/models"
	"example.com/shop/internal/order/services"
)

// Register mounts the order routes.
func Register(mux *http.ServeMux, svc *services.Service) {
	mux.HandleFunc("POST /orders", placeOrder(svc))
	mux.HandleFunc("GET /orders/{id}/summary", orderSummary(svc))
}

type placeRequest struct {
	ID       string        `json:"id"`
	Customer string        `json:"customer"`
	Lines    []models.Line `json:"lines"`
}

func placeOrder(svc *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req placeRequest
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		actor := r.Header.Get("X-Actor") // identity is carried here; permission is the service's decision
		err := svc.Place(r.Context(), actor, models.Order{ID: req.ID, Customer: req.Customer, Lines: req.Lines})
		switch {
		case errors.Is(err, services.ErrForbidden):
			http.Error(w, "forbidden", http.StatusForbidden)
		case errors.Is(err, models.ErrEmptyOrder):
			http.Error(w, "an order needs at least one line", http.StatusUnprocessableEntity)
		case err != nil:
			http.Error(w, "internal error", http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusCreated)
		}
	}
}

func orderSummary(svc *services.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s, err := svc.Summary(r.Context(), r.PathValue("id"))
		switch {
		case errors.Is(err, order.ErrNotFound):
			http.Error(w, "unknown order", http.StatusNotFound)
		case err != nil:
			http.Error(w, "internal error", http.StatusInternalServerError)
		default:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(s)
		}
	}
}
