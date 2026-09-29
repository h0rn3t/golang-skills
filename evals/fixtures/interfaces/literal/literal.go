package literal

// Store is converted to only through a keyed struct literal field.
type Store interface{ Get(k string) string }

type memStore struct{ m map[string]string }

func (s *memStore) Get(k string) string { return s.m[k] }

// Server holds its store in an unexported field.
type Server struct{ store Store }

// NewServer wires the in-memory store.
func NewServer() *Server { return &Server{store: &memStore{}} }

// Lister is converted to only through a slice literal element.
type Lister interface{ List() []string }

type fixed struct{}

func (fixed) List() []string { return nil }

var all = []Lister{fixed{}}

// Namer is converted to only through a map literal value.
type Namer interface{ Name() string }

type named string

func (n named) Name() string { return string(n) }

var byKey = map[string]Namer{"a": named("a")}

// Closer is converted to only through a send statement.
type Closer interface{ Close() error }

type nopCloser struct{}

func (nopCloser) Close() error { return nil }

// CloseLater queues a no-op closer.
func CloseLater(ch chan<- Closer) { ch <- nopCloser{} }
