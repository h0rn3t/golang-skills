package vault

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// maxUpload bounds one upload's body, quota aside.
const maxUpload = 256 << 20

// uploadsPerMinute is how many uploads one client address may start a minute.
const uploadsPerMinute = 30

// Event is one successful write, handed to the audit hook.
type Event struct {
	Tenant, Name, Action string
	At                   time.Time
}

// Server serves one Store over HTTP:
//
//	GET    /files/{tenant}          list the tenant's files
//	PUT    /files/{tenant}/{name}   store a file; POST does the same, then sends the browser to ?next=
//	GET    /files/{tenant}/{name}   fetch a file, with ETag and 304
//	DELETE /files/{tenant}/{name}   remove a file
//
// Every route needs the tenant's bearer token.
type Server struct {
	store  *Store
	tokens map[string]string // bearer token to tenant
	audit  func(ctx context.Context, ev Event) error
	log    *slog.Logger
	mux    *http.ServeMux

	mu   sync.Mutex
	hits map[string]*window // uploads started, by client address
}

// window counts events inside one minute.
type window struct {
	start time.Time
	n     int
}

// NewServer wires the routes. tokens maps each tenant's bearer token to the
// tenant; audit, when not nil, receives every successful write.
func NewServer(store *Store, tokens map[string]string, audit func(context.Context, Event) error, log *slog.Logger) *Server {
	s := &Server{store: store, tokens: tokens, audit: audit, log: log, mux: http.NewServeMux(), hits: make(map[string]*window)}
	s.mux.HandleFunc("GET /files/{tenant}", s.handleList)
	s.mux.HandleFunc("PUT /files/{tenant}/{name}", s.handlePut)
	s.mux.HandleFunc("POST /files/{tenant}/{name}", s.handlePut)
	s.mux.HandleFunc("GET /files/{tenant}/{name}", s.handleGet)
	s.mux.HandleFunc("DELETE /files/{tenant}/{name}", s.handleDelete)
	return s
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }

// ListenAndServe serves on addr until the listener fails or ctx ends, then
// gives in-flight requests five seconds to finish.
func (s *Server) ListenAndServe(ctx context.Context, addr string) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           s,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	errc := make(chan error, 1)
	go func() { errc <- srv.ListenAndServe() }()
	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
	}
	shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(shutdown)
}

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	tenant, ok := s.authorize(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"files": s.store.List(tenant), "used": s.store.Used(tenant)})
}

func (s *Server) handlePut(w http.ResponseWriter, r *http.Request) {
	tenant, ok := s.authorize(w, r)
	if !ok {
		return
	}
	if !s.allowUpload(clientIP(r)) {
		http.Error(w, "too many uploads", http.StatusTooManyRequests)
		return
	}
	labels := map[string]string{}
	for k, v := range r.Header {
		if label, ok := strings.CutPrefix(k, "X-Vault-Label-"); ok {
			labels[strings.ToLower(label)] = v[0]
		}
	}
	name := r.PathValue("name")
	m, err := s.store.Put(tenant, name, r.Header.Get("Content-Type"), labels, http.MaxBytesReader(w, r.Body, maxUpload))
	if err != nil {
		s.fail(w, err)
		return
	}
	s.record(r.Context(), Event{Tenant: tenant, Name: name, Action: "put", At: m.ModTime})
	if next := r.URL.Query().Get("next"); r.Method == http.MethodPost && next != "" {
		if !strings.HasPrefix(next, "/") {
			next = "/"
		}
		http.Redirect(w, r, next, http.StatusSeeOther)
		return
	}
	writeJSON(w, http.StatusCreated, m)
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	tenant, ok := s.authorize(w, r)
	if !ok {
		return
	}
	name := r.PathValue("name")
	f, m, err := s.store.Open(tenant, name)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer func() { _ = f.Close() }()
	etag := m.ETag()
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Header().Set("ETag", etag)
	w.Header().Set("Content-Type", m.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(m.Size, 10))
	w.WriteHeader(http.StatusOK)
	if _, err := io.Copy(w, f); err != nil {
		s.log.Debug("send interrupted", "tenant", tenant, "name", name, "err", err)
	}
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	tenant, ok := s.authorize(w, r)
	if !ok {
		return
	}
	name := r.PathValue("name")
	if err := s.store.Delete(tenant, name); err != nil {
		s.fail(w, err)
		return
	}
	s.record(r.Context(), Event{Tenant: tenant, Name: name, Action: "delete", At: time.Now()})
	w.WriteHeader(http.StatusNoContent)
}

// authorize returns the tenant of the path when the bearer token is that
// tenant's; otherwise it has answered 401.
func (s *Server) authorize(w http.ResponseWriter, r *http.Request) (string, bool) {
	tenant := r.PathValue("tenant")
	token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok || s.tokens[token] != tenant {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return "", false
	}
	return tenant, true
}

// record hands ev to the audit hook, if there is one. The entry is written
// even when the client has gone: the request's values stay, its
// cancellation does not.
func (s *Server) record(ctx context.Context, ev Event) {
	if s.audit == nil {
		return
	}
	if err := s.audit(context.WithoutCancel(ctx), ev); err != nil {
		s.log.Error("audit failed", "tenant", ev.Tenant, "name", ev.Name, "err", err)
	}
}

// fail answers err with the status it maps to; the client sees the message
// only for the errors it caused.
func (s *Server) fail(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrBadName):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, ErrExists):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, ErrQuota):
		http.Error(w, err.Error(), http.StatusInsufficientStorage)
	case errors.Is(err, fs.ErrNotExist):
		http.Error(w, "not found", http.StatusNotFound)
	default:
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			http.Error(w, "upload too large", http.StatusRequestEntityTooLarge)
			return
		}
		s.log.Error("request failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// clientIP returns the address of the client that started the request, as
// the proxy in front of the vault reports it.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// allowUpload counts an upload from addr against its per-minute allowance.
func (s *Server) allowUpload(addr string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	if len(s.hits) >= 4096 {
		for a, w := range s.hits {
			if now.Sub(w.start) >= time.Minute {
				delete(s.hits, a)
			}
		}
	}
	w := s.hits[addr]
	if w == nil || now.Sub(w.start) >= time.Minute {
		w = &window{start: now}
		s.hits[addr] = w
	}
	w.n++
	return w.n <= uploadsPerMinute
}

var bufPool = sync.Pool{New: func() any { return new(bytes.Buffer) }}

// encode renders v as JSON through a pooled buffer, so listing a large
// tenant does not allocate a fresh one per request.
func encode(v any) ([]byte, error) {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)
	if err := json.NewEncoder(buf).Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// writeJSON encodes v as the response body with status.
func writeJSON(w http.ResponseWriter, status int, v any) {
	body, err := encode(v)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body) // headers are sent; a failed write is the client's disconnect
}
