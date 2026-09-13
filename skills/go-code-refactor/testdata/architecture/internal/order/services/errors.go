package services

import "errors"

// ErrForbidden reports an actor the use case does not allow. Authorization is
// the service's decision; a handler only carries the identity.
var ErrForbidden = errors.New("order: forbidden")
