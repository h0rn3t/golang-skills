package ingest

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Record is one row of the remote record store.
type Record struct {
	ID    string
	Value string
}

// ErrInvalid reports a record the store rejects: one without an ID.
var ErrInvalid = errors.New("invalid record")

// MaxBatch is the most records one PutBatch call accepts.
const MaxBatch = 500

// Store is the client of the remote record store. Every call is one round
// trip to the service; this build of the client simulates the round trip and
// keeps the rows in memory. It is safe for concurrent use.
type Store struct {
	roundTrip time.Duration
	mu        sync.Mutex
	rows      map[string]string
	calls     int
}

// NewStore returns an empty store whose calls each take roundTrip.
func NewStore(roundTrip time.Duration) *Store {
	return &Store{roundTrip: roundTrip, rows: map[string]string{}}
}

// Put stores one record; a later Put of the same ID replaces the earlier one.
func (s *Store) Put(ctx context.Context, r Record) error {
	if err := s.wait(ctx); err != nil {
		return err
	}
	if r.ID == "" {
		return ErrInvalid
	}
	s.mu.Lock()
	s.rows[r.ID] = r.Value
	s.mu.Unlock()
	return nil
}

// PutBatch stores up to MaxBatch records in one round trip, in order, as
// successive Puts would: it stops at the first invalid record and returns
// how many it stored before it, with ErrInvalid.
func (s *Store) PutBatch(ctx context.Context, records []Record) (int, error) {
	if len(records) > MaxBatch {
		return 0, fmt.Errorf("batch of %d records exceeds %d", len(records), MaxBatch)
	}
	if err := s.wait(ctx); err != nil {
		return 0, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, r := range records {
		if r.ID == "" {
			return i, ErrInvalid
		}
		s.rows[r.ID] = r.Value
	}
	return len(records), nil
}

// Get returns the value stored for id.
func (s *Store) Get(id string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.rows[id]
	return v, ok
}

// Calls reports how many round trips the store has served.
func (s *Store) Calls() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls
}

func (s *Store) wait(ctx context.Context) error {
	s.mu.Lock()
	s.calls++
	s.mu.Unlock()
	t := time.NewTimer(s.roundTrip)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
