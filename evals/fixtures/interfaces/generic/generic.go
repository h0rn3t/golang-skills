package generic

// Getter is a generic interface implemented here through Getter[int] and
// returned by NewGetter, the producer-owned case.
type Getter[T any] interface{ Get() T }

type intGetter struct{}

func (intGetter) Get() int { return 1 }

// NewGetter returns the int getter.
func NewGetter() Getter[int] { return intGetter{} }
