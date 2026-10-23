package constraint

import (
	"io"
	"strconv"
)

// Reader is another name for io.Reader, declared in io, not here.
type Reader = io.Reader

// Number may only be a type constraint: it has a type set.
type Number interface {
	~int | ~float64
	String() string
}

// Keyed is a basic interface Index takes as its constraint.
type Keyed interface {
	Key() string
}

// ID satisfies Number.
type ID int

func (id ID) String() string { return strconv.Itoa(int(id)) }

// User satisfies Keyed.
type User struct{ Name string }

// Key returns the user's name.
func (u User) Key() string { return u.Name }

// File satisfies io.Reader.
type File struct{}

func (File) Read([]byte) (int, error) { return 0, io.EOF }

// Index maps items by key.
func Index[T Keyed](items []T) map[string]T {
	m := make(map[string]T, len(items))
	for _, it := range items {
		m[it.Key()] = it
	}
	return m
}

// Sum adds xs.
func Sum[T Number](xs ...T) T {
	var total T
	for _, x := range xs {
		total += x
	}
	return total
}

// Users indexes one user; the instantiation makes the compiler check User
// against Keyed.
func Users() map[string]User { return Index([]User{{Name: "a"}}) }
