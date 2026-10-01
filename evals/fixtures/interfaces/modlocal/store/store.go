package store

import "golang-skills/evals/fixtures/interfaces/modlocal/model"

// Store is implemented here and returned by NewStore, the producer-owned case.
type Store interface {
	Get(id string) (model.Item, error)
}

type memStore struct{}

func (memStore) Get(string) (model.Item, error) { return model.Item{}, nil }

// NewStore returns the in-memory store.
func NewStore() Store { return memStore{} }

// Finder wants a model.Item. cache returns a model.Other, so nothing here
// implements Finder: an unresolved import must not make the two match.
type Finder interface {
	Find(id string) (model.Item, error)
}

type cache struct{}

func (cache) Find(string) (model.Other, error) { return model.Other{}, nil }
