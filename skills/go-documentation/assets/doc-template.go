// Package example appends text records to a log file.
//
// # Getting Started
//
// Open a [Log], append to it, and close it when done; Close reports a write
// that failed late, so its error is returned too:
//
//	l, err := example.Open("app.log")
//	if err != nil {
//		return err
//	}
//	if err := l.Append("started"); err != nil {
//		return errors.Join(err, l.Close())
//	}
//	return l.Close()
package example

import (
	"errors"
	"fmt"
	"os"
	"sync"
)

// ErrEmpty is returned by [Log.Append] for an empty record.
var ErrEmpty = errors.New("example: empty record")

// A Log appends records to a file, one per line.
//
// A Log is safe for concurrent use. Call [Log.Close] to release the file.
type Log struct {
	mu sync.Mutex
	f  *os.File
}

// Open opens the log at path for appending, creating the file if it does not
// exist.
func Open(path string) (*Log, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, err
	}
	return &Log{f: f}, nil
}

// Append writes record followed by a newline.
func (l *Log) Append(record string) error {
	if record == "" {
		return ErrEmpty
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	_, err := fmt.Fprintln(l.f, record)
	return err
}

// Close closes the underlying file.
func (l *Log) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.f.Close()
}
