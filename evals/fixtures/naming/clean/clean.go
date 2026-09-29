package cache

const MaxRetries = 3

type Store struct {
	name string
}

func (s *Store) Name() string {
	return s.name
}

// A Get method that takes parameters is a lookup, not a simple accessor.
func (s *Store) GetEntry(key string) string {
	return key
}
