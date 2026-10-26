// Package tiles renders signed map tiles on demand and keeps the rendered
// seeds in memory.
package tiles

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// maxTileBytes caps the canvas one request may ask for.
const maxTileBytes = 64 << 20

// Server renders tiles and serves them over HTTP.
type Server struct {
	mu    sync.RWMutex
	seeds map[string][]byte

	hits, misses atomic.Int64

	// counts holds the requests of each client in the current minute. A
	// crawl brings a few hundred thousand clients in one minute; a normal
	// minute, a few thousand.
	countMu sync.Mutex
	counts  map[string]int
	limit   int

	secret  []byte
	jobs    chan string
	quit    chan struct{}
	wg      sync.WaitGroup
	latency func(time.Duration)
}

// New returns a Server that signs tile keys with secret, allows each client
// limit requests per minute, and prerenders queued keys on workers
// goroutines. latency receives the duration of every request served.
func New(secret []byte, limit, workers int, latency func(time.Duration)) *Server {
	s := &Server{
		seeds:   make(map[string][]byte),
		counts:  make(map[string]int),
		limit:   limit,
		secret:  secret,
		jobs:    make(chan string),
		quit:    make(chan struct{}),
		latency: latency,
	}
	for range workers {
		s.wg.Go(s.prerender)
	}
	return s
}

func (s *Server) prerender() {
	for {
		select {
		case <-s.quit:
			return
		case key := <-s.jobs:
			s.store(key, render(key))
		}
	}
}

// Prerender queues key for a worker to render ahead of its first request.
func (s *Server) Prerender(key string) {
	select {
	case s.jobs <- key:
	case <-s.quit:
	}
}

// Stop ends the prerender workers. When it returns none of them is running,
// so the caller may close what they use.
func (s *Server) Stop() {
	close(s.quit)
}

func (s *Server) cached(key string) ([]byte, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.seeds[key]
	return b, ok
}

func (s *Server) store(key string, b []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seeds[key] = b
}

// Seed returns the rendered seed of key, rendering it on a miss.
func (s *Server) Seed(key string) []byte {
	s.mu.RLock()
	b, ok := s.cached(key)
	s.mu.RUnlock()
	if ok {
		s.hits.Add(1)
		return b
	}
	s.misses.Add(1)
	b = render(key)
	s.store(key, b)
	return b
}

// Healthy reports whether at least half of the seed lookups were hits. The
// readiness probe takes a pod out of rotation while it is not; a pod that has
// served nothing yet is healthy.
func (s *Server) Healthy() bool {
	hits, misses := float64(s.hits.Load()), float64(s.misses.Load())
	return hits/(hits+misses) >= 0.5
}

func (s *Server) allow(client string) bool {
	s.countMu.Lock()
	defer s.countMu.Unlock()
	s.counts[client]++
	return s.counts[client] <= s.limit
}

// ResetCounts starts a new minute for every client; a ticker in main calls
// it once a minute.
func (s *Server) ResetCounts() {
	s.countMu.Lock()
	defer s.countMu.Unlock()
	clear(s.counts)
}

// ServeHTTP serves GET /tile?key=K&sig=S&w=W&h=H: an RGBA canvas of W by H
// pixels filled from the seed of K, where S is the hex HMAC-SHA256 of K.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now().UTC()
	q := r.URL.Query()
	client := r.Header.Get("X-Client")
	if !s.allow(client) {
		http.Error(w, "slow down", http.StatusTooManyRequests)
		return
	}

	key := q.Get("key")
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(key))
	sig, err := hex.DecodeString(q.Get("sig"))
	if err != nil || !hmac.Equal(sig, mac.Sum(nil)) {
		http.Error(w, "bad signature", http.StatusForbidden)
		return
	}

	width, errW := strconv.Atoi(q.Get("w"))
	height, errH := strconv.Atoi(q.Get("h"))
	if errW != nil || errH != nil {
		http.Error(w, "bad size", http.StatusBadRequest)
		return
	}
	size := width * height * 4
	if size > maxTileBytes {
		http.Error(w, "tile too large", http.StatusRequestEntityTooLarge)
		return
	}

	canvas := make([]byte, size)
	seed := s.Seed(key)
	for i := range canvas {
		canvas[i] = seed[i%len(seed)]
	}

	log.Printf("tile %s for client %s: %d bytes", key, client, size)
	w.Header().Set("Content-Type", "application/octet-stream")
	_, _ = w.Write(canvas) // the client is gone; nothing to tell it
	s.latency(time.Now().UTC().Sub(start))
}

// SaveSecret writes the signing secret to path, readable by the service user
// only. Installs from releases before 2.0 left that file at mode 0644.
func SaveSecret(path string, secret []byte) error {
	return os.WriteFile(path, secret, 0o600)
}

func render(key string) []byte {
	sum := sha256.Sum256([]byte(key))
	return sum[:]
}
