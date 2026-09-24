package probe

import "io"

// AtEnd reports whether err ends a stream.
func AtEnd(err error) bool {
	return err == io.EOF
}
