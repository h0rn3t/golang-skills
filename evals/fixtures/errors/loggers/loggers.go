package loggers

import (
	"fmt"
	"log"
)

// sink stands in for logrus, zap's sugared logger, and a catalog that
// happens to have an Error method.
type sink interface {
	Error(msg string, args ...any)
	Errorw(msg string, kv ...any)
}

var logrus, zap sink

type opError struct{ Err error }

func (e *opError) Error() string { return e.Err.Error() }

type Service struct {
	sugar    sink
	auditLog sink
	catalog  sink
}

func (s *Service) Package(err error) error {
	if err != nil {
		logrus.Error("package", err) // finding: logrus
		return err
	}
	return nil
}

func (s *Service) Sugar(err error) error {
	if err != nil {
		s.sugar.Errorw("sugar", "err", err) // finding: sugar
		return err
	}
	return nil
}

func (s *Service) Rewrap(err error) error {
	if err != nil {
		log.Printf("rewrap: %v", err) // finding: rewrap
		err = fmt.Errorf("rewrap: %w", err)
		return err
	}
	return nil
}

func (s *Service) Typed(err error) error {
	if err != nil {
		zap.Error("typed", err) // finding: typed
		return &opError{Err: err}
	}
	return nil
}

func (s *Service) Audit(err error) error {
	if err != nil {
		s.auditLog.Error("audit", err) // finding: audit
		return err
	}
	return nil
}

func (s *Service) Lookup(err error) error {
	if err != nil {
		s.catalog.Error("lookup", err) // a catalog, not a logger
		return err
	}
	return nil
}

func HasText(err error) bool {
	return err.Error() != "" // empty text, not matching
}
