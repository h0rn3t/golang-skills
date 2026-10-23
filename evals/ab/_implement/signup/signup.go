// Package signup creates user accounts for the registration form.
package signup

import (
	"context"
	"errors"
)

// User is a registered account.
type User struct {
	ID       int64
	Email    string
	Username string
}

// Errors Register returns for a value another account already holds.
var (
	ErrEmailTaken    = errors.New("signup: email already registered")
	ErrUsernameTaken = errors.New("signup: username already taken")
)

// ErrNoRows is returned by Users.FindByEmail and Users.FindByUsername when no
// account matches.
var ErrNoRows = errors.New("users: no rows")

// PgError is the error the database driver returns for a failed statement.
type PgError struct {
	// Code is the SQLSTATE: "23505" is a unique violation.
	Code string
	// ConstraintName names the violated constraint, if any.
	ConstraintName string
	Message        string
}

func (e *PgError) Error() string { return "ERROR: " + e.Message + " (SQLSTATE " + e.Code + ")" }

// Users is the users table. Its schema:
//
//	CREATE TABLE users (
//	    id       bigserial PRIMARY KEY,
//	    email    text NOT NULL CONSTRAINT users_email_key UNIQUE,
//	    username text NOT NULL CONSTRAINT users_username_key UNIQUE
//	);
//
// Every method runs one statement on a connection pool, so two calls may run
// on different connections at the same time.
type Users interface {
	// FindByEmail returns the account with this email, or ErrNoRows.
	FindByEmail(ctx context.Context, email string) (User, error)
	// FindByUsername returns the account with this username, or ErrNoRows.
	FindByUsername(ctx context.Context, username string) (User, error)
	// Insert adds an account and returns its id. A statement that fails
	// returns the driver's error, a *PgError for a constraint violation.
	Insert(ctx context.Context, email, username string) (int64, error)
}

// Register creates the account for a submitted registration form.
//
// An email another account already has is ErrEmailTaken, and a username
// another account already has is ErrUsernameTaken, matched with errors.Is.
// Any other database failure is returned as an error that is neither of the
// two. On success it returns the new account. Register serves every form
// submission, many at once.
func Register(ctx context.Context, users Users, email, username string) (User, error) {
	panic("not implemented")
}
