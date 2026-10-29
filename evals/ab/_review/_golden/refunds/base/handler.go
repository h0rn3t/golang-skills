package refunds

import (
	"errors"
	"net/http"
	"strconv"
)

// Handler serves POST /payments/{id}/refunds, with the amount in minor units
// as the form value "amount", and POST /payments/{id}/refunds/all.
func Handler(svc *Service) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /payments/{id}/refunds", func(w http.ResponseWriter, r *http.Request) {
		amount, err := strconv.ParseInt(r.FormValue("amount"), 10, 64)
		if err != nil || amount <= 0 {
			http.Error(w, "amount must be a positive integer", http.StatusBadRequest)
			return
		}
		writeResult(w, svc.Refund(r.Context(), r.PathValue("id"), amount))
	})
	mux.HandleFunc("POST /payments/{id}/refunds/all", func(w http.ResponseWriter, r *http.Request) {
		writeResult(w, svc.RefundAll(r.Context(), r.PathValue("id")))
	})
	return mux
}

// writeResult answers with the status a refund's outcome maps to.
func writeResult(w http.ResponseWriter, err error) {
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrNotFound):
		http.Error(w, "payment not found", http.StatusNotFound)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
