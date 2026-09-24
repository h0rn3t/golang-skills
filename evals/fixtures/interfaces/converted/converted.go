package converted

// Store keeps items.
type Store interface {
	Get(id string) (string, error)
}

type memStore struct{}

func (*memStore) Get(string) (string, error) { return "", nil }

// NewStore returns the in-memory store.
func NewStore() Store { return &memStore{} }
