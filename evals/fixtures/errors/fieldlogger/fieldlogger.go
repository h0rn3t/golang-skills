package fieldlogger

import (
	"fmt"
	"log/slog"
)

type Store struct{ logger *slog.Logger }

func (s *Store) Save(id int) error {
	if err := write(id); err != nil {
		s.logger.Error("save failed", "id", id, "err", err)
		return fmt.Errorf("save %d: %w", id, err)
	}
	return nil
}

func (s *Store) Load(id int) error {
	if err := write(id); err != nil {
		return fmt.Errorf("load %d: %w", id, err)
	}
	s.logger.Info("loaded", "id", id)
	return nil
}

func write(_ int) error { return nil }
