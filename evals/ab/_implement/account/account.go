// Package account serves the save button of the account settings page.
package account

import (
	"context"
	"errors"
	"net/http"
)

// Account is the record as the store keeps it. The store writes records as
// JSON, which is why the fields carry tags.
type Account struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	DisplayName  string `json:"display_name"`
	Phone        string `json:"phone,omitempty"`
	PasswordHash string `json:"password_hash"`
	Role         string `json:"role"`
}

// ErrNotFound is returned by a Store for an unknown account ID.
var ErrNotFound = errors.New("account not found")

// Store keeps accounts. It is safe for concurrent use.
type Store interface {
	Get(ctx context.Context, id string) (Account, error)
	Put(ctx context.Context, a Account) error
}

// NewHandler returns the handler that serves PATCH /accounts/{id}.
//
// The body is one JSON object naming only the settings the user changed: any
// of "email", "display_name", and "phone". A setting the body does not name
// keeps its stored value. "phone": null removes the phone number. "email" and
// "display_name" are required, so an empty or null value for either is a 400.
// Any other member, a body that is not exactly one JSON object, or a body over
// 64 KiB is a 400, and a 400 changes nothing.
//
// An unknown account is a 404. On success the handler saves the account and
// answers 200 with the account as the settings page shows it: a JSON object
// with exactly the members "id", "email", "display_name", and "phone" (""
// when there is none). A store failure is a 500 whose body does not carry the
// error.
func NewHandler(store Store) http.Handler {
	panic("not implemented")
}
