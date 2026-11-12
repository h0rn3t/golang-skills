// Package frontdesk serves a clinic's booking API: staff book appointment
// slots, mark the days the clinic is closed, and attach referral letters;
// patients get a reminder the day before their appointment.
package frontdesk

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

// maxReferral is the largest referral letter the API stores.
const maxReferral = 8 << 20

// Mailer sends a patient the reminder for an appointment.
type Mailer interface {
	Remind(ctx context.Context, patient string, slot time.Time) error
}

// Server serves the booking API and sends the reminders.
type Server struct {
	db      *sql.DB
	mail    Mailer
	zone    *time.Location // the clinic's; days are counted in it
	letters *os.Root       // referral letters, one file per appointment

	booking sync.Mutex    // held from the slot check to the insert
	uploads chan struct{} // bounds the letters this process receives at once

	mu     sync.Mutex
	closed map[string]bool // days the clinic is closed, as time.DateOnly
}

// NewServer returns a Server that keeps referral letters under dir, the
// -letters flag of the binary. It loads the closed days once; handleClose
// keeps them current from then on.
func NewServer(ctx context.Context, db *sql.DB, mail Mailer, zone *time.Location, dir string) (*Server, error) {
	rows, err := db.QueryContext(ctx, `SELECT day FROM closures`)
	if err != nil {
		return nil, fmt.Errorf("load closures: %w", err)
	}
	defer rows.Close()
	closed := map[string]bool{}
	for rows.Next() {
		var day time.Time
		if err := rows.Scan(&day); err != nil {
			return nil, fmt.Errorf("scan closure: %w", err)
		}
		closed[day.Format(time.DateOnly)] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("load closures: %w", err)
	}
	letters, err := os.OpenRoot(dir)
	if err != nil {
		return nil, fmt.Errorf("open letters directory: %w", err)
	}
	return &Server{db: db, mail: mail, zone: zone, letters: letters, uploads: make(chan struct{}, 8), closed: closed}, nil
}

// Close releases the letters directory.
func (s *Server) Close() error { return s.letters.Close() }

// Routes returns the booking API. The router in front of it has already
// authenticated the member of staff and checked that an appointment named in
// the path exists.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /appointments", s.handleBook)
	mux.HandleFunc("PUT /closures/{day}", s.handleClose)
	mux.HandleFunc("PUT /appointments/{id}/referral", s.handlePutReferral)
	mux.HandleFunc("GET /appointments/{id}/referral", s.handleGetReferral)
	return mux
}

// internalError logs err and answers 500.
func internalError(w http.ResponseWriter, r *http.Request, op string, err error) {
	slog.ErrorContext(r.Context(), "request failed", "op", op, "err", err)
	http.Error(w, "internal error", http.StatusInternalServerError)
}

