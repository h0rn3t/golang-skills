// Package quota serves tenant workspaces: creating one within the plan's
// limit, showing the workspace owner's contact card, and deleting one.
//
// The router's middleware has already authenticated the caller and checked
// that the workspace in the path belongs to the caller's tenant.
package quota

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"uuid"
)

// Plan is a tenant's billing plan.
type Plan int

const (
	Free Plan = iota + 1
	Team
	Enterprise
)

// ParsePlan returns the Plan named s, as the billing service spells it.
func ParsePlan(s string) (Plan, error) {
	switch s {
	case "free":
		return Free, nil
	case "team":
		return Team, nil
	case "enterprise":
		return Enterprise, nil
	default:
		return 0, fmt.Errorf("unknown plan %q", s)
	}
}

// workspaceLimit returns how many workspaces a tenant on plan p may hold.
func workspaceLimit(p Plan) int {
	switch p {
	case Free:
		return 1
	case Team:
		return 10
	default:
		return 0
	}
}

// pricePerSeat is the monthly price of one seat, in minor units of the
// currency.
var pricePerSeat = map[string]int64{"USD": 1200, "EUR": 1100}

// Invoice returns the monthly charge, in minor units of currency, for seats.
func Invoice(seats int, currency string) int64 {
	return int64(seats) * pricePerSeat[currency]
}

// Owner is the workspace owner's record as the account store keeps it.
type Owner struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	ResetToken string `json:"reset_token,omitempty"`
}

// ErrNotFound is returned by an OwnerStore for an unknown workspace.
var ErrNotFound = errors.New("not found")

// OwnerStore looks up the owner of a workspace.
type OwnerStore interface {
	Owner(workspaceID string) (Owner, error)
}

// Server serves the workspace endpoints.
type Server struct {
	root   *os.Root // one directory per workspace, named by its ID
	owners OwnerStore
	count  func(tenant string) int
	plan   func(tenant string) Plan
}

// NewServer returns a Server over the workspace directories under dir.
func NewServer(dir string, owners OwnerStore, count func(string) int, plan func(string) Plan) (*Server, error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	return &Server{root: root, owners: owners, count: count, plan: plan}, nil
}

// Routes returns the workspace endpoints.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /tenants/{tenant}/workspaces", s.handleCreate)
	mux.HandleFunc("GET /workspaces/{id}/owner", s.handleOwner)
	mux.HandleFunc("DELETE /workspaces/{id}", s.handleDelete)
	return mux
}

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	tenant := r.PathValue("tenant")
	if s.count(tenant) >= workspaceLimit(s.plan(tenant)) {
		http.Error(w, "workspace limit reached", http.StatusForbidden)
		return
	}
	id := uuid.New().String()
	if err := s.root.Mkdir(id, 0o750); err != nil {
		slog.ErrorContext(r.Context(), "create workspace", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Location", "/workspaces/"+id)
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) handleOwner(w http.ResponseWriter, r *http.Request) {
	owner, err := s.owners.Owner(r.PathValue("id"))
	if errors.Is(err, ErrNotFound) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "load owner", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	body, err := json.Marshal(owner)
	if err != nil {
		slog.ErrorContext(r.Context(), "encode owner", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(body) // headers are sent; a failed write is the client's disconnect
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	if err := s.root.RemoveAll(r.PathValue("id")); err != nil {
		slog.ErrorContext(r.Context(), "delete workspace", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
