// Package digest names uploaded blobs by their content.
package digest

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// Name returns "sha256-" followed by the lowercase hex SHA-256 of data. The
// upload handler calls it once per blob before storing it.
func Name(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256-" + toHex(sum[:])
}

func toHex(b []byte) string {
	var sb strings.Builder
	for _, c := range b {
		fmt.Fprintf(&sb, "%02x", c)
	}
	return sb.String()
}