// handleBook books a slot for a patient. A slot holds one appointment.
func (s *Server) handleBook(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<16)
	patient := r.PostFormValue("patient")
	slot, _ := time.Parse(time.RFC3339, r.PostFormValue("slot"))
	if patient == "" {
		http.Error(w, "patient is required", http.StatusBadRequest)
		return
	}
	if s.isClosed(slot) {
		http.Error(w, "the clinic is closed that day", http.StatusConflict)
		return
	}

	s.booking.Lock()
	defer s.booking.Unlock()
	var taken bool
	err := s.db.QueryRowContext(r.Context(), `SELECT EXISTS (SELECT 1 FROM appointments WHERE slot = $1)`, slot).Scan(&taken)
	if err != nil {
		internalError(w, r, "check slot", err)
		return
	}
	if taken {
		http.Error(w, "slot already booked", http.StatusConflict)
		return
	}
	if _, err := s.db.ExecContext(r.Context(), `INSERT INTO appointments (patient, slot) VALUES ($1, $2)`, patient, slot); err != nil {
		internalError(w, r, "book slot", err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// handleClose marks a day the clinic is closed; no slot can be booked on it.
// Appointments already booked on that day stay, and staff rebook them.
func (s *Server) handleClose(w http.ResponseWriter, r *http.Request) {
	day, err := time.ParseInLocation(time.DateOnly, r.PathValue("day"), s.zone)
	if err != nil {
		http.Error(w, "day must be YYYY-MM-DD", http.StatusBadRequest)
		return
	}
	date := day.Format(time.DateOnly)
	if _, err := s.db.ExecContext(r.Context(), `INSERT INTO closures (day) VALUES ($1) ON CONFLICT DO NOTHING`, date); err != nil {
		internalError(w, r, "close day", err)
		return
	}
	s.mu.Lock()
	s.closed[date] = true
	s.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

// isClosed reports whether the clinic is closed on the day of at.
func (s *Server) isClosed(at time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed[at.In(s.zone).Format(time.DateOnly)]
}

// handlePutReferral stores the referral letter of an appointment, replacing
// the one stored before.
func (s *Server) handlePutReferral(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 63)
	if err != nil {
		http.Error(w, "invalid appointment id", http.StatusBadRequest)
		return
	}
	select {
	case s.uploads <- struct{}{}:
		defer func() { <-s.uploads }()
	default:
		http.Error(w, "too many uploads", http.StatusServiceUnavailable)
		return
	}

	name := strconv.FormatUint(id, 10) + ".pdf"
	part := name + "." + rand.Text() + ".part" // a reader never sees a letter that is still arriving
	f, err := s.letters.OpenFile(part, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o640)
	if err != nil {
		internalError(w, r, "create letter", err)
		return
	}
	_, err = io.Copy(f, http.MaxBytesReader(w, r.Body, maxReferral))
	if err = errors.Join(err, f.Close()); err == nil {
		err = s.letters.Rename(part, name)
	}
	if err != nil {
		_ = s.letters.Remove(part) // the letter stored before, if any, stays
		if _, tooLarge := errors.AsType[*http.MaxBytesError](err); tooLarge {
			http.Error(w, "letter too large", http.StatusRequestEntityTooLarge)
			return
		}
		internalError(w, r, "store letter", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleGetReferral serves the referral letter of an appointment.
func (s *Server) handleGetReferral(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 63)
	if err != nil {
		http.Error(w, "invalid appointment id", http.StatusBadRequest)
		return
	}
	f, err := s.letters.Open(strconv.FormatUint(id, 10) + ".pdf")
	if errors.Is(err, fs.ErrNotExist) {
		http.Error(w, "no referral letter", http.StatusNotFound)
		return
	}
	if err != nil {
		internalError(w, r, "open letter", err)
		return
	}
	defer func() { _ = f.Close() }() // opened for reading: Close has no write to report
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `attachment; filename="referral.pdf"`)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(w, r, "", time.Time{}, f)
}

// Run sends the reminders for the next day's appointments once a day until
// ctx is done.
func (s *Server) Run(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		if err := s.remind(ctx, time.Now()); err != nil {
			slog.ErrorContext(ctx, "send reminders", "err", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// remind emails every patient whose appointment is on the day after now and
// who has not been reminded yet.
func (s *Server) remind(ctx context.Context, now time.Time) error {
	y, m, d := now.In(s.zone).Date()
	from := time.Date(y, m, d+1, 0, 0, 0, 0, s.zone)
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, patient, slot FROM appointments WHERE slot >= $1 AND slot < $2 AND reminded_at IS NULL`,
		from, from.AddDate(0, 0, 1))
	if err != nil {
		return fmt.Errorf("list appointments: %w", err)
	}
	defer rows.Close()
	type appointment struct {
		id      int64
		patient string
		slot    time.Time
	}
	var due []appointment
	for rows.Next() {
		var a appointment
		if err := rows.Scan(&a.id, &a.patient, &a.slot); err != nil {
			return fmt.Errorf("scan appointment: %w", err)
		}
		due = append(due, a)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("list appointments: %w", err)
	}
	for _, a := range due {
		if err := s.mail.Remind(ctx, a.patient, a.slot); err != nil {
			slog.ErrorContext(ctx, "send reminder", "appointment_id", a.id, "err", err)
			continue
		}
		if _, err := s.db.ExecContext(ctx, `UPDATE appointments SET reminded_at = now() WHERE id = $1`, a.id); err != nil {
			return fmt.Errorf("mark appointment %d reminded: %w", a.id, err)
		}
	}
	return nil
}
