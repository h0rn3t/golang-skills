// Package purge removes tenant workspaces from disk when a tenant deletes
// one.
package purge

import "errors"

// ErrInvalidID is the error Delete wraps for an id that names no workspace.
var ErrInvalidID = errors.New("invalid workspace id")

// Workspaces is the directory that holds one subdirectory per workspace.
// Each subdirectory is named by its workspace's ID, a UUID in the canonical
// lowercase form, and holds that workspace's files. All tenants' workspaces
// share the directory.
type Workspaces struct {
	// The unexported fields are the implementation's own.
}

// Open returns the Workspaces held in dir.
func Open(dir string) (*Workspaces, error) {
	panic("not implemented")
}

// Close releases the directory.
func (ws *Workspaces) Close() error {
	panic("not implemented")
}

// Delete removes workspace id and everything in it.
//
// id is the value the client sent in DELETE /workspaces/{id}, after the
// router has checked that the workspace belongs to the caller's tenant. An id
// that is not a workspace ID returns an error wrapping ErrInvalidID and
// removes nothing. Deleting a workspace that does not exist is not an error.
func (ws *Workspaces) Delete(id string) error {
	panic("not implemented")
}
