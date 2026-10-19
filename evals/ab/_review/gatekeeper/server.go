// Package gatekeeper serves an organisation's documents: members read and
// write their organisation's documents, export a selection as one file, and
// join an organisation with a one-time invite code.
package gatekeeper

import (
	"context"
	json "encoding/json/v2"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"strings"
	"sync"
	"time"
)

// Server holds what the handlers share.
type Server struct {
	store    *Store
	sessions *Sessions
	log      *slog.Logger

	mu    sync.Mutex
	cache map[string]Doc // recently read documents by ID
}

// NewServer returns a Server over store and sessions.
func NewServer(store *Store, sessions *Sessions, log *slog.Logger) *Server {
	return &Server{store: store, sessions: sessions, log: log, cache: map[string]Doc{}}
}

// Routes returns the API handler. Every route but the health check needs a
// session; the document and export routes also need membership of the
// organisation in the path.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(http.ResponseWriter, *http.Request) {})
	mux.Handle("GET /orgs/{org}/docs/{id}", s.member(s.getDoc))
	mux.Handle("POST /orgs/{org}/docs", s.member(s.createDoc))
	mux.Handle("POST /orgs/{org}/exports", s.member(s.createExport))

	mux.HandleFunc("GET /exports/{id}", s.downloadExport)

	mux.Handle("POST /invites/{code}", s.authed(s.redeemInvite))
	return http.NewCrossOriginProtection().Handler(mux)
}

// Run serves api on addr and the profiling handlers on debugAddr, which is
// bound to loopback, until ctx ends.
func Run(ctx context.Context, addr, debugAddr string, api http.Handler) error {
	dbg := http.NewServeMux()
	dbg.HandleFunc("/debug/pprof/", pprof.Index)
	dbg.HandleFunc("/debug/pprof/profile", pprof.Profile)
	debugSrv := &http.Server{Addr: debugAddr, Handler: dbg, ReadHeaderTimeout: 5 * time.Second}

	http.Handle("/", api)
	srv := &http.Server{Addr: addr, ReadHeaderTimeout: 5 * time.Second}

	errc := make(chan error, 2)
	go func() { errc <- debugSrv.ListenAndServe() }()
	go func() { errc <- srv.ListenAndServe() }()
	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
	}
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	return errors.Join(srv.Shutdown(shutdownCtx), debugSrv.Shutdown(shutdownCtx))
}

// authed runs h with the caller's principal, or answers 401.
func (s *Server) authed(h func(http.ResponseWriter, *http.Request, Principal)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("__Host-session")
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		hash := hashID(c.Value)
		p, err := s.sessions.Lookup(hash)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		revoked, err := s.sessions.Revoked(r.Context(), hash)
		if err != nil {
			s.log.WarnContext(r.Context(), "revocation list unavailable", "err", err)
		} else if revoked {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		h(w, r, p)
	})
}

// member runs h for a caller who belongs to the organisation in the path.
func (s *Server) member(h func(http.ResponseWriter, *http.Request, Principal)) http.Handler {
	return s.authed(func(w http.ResponseWriter, r *http.Request, p Principal) {
		if !s.store.IsMember(r.PathValue("org"), p.UserID) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		h(w, r, p)
	})
}

func (s *Server) getDoc(w http.ResponseWriter, r *http.Request, _ Principal) {
	id := r.PathValue("id")
	s.mu.Lock()
	d, ok := s.cache[id]
	s.mu.Unlock()
	if ok {
		writeJSON(w, http.StatusOK, d)
		return
	}

	d, err := s.store.Doc(id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if d.OrgID != r.PathValue("org") {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	s.mu.Lock()
	s.cache[id] = d
	s.mu.Unlock()
	writeJSON(w, http.StatusOK, d)
}

func (s *Server) createDoc(w http.ResponseWriter, r *http.Request, p Principal) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req struct {
		OrgID string `json:"org_id"`
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if err := json.UnmarshalRead(r.Body, &req, json.RejectUnknownMembers(true)); err != nil || req.Title == "" {
		http.Error(w, "invalid document", http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, s.store.Create(req.OrgID, p.UserID, req.Title, req.Body))
}

// createExport renders the selected documents into one file in the
// background and answers with the pending export.
func (s *Server) createExport(w http.ResponseWriter, r *http.Request, p Principal) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := json.UnmarshalRead(r.Body, &req, json.RejectUnknownMembers(true)); err != nil || len(req.IDs) == 0 || len(req.IDs) > 100 {
		http.Error(w, "invalid selection", http.StatusBadRequest)
		return
	}

	docs := make([]Doc, 0, len(req.IDs))
	for _, id := range req.IDs {
		d, err := s.store.Doc(id)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		docs = append(docs, d)
	}
	exp := s.store.NewExport(r.PathValue("org"), p.UserID)
	go s.render(exp.ID, docs)
	writeJSON(w, http.StatusAccepted, exp)
}

// render writes each document under its title and the first line of its body
// as a summary, then marks the export ready.
func (s *Server) render(exportID string, docs []Doc) {
	var b strings.Builder
	for _, d := range docs {
		summary := d.Body[:strings.IndexByte(d.Body, '\n')]
		fmt.Fprintf(&b, "# %s\n%s\n\n%s\n\n", d.Title, summary, d.Body)
	}
	s.store.FinishExport(exportID, b.String())
}

// downloadExport serves a finished export to the member who made it.
func (s *Server) downloadExport(w http.ResponseWriter, r *http.Request) {
	exp, err := s.store.Export(r.PathValue("id"))
	if err != nil || !exp.Ready {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="export.txt"`)
	_, _ = w.Write([]byte(exp.Data)) // headers are sent; a failed write is the client's disconnect
}

// redeemInvite adds the caller to the invite's organisation; each code works
// once and only before it expires.
func (s *Server) redeemInvite(w http.ResponseWriter, r *http.Request, p Principal) {
	inv, err := s.store.Invite(r.PathValue("code"))
	if err != nil || inv.Used || time.Now().After(inv.Expires) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	s.store.AddMember(inv.OrgID, p.UserID)
	s.store.MarkUsed(inv.Code)
	w.WriteHeader(http.StatusNoContent)
}

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
